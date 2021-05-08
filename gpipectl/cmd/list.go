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
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	host string
	port int16

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "get pending resouces in scheduler",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := http.Get("http://" + host + ":" + strconv.Itoa(int(port)) + "/list")
			if err != nil {
				fmt.Println(err)
				return
			}

			defer resp.Body.Close()

			b, _ := ioutil.ReadAll(resp.Body)
			var fixed bytes.Buffer
			if err := json.Indent(&fixed, b, "", "\t"); err != nil {
				fmt.Println(err)
				return
			}

			fixed.WriteTo(os.Stdout)
			fmt.Println()
		},
	}
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().Int16VarP(&port, "port", "p", 8000, "server port")
	listCmd.Flags().StringVar(&host, "host", "0.0.0.0", "server host")
}
