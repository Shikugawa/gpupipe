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
	"os/exec"
	"time"
)

type ProcessState int

const (
	Pending ProcessState = iota
	CanSpawn
	Active
	Finished // TODO: 今の実装だとプロセス自体が終了しても永遠にFinishedにならない
)

type Process struct {
	command      string
	issuedTime   time.Time
	GpuId        []int
	ProcessState ProcessState
}

func (p *Process) Spawn() error {
	cmd := exec.Command(p.command)
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func NewProcess(command string, gpuId []int) *Process {
	return &Process{
		command:      command,
		issuedTime:   time.Now(),
		GpuId:        gpuId,
		ProcessState: Pending,
	}
}
