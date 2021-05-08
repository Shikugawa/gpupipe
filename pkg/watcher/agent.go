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

package watcher

import (
	"fmt"
	"time"

	"github.com/Shikugawa/gpupipe/pkg/gpu"
)

type Agent struct {
	gpuInfoRequestInterval  time.Duration
	memoryUsageLowWatermark int
}

func NewAgent(requestInterval, memoryUsageLowWatermark int) *Agent {
	if memoryUsageLowWatermark > 100 {
		memoryUsageLowWatermark = 100
	}

	return &Agent{
		gpuInfoRequestInterval:  time.Duration(requestInterval) * time.Second,
		memoryUsageLowWatermark: memoryUsageLowWatermark,
	}
}

func (w *Agent) Run(ch chan<- []int) {
	for {
		infos, err := gpu.GetGpuInfo()
		if err != nil {
			fmt.Println(err)
			continue
		}

		var target []int
		for _, info := range infos {
			if info.MemoryUsage <= w.memoryUsageLowWatermark {
				target = append(target, info.Index)
			}
		}
		ch <- target
		time.Sleep(w.gpuInfoRequestInterval)
	}
}
