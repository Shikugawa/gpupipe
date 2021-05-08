package gpiped

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
