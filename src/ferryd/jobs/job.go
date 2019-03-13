//
// Copyright © 2017-2019 Solus Project
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
	"time"
)

const JobSchema = `
CREATE TABLE IF NOT EXISTS jobs (
    id         INTEGER PRIMARY KEY,
    type       INTEGER,
    src_repo   STRING,
    dst_repo   STRING,
    sources    TEXT,
    release    INTEGER,
    max_keep   INTEGER,
    mode       INTEGER,
    created    DATETIME,
    started    DATETIME,
    finished   DATETIME,
    status     INTEGER,
    message    TEXT
)
`

// Job Status
const (
	New       = 0
	Running   = 1
	Failed    = 2
	Cancelled = 3
	Completed = 4
)

// Job is an entry in the Job Table
type Job struct {
	ID   int
	Type JobType
	// Job-specific arguments
	SrcRepo    string `db:src_repo`
	DstRepo    string `db:dst_repo`
	SourcesRaw string `db:sources`
	Sources    []string
	Release    int
	MaxKeep    int `db:max_keep`
	Mode       int
	// Job tracking
	Created  time.Time
	Started  time.Time
	Finished time.Time
	Status   int
	Message  string
}

// Queries for retrieving Jobs of a particular status
const (
	newJobs       = "SELECT * FROM jobs WHERE status=0"
	runningJobs   = "SELECT * FROM jobs WHERE status=1"
	failedJobs    = "SELECT * FROM jobs WHERE status=2"
	cancelledJobs = "SELECT * FROM jobs WHERE status=3"
	completedJobs = "SELECT * FROM jobs WHERE status=4"
)

// Query for creating a new Job
const insertJob = `
INSERT INTO jobs (
    id, type,
    src_repo, dst_repo, sources, release, max_keep, mode,
    created, started, finished, status, message
) VALUES (
    NULL, :type,
    :src_repo, :dst_repo, :sources, :release, :max_keep, :mode,
    :created, NULL, NULL, :status, NULL
)
`

// Queries for updating the status of a job
const (
	markRunning   = "UPDATE jobs SET status=:status, started=:started WHERE id=:id"
	markFailed    = "UPDATE jobs SET status=:status, finished=:finished, message=:message WHERE id=:id"
	markCancelled = "UPDATE jobs SET status=:status, finished=:finished, message=:message WHERE id=:id"
	markCompleted = "UPDATE jobs SET status=:status, finished=:finished, message=:message WHERE id=:id"
)

// Queries for Cleaning up the Job queue
const (
	clearFailedJobs    = "DELETE FROM jobs WHERE status=2"
	clearCancelledJobs = "DELETE FROM jobs WHERE status=3"
	clearCompletedJobs = "DELETE FROM jobs WHERE status=4"
)