package gpiped

import (
	"time"
)

type Watcher struct {
	gpuInfoRequestInterval  time.Duration
	memoryUsageLowWatermark int
}

func NewWatcher() *Watcher {
	return &Watcher{
		gpuInfoRequestInterval:  5 * time.Second,
		memoryUsageLowWatermark: 10,
	}
}

func (w *Watcher) Run(ch chan<- []int) {
	for {
		infos, err := GetGpuInfo()
		if err != nil {
			continue
		}
		var target []int
		for _, info := range infos {
			if info.MemoryUsage > w.memoryUsageLowWatermark {
				target = append(target, info.Index)
			}
		}
		ch <- target
		time.Sleep(w.gpuInfoRequestInterval)
	}
}
