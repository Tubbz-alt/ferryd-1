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

package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/cmd"
	"github.com/getsolus/ferryd/api/v1"
	"github.com/getsolus/ferryd/repo"
	"os"
)

// List fulfills the "list-repo" sub-command
var List = &cmd.CMD{
	Name:  "list-repo",
	Alias: "lr",
	Short: "List all tracked repos",
	Args:  &ListArgs{},
	Run:   ListRun,
}

// ListArgs are the arguments to the "list-repo" sub-command
type ListArgs struct{}

// ListRun executes the "list-repo" sub-command
func ListRun(r *cmd.RootCMD, c *cmd.CMD) {
	// Convert our flags
	flags := r.Flags.(*GlobalFlags)
	//args  := c.Args.(*ListArgs)
	// Create a Client
	client := v1.NewClient(flags.Socket)
	defer client.Close()
	// Request a summary
	var f repo.FullSummary
	var err error
	if f, err = client.Repos(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while listing repos: %v\n", err)
		os.Exit(1)
	}
	// Print the summary
	f.Print(os.Stdout)
}
