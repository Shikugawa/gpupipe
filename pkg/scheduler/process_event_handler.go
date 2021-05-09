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
	"time"
)

type ProcessEventHandler struct {
	callback           SchedulerCallback
	TaskStatusChannels map[string]*chan bool
}

func (p *ProcessEventHandler) AddTaskStatusChannel(id string, ch *chan bool) {
	p.TaskStatusChannels[id] = ch
}

func (p *ProcessEventHandler) Run() {
	for {
		// TODO: ここでロック取るとデッドロックしそう
		for id, ch := range p.TaskStatusChannels {
			select {
			case status := <-*ch:
				if status {
					p.callback.OnSuccess(id)
				} else {
					p.callback.OnError(id)
				}
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}
}

func NewProcessEventHandler(callback SchedulerCallback) *ProcessEventHandler {
	return &ProcessEventHandler{
		callback:           callback,
		TaskStatusChannels: make(map[string]*chan bool),
	}
}
