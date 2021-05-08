package gpiped

import (
	"fmt"
	"log"
	"sort"
)

type Scheduler interface {
	Publish(command string, gpuId []int) error
	Run()
}

type GreedyScheduler struct {
	queue               []*Process
	watcher             *Watcher
	targetGpuId         chan []int
	maxPendingQueueSize int
}

func (s *GreedyScheduler) Publish(command string, gpuId []int) error {
	if len(s.queue) >= s.maxPendingQueueSize {
		return fmt.Errorf("Failed to publish pending process with queue size overflow")
	}
	s.queue = append(s.queue, NewProcess(command, gpuId))
	return nil
}

func (s *GreedyScheduler) Run() {
	for {
		select {
		case ids := <-s.targetGpuId:
			for _, p := range s.queue {
				if p.ProcessState != Pending {
					continue
				}

				canSpawn := false

				for _, requiredGpuId := range p.GpuId {
					found := false

					for _, availableGpuId := range ids {
						if availableGpuId == requiredGpuId {
							found = true
							continue
						}
					}

					if !found {
						canSpawn = false
						break
					}
				}

				if canSpawn {
					p.ProcessState = CanSpawn
				}
			}

			var canSpawnProcesses []*Process

			for _, p := range s.queue {
				if p.ProcessState == CanSpawn {
					canSpawnProcesses = append(canSpawnProcesses, p)
				}
			}

			sort.Slice(canSpawnProcesses, func(i, j int) bool {
				return canSpawnProcesses[i].issuedTime.Before(canSpawnProcesses[j].issuedTime)
			})

			issueProcess := canSpawnProcesses[0]
			issueProcess.ProcessState = Active
			if err := issueProcess.Spawn(); err != nil {
				issueProcess.ProcessState = CanSpawn
				log.Println("failed to spawn")
			}

			for _, p := range s.queue {
				if p.ProcessState == CanSpawn {
					p.ProcessState = Pending
				}
			}
		}
	}
}

func NewGreedyScheduler() *GreedyScheduler {
	targetGpuId := make(chan []int)
	watcher := NewWatcher()
	watcher.Run(targetGpuId)
	maxPendingQueueSize := 10

	return &GreedyScheduler{
		queue:               make([]*Process, maxPendingQueueSize),
		watcher:             watcher,
		targetGpuId:         targetGpuId,
		maxPendingQueueSize: maxPendingQueueSize,
	}
}
