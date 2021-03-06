//
// Copyright © 2017-2020 Solus Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package jobs

import (
	"errors"
	"fmt"
	log "github.com/DataDrake/waterlog"
	"github.com/getsolus/ferryd/config"
	"github.com/jmoiron/sqlx"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrNoJobReady is returned when there are no available jobs or the next job is blocked by a running job
	ErrNoJobReady = errors.New("No jobs ready to run")
)

const (
	// DB is the filename of the jobs database
	DB = "jobs.db"
	// SQLiteOpts is a list of options for the go-sqlite3 driver
	SQLiteOpts = "?cache=shared"
)

// Store handles the storage and manipulation of incomplete jobs
type Store struct {
	sync.Mutex

	db   *sqlx.DB
	next *Job
	stop chan bool
	done chan bool
}

// NewStore creates a fully initialized Store and sets up Bolt Buckets as needed
func NewStore() (*Store, error) {
	// Open the database if we can
	db, err := sqlx.Open("sqlite3", filepath.Join(config.Current.BaseDir, DB)+SQLiteOpts)
	if err != nil {
		return nil, err
	}
	// See: https://github.com/mattn/go-sqlite3/issues/209
	db.SetMaxOpenConns(1)

	// Create "jobs" table if missing
	db.MustExec(JobSchema)

	s := &Store{
		db:   db,
		next: nil,
		stop: make(chan bool),
		done: make(chan bool),
	}
	// reset running jobs and return
	return s, s.UnclaimRunning()
}

// Close will clean up our private job database
func (s *Store) Close() error {
	if s.db != nil {
		err := s.db.Close()
		s.db = nil
		return err
	}
	return nil
}

// GetJob retrieves a Job from the DB
func (s *Store) GetJob(id int) (*Job, error) {
	var j Job
	err := s.db.Get(&j, getJob, id)
	return &j, err
}

// UnclaimRunning will find all claimed jobs and unclaim them again
func (s *Store) UnclaimRunning() error {
	var err error
	s.Lock()
	if _, err = s.db.Exec(clearRunningJobs); err != nil {
		err = fmt.Errorf("Failed to unclaim running jobs, reason: '%s'", err.Error())
	}
	s.Unlock()
	return err
}

// Push inserts a new Job into the queue
func (s *Store) Push(j *Job) (id int, err error) {
	s.Lock()
	// Set Job parameters
	j.Status = New
	j.Created.Time = time.Now().UTC()
	j.Created.Valid = true
	// Start a DB transaction
	tx, err := s.db.Beginx()
	if err != nil {
		goto UNLOCK
	}
	// Insert the New Job
	if err = j.Create(tx); err != nil {
		tx.Rollback()
		goto UNLOCK
	}
	id = j.ID
	// Complete the transaction
	err = tx.Commit()
UNLOCK:
	s.Unlock()
	return id, err
}

func (s *Store) findNewJob() {
	// get the currently runnign jobs
	var active []Job
	if err := s.db.Select(&active, runningJobs); err != nil {
		return
	}
	// Check for serial jobs that are blocking
	for _, j := range active {
		if !IsParallel[j.Type] {
			return
		}
	}
	// Otherwise, get the next available job
	var next Job
	if err := s.db.Get(&next, nextJob); err != nil {
		return
	}
	// Check if we are blocked by parallel jobs
	if !IsParallel[next.Type] && len(active) > 0 {
		return
	}
	s.next = &next
}

// Claim gets the first available job, if one exists and is not blocked by running jobs
func (s *Store) Claim() (j *Job, err error) {
	var tx *sqlx.Tx
	s.Lock()
	if s.next == nil {
		err = ErrNoJobReady
		goto UNLOCK
	}
	// claim the next job
	s.next.Status = Running
	s.next.Started.Time = time.Now().UTC()
	s.next.Started.Valid = true
	// Start a DB transaction
	tx, err = s.db.Beginx()
	if err != nil {
		goto UNLOCK
	}
	// Save the status change
	if err = s.next.Save(tx); err != nil {
		tx.Rollback()
		goto UNLOCK
	}
	// Finish the transaction
	if err = tx.Commit(); err != nil {
		goto UNLOCK
	}
	// find the next replacement job
	j, s.next = s.next, nil
UNLOCK:
	s.findNewJob()
	s.Unlock()
	return
}

// Retire marks a job as completed and updates the DB record
func (s *Store) Retire(j *Job) error {
	s.Lock()
	// Start a DB transaction
	tx, err := s.db.Beginx()
	if err != nil {
		goto UNLOCK
	}
	// Mark as finished
	j.Finished.Time = time.Now().UTC()
	j.Finished.Valid = true
	if err = j.Save(tx); err != nil {
		tx.Rollback()
		goto UNLOCK
	}
	// Finish the transaction
	err = tx.Commit()
UNLOCK:
	s.Unlock()
	return err
}

// Active will attempt to return a list of active jobs within
// the scheduler suitable for consumption by the CLI client
func (s *Store) Active() (List, error) {
	var list List
	var list2 List
	// Get all new jobs
	if err := s.db.Select(&list, newJobs); err != nil {
		log.Errorf("Failed to read new jobs, reason: '%s'", err.Error())
		return nil, err
	}
	// Get All running jobs
	if err := s.db.Select(&list2, runningJobs); err != nil {
		log.Errorf("Failed to read active jobs, reason: '%s'", err.Error())
		return nil, err
	}
	// Append them together
	list = append(list, list2...)
	return list, nil
}

// Completed will return all successfully completed jobs still stored
func (s *Store) Completed() (List, error) {
	var list List
	var err error
	if err = s.db.Select(&list, completedJobs); err != nil {
		err = fmt.Errorf("Failed to read completed jobs, reason: '%s'", err.Error())
	}
	return list, err
}

// Failed will return all failed jobs that are still stored
func (s *Store) Failed() (List, error) {
	var list List
	var err error
	if err = s.db.Select(&list, failedJobs); err != nil {
		err = fmt.Errorf("Failed to read failed jobs, reason: '%s'", err.Error())
	}
	return list, err
}

// ResetCompleted will remove all completion records from our store and reset the pointer
func (s *Store) ResetCompleted() error {
	s.Lock()
	var err error
	if _, err = s.db.Exec(clearCompletedJobs); err != nil {
		err = fmt.Errorf("Failed to clear completed jobs, reason: '%s'", err.Error())
	}
	s.Unlock()
	return err
}

// ResetFailed will remove all fail records from our store and reset the pointer
func (s *Store) ResetFailed() error {
	s.Lock()
	var err error
	if _, err = s.db.Exec(clearFailedJobs); err != nil {
		err = fmt.Errorf("Failed to clear failed jobs, reason: '%s'", err.Error())
	}
	s.Unlock()
	return err
}

// ResetQueued will remove all unexecuted records from our store and reset the pointer
func (s *Store) ResetQueued() error {
	s.Lock()
	var err error
	if _, err = s.db.Exec(clearQueuedJobs); err != nil {
		err = fmt.Errorf("Failed to clear queued jobs, reason: '%s'", err.Error())
	}
	s.Unlock()
	return err
}
