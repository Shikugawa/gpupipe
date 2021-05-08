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
	"os"
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

func ProcessStateToString(state ProcessState) string {
	if state == Pending {
		return "Pending"
	} else if state == CanSpawn {
		return "CanSpawn"
	} else if state == Active {
		return "Active"
	} else if state == Finished {
		return "Finished"
	} else {
		return ""
	}
}

type Process struct {
	rootPath     string       `json:"rootpath"`
	command      string       `json:"command"`
	issuedTime   time.Time    `json:"issued_time"`
	GpuId        []int        `json:"gpu_id"`
	ProcessState ProcessState `json:"process_state"`
}

func (p *Process) Spawn() error {
	err := os.Chdir(p.rootPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(p.command)
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil
}

func NewProcess(rootpath, command string, gpuId []int) *Process {
	return &Process{
		rootPath:     rootpath,
		command:      command,
		issuedTime:   time.Now(),
		GpuId:        gpuId,
		ProcessState: Pending,
	}
}
