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

package process

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/Shikugawa/gpupipe/pkg/types"
	"github.com/google/uuid"
)

type Process struct {
	Id           string       `json:"id"`
	Pid          int          `json:"pid"`
	RootPath     string       `json:"rootpath"`
	Command      []string     `json:"command"`
	IssuedTime   time.Time    `json:"issued_time"`
	GpuId        []int        `json:"gpu_id"`
	ProcessState ProcessState `json:"process_state"`
	LogPath      string       `json:"log_path"`
	ErrLogPath   string       `json:"err_log_path"`
}

func (p *Process) Spawn(ch *chan bool) {
	var outFd *os.File
	var errFd *os.File

	if len(p.LogPath) != 0 {
		tmpFd, err := os.OpenFile(p.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			*ch <- false
		}
		outFd = tmpFd
	} else {
		outFd, _ = os.Open(os.DevNull)
	}
	defer outFd.Close()

	if len(p.ErrLogPath) != 0 {
		tmpFd, err := os.OpenFile(p.ErrLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			*ch <- false
		}
		outFd = tmpFd
	} else {
		errFd, _ = os.Open(os.DevNull)
	}

	defer errFd.Close()

	cmd := exec.Command(p.Command[0], p.Command[1:]...)
	log.Println(cmd.String())
	cmd.Dir = p.RootPath
	cmd.Stdout = outFd
	cmd.Stderr = errFd

	if err := cmd.Start(); err != nil {
		log.Println(err)
		*ch <- false
	}

	p.Pid = cmd.Process.Pid

	// TODO: プロセスが異常終了した場合はfalseにしたい
	cmd.Wait()
	*ch <- true
}

func (p *Process) Terminate() error {
	if p.ProcessState != Active || p.Pid == 0 {
		return fmt.Errorf("this process has stopped already")
	}
	if err := syscall.Kill(p.Pid, syscall.SIGTERM); err != nil {
		return err
	}
	return nil
}

func NewProcess(r *types.ProcessPublishRequest) *Process {
	return &Process{
		Id:           uuid.NewString(),
		RootPath:     r.RootPath,
		Command:      r.Command,
		IssuedTime:   time.Now(),
		GpuId:        r.TargetGpu,
		ProcessState: Pending,
		LogPath:      r.LogPath,
		ErrLogPath:   r.ErrLogPath,
	}
}
