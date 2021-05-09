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
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"github.com/Shikugawa/gpupipe/pkg/process"
	"github.com/Shikugawa/gpupipe/pkg/types"
	"github.com/Shikugawa/gpupipe/pkg/watcher"
)

type GreedyScheduler struct {
	queue               *list.List
	watcher             *watcher.Agent
	targetGpuId         chan []int
	maxPendingQueueSize int
	ProcessEventHandler *ProcessEventHandler
}

func (s *GreedyScheduler) Publish(r *types.ProcessPublishRequest) error {
	if s.queue.Len() >= s.maxPendingQueueSize {
		return fmt.Errorf("failed to publish pending process with queue size overflow")
	}
	s.queue.PushBack(process.NewProcess(r))
	return nil
}

func (s *GreedyScheduler) List() ([]byte, error) {
	processSet := make(map[string][]process.Process, 0)
	processSet["processes"] = make([]process.Process, 0)

	for e := s.queue.Front(); e != nil; e = e.Next() {
		queuedProcess := e.Value.(*process.Process)
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
			for e := s.queue.Front(); e != nil; e = e.Next() {
				queuedProcess := e.Value.(*process.Process)

				if queuedProcess.ProcessState != process.Pending {
					if queuedProcess.ProcessState == process.Finished {
						s.queue.Remove(e)
					}
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
				queuedProcess.ProcessState = process.CanSpawn
			}

			var canSpawnProcess []*process.Process

			for e := s.queue.Front(); e != nil; e = e.Next() {
				queuedProcess := e.Value.(*process.Process)

				if queuedProcess.ProcessState == process.CanSpawn {
					canSpawnProcess = append(canSpawnProcess, queuedProcess)
				}
			}

			if len(canSpawnProcess) == 0 {
				log.Printf("no ready process")
				continue
			}

			sort.Slice(canSpawnProcess, func(i, j int) bool {
				return canSpawnProcess[i].IssuedTime.Before(canSpawnProcess[j].IssuedTime)
			})

			shouldSpawnProcess := canSpawnProcess[0]
			shouldSpawnProcess.ProcessState = process.Active

			ch := make(chan bool)
			go shouldSpawnProcess.Spawn(&ch)
			s.ProcessEventHandler.AddTaskStatusChannel(shouldSpawnProcess.Id, &ch)

			for e := s.queue.Front(); e != nil; e = e.Next() {
				queuedProcess := e.Value.(*process.Process)

				if queuedProcess.ProcessState == process.CanSpawn {
					queuedProcess.ProcessState = process.Pending
				}
			}
		}
	}
}

func (s *GreedyScheduler) OnSuccess(id string) {
	for e := s.queue.Front(); e != nil; e = e.Next() {
		p := e.Value.(*process.Process)

		if p.Id == id {
			p.ProcessState = process.Finished
			log.Printf("finish to exec %s", id)
			break
		}
	}
}

func (s *GreedyScheduler) OnError(id string) {
	for e := s.queue.Front(); e != nil; e = e.Next() {
		p := e.Value.(*process.Process)

		if p.Id == id {
			p.ProcessState = process.CanSpawn
			log.Printf("failed to spawn %s", id)
			break
		}
	}
}

func NewGreedyScheduler(maxPendingQueueSize, gpuInfoRequestInterval, memoryUsageLowWatermark int) *GreedyScheduler {
	targetGpuId := make(chan []int)
	watcher := watcher.NewAgent(gpuInfoRequestInterval, memoryUsageLowWatermark)
	go watcher.Run(targetGpuId)

	scheduler := GreedyScheduler{
		queue:               list.New(),
		watcher:             watcher,
		targetGpuId:         targetGpuId,
		maxPendingQueueSize: maxPendingQueueSize,
	}

	processEventHandler := NewProcessEventHandler(&scheduler)
	go processEventHandler.Run()

	scheduler.ProcessEventHandler = processEventHandler
	return &scheduler
}
