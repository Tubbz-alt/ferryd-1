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
	"os"
)

// Status fulfills the "status" sub-command
var Status = &cmd.CMD{
	Name:  "status",
	Alias: "hi",
	Short: "Get the status of the currently running ferryd",
	Args:  &StatusArgs{},
	Run:   StatusRun,
}

// StatusArgs are the arguments to the "status" sub-command
type StatusArgs struct{}

// StatusRun executes the "status" sub-command
func StatusRun(r *cmd.RootCMD, c *cmd.CMD) {
	// Convert our flags
	flags := r.Flags.(*GlobalFlags)
	//args  := c.Args.(*StatusArgs)
	// Create a Client
	client := v1.NewClient(flags.Socket)
	defer client.Close()
	// Request a status update
	status, err := client.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while getting status: %v\n", err)
		os.Exit(1)
	}
	// Print out the status
	status.Print(os.Stdout)
}
