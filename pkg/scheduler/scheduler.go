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

	"github.com/Shikugawa/gpupipe/pkg/gpu"
	"github.com/Shikugawa/gpupipe/pkg/process"
	"github.com/Shikugawa/gpupipe/pkg/types"
	"github.com/Shikugawa/gpupipe/pkg/watcher"
)

type Scheduler struct {
	Queue                          *list.List
	Watcher                        *watcher.Agent
	TargetGpuInfos                 chan []gpu.GpuInfo
	MaxPendingQueueSize            int
	ProcessEventHandler            *ProcessEventHandler
	SchedulePlugin                 SchedulePlugin
	defaultMemoryUsageLowWatermark int
}

func (s *Scheduler) Publish(r *types.ProcessPublishRequest) error {
	if s.Queue.Len() >= s.MaxPendingQueueSize {
		return fmt.Errorf("failed to publish pending process with queue size overflow")
	}
	s.Queue.PushBack(process.NewProcess(r))
	return nil
}

func (s *Scheduler) List() ([]byte, error) {
	processSet := make(map[string][]process.Process)
	processSet["processes"] = make([]process.Process, 0)

	for e := s.Queue.Front(); e != nil; e = e.Next() {
		queuedProcess := e.Value.(*process.Process)
		processSet["processes"] = append(processSet["processes"], *queuedProcess)
	}

	b, err := json.Marshal(processSet)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *Scheduler) Delete(id string) bool {
	for e := s.Queue.Front(); e != nil; e = e.Next() {
		queuedProcess := e.Value.(*process.Process)
		if queuedProcess.Id == id {
			s.terminateActiveProcess(queuedProcess)
			s.Queue.Remove(e)
			return true
		}
	}
	return false
}

func (s *Scheduler) TerminateAllActiveProcess() {
	for e := s.Queue.Front(); e != nil; e = e.Next() {
		s.terminateActiveProcess(e.Value.(*process.Process))
	}
}

func (s *Scheduler) terminateActiveProcess(p *process.Process) error {
	if p.ProcessState == process.Active {
		if err := p.Terminate(); err != nil {
			return err
		}
		p.ProcessState = process.Finished
	}
	return nil
}

func (s *Scheduler) Run() {
	for {
		currentTargetGpuInfos := <-s.TargetGpuInfos
		for e := s.Queue.Front(); e != nil; e = e.Next() {
			queuedProcess := e.Value.(*process.Process)

			if queuedProcess.ProcessState != process.Pending {
				if queuedProcess.ProcessState == process.Finished {
					s.Queue.Remove(e)
				}
				continue
			}

			canSpawn := true

			for _, requestGpuId := range queuedProcess.GpuId {
				requestGpuIdAvailable := true

				for _, gpuInfo := range currentTargetGpuInfos {
					if requestGpuId == gpuInfo.Index {
						memoryUsageLowWatermark := 0

						if queuedProcess.MemoryUsageLowWatermark == 0 {
							memoryUsageLowWatermark = s.defaultMemoryUsageLowWatermark
						} else {
							memoryUsageLowWatermark = queuedProcess.MemoryUsageLowWatermark
						}
						if memoryUsageLowWatermark < gpuInfo.MemoryUsage {
							requestGpuIdAvailable = false
							break
						}
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

		for e := s.Queue.Front(); e != nil; e = e.Next() {
			queuedProcess := e.Value.(*process.Process)

			if queuedProcess.ProcessState == process.CanSpawn {
				canSpawnProcess = append(canSpawnProcess, queuedProcess)
			}
		}

		if len(canSpawnProcess) == 0 {
			log.Printf("no ready process")
			continue
		}

		shouldSpawnProcess := s.SchedulePlugin.Select(canSpawnProcess)
		shouldSpawnProcess.ProcessState = process.Active

		ch := make(chan bool)
		go shouldSpawnProcess.Spawn(&ch)
		s.ProcessEventHandler.AddTaskStatusChannel(shouldSpawnProcess.Id, &ch)

		for e := s.Queue.Front(); e != nil; e = e.Next() {
			queuedProcess := e.Value.(*process.Process)

			if queuedProcess.ProcessState == process.CanSpawn {
				queuedProcess.ProcessState = process.Pending
			}
		}
	}
}

func (s *Scheduler) OnSuccess(id string) {
	for e := s.Queue.Front(); e != nil; e = e.Next() {
		p := e.Value.(*process.Process)

		if p.Id == id {
			p.ProcessState = process.Finished
			log.Printf("finish to exec %s", id)
			break
		}
	}
}

func (s *Scheduler) OnError(id string) {
	for e := s.Queue.Front(); e != nil; e = e.Next() {
		p := e.Value.(*process.Process)

		if p.Id == id {
			p.ProcessState = process.CanSpawn
			log.Printf("failed to spawn %s", id)
			break
		}
	}
}

func NewScheduler(maxPendingQueueSize, gpuInfoRequestInterval, defaultMemoryUsageLowWatermark int, plugin SchedulePlugin) *Scheduler {
	targetGpuInfos := make(chan []gpu.GpuInfo)
	watcher := watcher.NewAgent(gpuInfoRequestInterval)
	go watcher.Run(targetGpuInfos)

	if defaultMemoryUsageLowWatermark > 100 {
		defaultMemoryUsageLowWatermark = 100
	}

	if defaultMemoryUsageLowWatermark < 0 {
		defaultMemoryUsageLowWatermark = 0
	}

	scheduler := Scheduler{
		Queue:                          list.New(),
		Watcher:                        watcher,
		TargetGpuInfos:                 targetGpuInfos,
		MaxPendingQueueSize:            maxPendingQueueSize,
		SchedulePlugin:                 plugin,
		defaultMemoryUsageLowWatermark: defaultMemoryUsageLowWatermark,
	}

	processEventHandler := NewProcessEventHandler(&scheduler)
	go processEventHandler.Run()

	scheduler.ProcessEventHandler = processEventHandler
	return &scheduler
}
