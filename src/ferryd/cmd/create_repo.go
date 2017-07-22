//
// Copyright © 2017 Ikey Doherty <ikey@solus-project.com>
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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"libferry"
	"os"
)

var createRepoCmd = &cobra.Command{
	Use:   "create-repo",
	Short: "create a new repository",
	Long:  "Create a new repository, if it doesn't exist",
	Run:   createRepo,
}

func init() {
	RootCmd.AddCommand(createRepoCmd)
}

func createRepo(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "create-repo takes exactly 1 argument")
		return
	}

	repoName := args[0]

	repoDir := "./ferry"
	if err := os.MkdirAll(repoDir, 00755); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create required directory \"%s\": %s", repoDir, err)
		return
	}

	// TODO: Get the right cwd always ..
	manager, err := libferry.NewManager(repoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	defer manager.Close()

	if err := manager.CreateRepo(repoName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("Created repository: %s\n", repoName)
}
