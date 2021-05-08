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

package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/Shikugawa/gpupipe/pkg/watcher"
)

type GreedyScheduler struct {
	queue               []*Process
	watcher             *watcher.Agent
	targetGpuId         chan []int
	maxPendingQueueSize int
}

func (s *GreedyScheduler) Publish(rootpath string, command []string, gpuId []int) error {
	if len(s.queue) >= s.maxPendingQueueSize {
		return fmt.Errorf("failed to publish pending process with queue size overflow")
	}
	s.queue = append(s.queue, NewProcess(rootpath, command, gpuId))
	return nil
}

func (s *GreedyScheduler) List() ([]byte, error) {
	processSet := make(map[string][]Process, 0)
	processSet["processes"] = make([]Process, 0)

	for _, queuedProcess := range s.queue {
		processSet["processes"] = append(processSet["processes"], *queuedProcess)
	}

	b, err := json.Marshal(processSet)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *GreedyScheduler) Run() {
	for {
		select {
		case availableGpuIds := <-s.targetGpuId:
			for _, queuedProcess := range s.queue {
				if queuedProcess.ProcessState != Pending {
					continue
				}

				canSpawn := true
				for _, requestGpuId := range queuedProcess.GpuId {
					requestGpuIdAvailable := false
					for _, availableGpuId := range availableGpuIds {
						if availableGpuId == requestGpuId {
							requestGpuIdAvailable = true
							break
						}
					}

					if !requestGpuIdAvailable {
						log.Printf("requested GPU ID %d has not be available", requestGpuId)
						canSpawn = false
						break
					}
				}

				if !canSpawn {
					log.Printf("process can't be executed")
					continue
				}
				queuedProcess.ProcessState = CanSpawn
			}

			var canSpawnProcessIdx []int
			for i, queuedProcess := range s.queue {
				if queuedProcess.ProcessState == CanSpawn {
					canSpawnProcessIdx = append(canSpawnProcessIdx, i)
				}
			}

			if len(canSpawnProcessIdx) == 0 {
				log.Printf("no ready process")
				continue
			}

			sort.Slice(canSpawnProcessIdx, func(i, j int) bool {
				return s.queue[i].IssuedTime.Before(s.queue[j].IssuedTime)
			})

			shouldSpawnProcessIdx := canSpawnProcessIdx[0]
			s.queue[shouldSpawnProcessIdx].ProcessState = Active

			if err := s.queue[shouldSpawnProcessIdx].Spawn(); err != nil {
				s.queue[shouldSpawnProcessIdx].ProcessState = CanSpawn
				log.Println("failed to spawn", err)
			}

			for _, queuedProcess := range s.queue {
				if queuedProcess.ProcessState == CanSpawn {
					queuedProcess.ProcessState = Pending
				}
			}
		}
	}
}

func NewGreedyScheduler(maxPendingQueueSize, gpuInfoRequestInterval, memoryUsageLowWatermark int) *GreedyScheduler {
	targetGpuId := make(chan []int)
	watcher := watcher.NewAgent(gpuInfoRequestInterval, memoryUsageLowWatermark)
	go watcher.Run(targetGpuId)

	return &GreedyScheduler{
		queue:               make([]*Process, 0),
		watcher:             watcher,
		targetGpuId:         targetGpuId,
		maxPendingQueueSize: maxPendingQueueSize,
	}
}
