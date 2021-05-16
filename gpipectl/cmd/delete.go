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
	"net/http"
	"strconv"

	"github.com/Shikugawa/gpupipe/pkg/types"
	"github.com/spf13/cobra"
)

var (
	id        string
	deleteAll bool

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "delete pending GPU process",
		Run: func(cmd *cobra.Command, args []string) {
			var request types.ProcessDeleteRequest
			request.Id = id

			requestRaw, _ := json.Marshal(request)
			resp, err := http.Post("http://"+host+":"+strconv.Itoa(int(port))+"/delete", "application/json", bytes.NewBuffer(requestRaw))
			if err != nil {
				fmt.Println(err)
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				fmt.Println("error")
				return
			}

			fmt.Println("succeess")
		},
	}
)

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().Int16VarP(&port, "port", "p", 8000, "server port")
	deleteCmd.Flags().StringVar(&host, "host", "0.0.0.0", "server host")
	deleteCmd.Flags().StringVar(&id, "id", "", "process id")
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "delete all process or not")
}
