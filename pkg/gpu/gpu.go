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

package gpu

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var query = []string{
	"index",
	"uuid",
	"name",
	"timestamp",
	"memory.total",
	"memory.free",
	"memory.used",
	"utilization.gpu",
	"utilization.memory",
}

type GpuInfo struct {
	Index       int    `json:"index"`
	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	Timestamp   string `json:"timestamp"`
	TotalMemory int64  `json:"memory.total"`
	MemoryFree  int64  `json:"memory.free"`
	MemoryUsed  int64  `json:"memory.used"`
	GpuUsage    int    `json:"utilization.gpu"`
	MemoryUsage int    `json:"utilization.memory"`
}

func GetGpuInfo() ([]GpuInfo, error) {
	cmd := exec.Command("nvidia-smi", fmt.Sprintf("--query-gpu=%s", strings.Join(query, ",")), "--format=csv,noheader,nounits")

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		return nil, err
	}
	if errbuf.Len() != 0 {
		return nil, fmt.Errorf(errbuf.String())
	}

	resultStr := strings.TrimSuffix(outbuf.String(), "\n")
	result := strings.Split(resultStr, "\n")

	var infos []GpuInfo

	for _, r := range result {
		var info GpuInfo
		rawInfo := strings.Split(r, ", ")

		info.Index, _ = strconv.Atoi(rawInfo[0])
		info.Uuid = rawInfo[1]
		info.Name = rawInfo[2]
		info.Timestamp = rawInfo[3]
		info.TotalMemory, _ = strconv.ParseInt(rawInfo[4], 10, 64)
		info.MemoryFree, _ = strconv.ParseInt(rawInfo[5], 10, 64)
		info.MemoryUsed, _ = strconv.ParseInt(rawInfo[6], 10, 64)
		info.GpuUsage, _ = strconv.Atoi(rawInfo[7])
		info.MemoryUsage, _ = strconv.Atoi(rawInfo[8])

		infos = append(infos, info)
	}

	return infos, nil
}

func CheckGpuId(id []int) bool {
	infos, err := GetGpuInfo()
	if err != nil {
		return false
	}

	for _, i := range id {
		for _, info := range infos {
			if info.Index == i {
				continue
			}
		}
		return false
	}

	return true
}
