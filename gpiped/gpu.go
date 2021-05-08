package gpiped

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

var query = []string{
	"index",
	"uuid",
	"name",
	"timestamp",
	"memory.total",
	"memory.free",
	"memory.used",
	"utilization.gpu",
	"utilization.memory",
}

type GpuInfo struct {
	Index       int    `json:"index"`
	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	timestamp   string `json:"timestamp"`
	TotalMemory int64  `json:"memory.total"`
	MemoryFree  int64  `json:"memory.free"`
	MemoryUsed  int64  `json:"memory.used"`
	GpuUsage    int    `json:"utilization.gpu"`
	MemoryUsage int    `json:"utilization.memory"`
}

func GetGpuInfo() ([]GpuInfo, error) {
	cmd := exec.Command("nvidia-smi", fmt.Sprintf("--query-gpu=%s", strings.Join(query, ","), "--format=csv,noheader,nounits"))

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if errbuf.Len() != 0 {
		return nil, fmt.Errorf(errbuf.String())
	}

	result := strings.Split(outbuf.String(), "\n")

	var infos []GpuInfo

	for _, r := range result {
		var infomap map[string]string
		line := strings.Split(r, ", ")

		for i := 0; i < len(line); i++ {
			infomap[query[i]] = line[i]
		}

		infomapbytes, err := json.Marshal(infomap)
		if err != nil {
			return nil, err
		}

		var info GpuInfo
		if err := json.Unmarshal(infomapbytes, &info); err != nil {
			return nil, err
		}

		infos = append(infos, info)
	}

	return infos, nil
}

func CheckGpuId(id []int) bool {
	infos, err := GetGpuInfo()
	if err != nil {
		return false
	}

	for _, i := range id {
		for _, info := range infos {
			if info.Index == i {
				continue
			}
		}
		return false
	}

	return true
}
