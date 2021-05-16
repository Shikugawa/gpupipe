// Copyright 2021 Rei Shimizu

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Shikugawa/gpupipe/pkg/types"
	"github.com/spf13/cobra"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "generate single line command from task definition",
		Run: func(cmd *cobra.Command, args []string) {
			target, err := ioutil.ReadFile(targetJson)
			if err != nil {
				fmt.Println(err)
				return
			}

			var request types.ProcessPublishRequest
			if err := json.NewDecoder(bytes.NewBuffer(target)).Decode(&request); err != nil {
				fmt.Println(err)
				return
			}

			var targetCommand string

			for _, unit := range request.Command {
				targetCommand += unit + " "
			}

			fmt.Println(targetCommand)
		},
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&targetJson, "target", "", "target process")
	generateCmd.MarkFlagRequired("target")
}
