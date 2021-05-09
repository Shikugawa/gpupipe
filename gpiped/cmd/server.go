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

package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Shikugawa/gpupipe/pkg/scheduler"
	"github.com/Shikugawa/gpupipe/pkg/scheduler/plugin"
	"github.com/Shikugawa/gpupipe/pkg/server"
	"github.com/spf13/cobra"
)

var (
	maxPendingQueueSize     int16
	gpuInfoRequestInterval  int16
	memoryUsageLowWatermark int8
	port                    int16

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run gpiped server",
		Run: func(cmd *cobra.Command, args []string) {
			sched := scheduler.NewScheduler(
				int(maxPendingQueueSize), int(gpuInfoRequestInterval), int(memoryUsageLowWatermark), plugin.NewGreedyPlugin())
			go sched.Run()

			srv := server.NewServer(sched).Start(strconv.Itoa(int(port)))

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)
			<-sig

			sched.TerminateActiveProcess()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Println("failed to graceful shutdown server:", err)
				return
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().Int16VarP(&port, "port", "p", 8000, "server port")
	runCmd.Flags().Int16VarP(&maxPendingQueueSize, "queue_size", "q", 10, "the number of pending queue limit")
	runCmd.Flags().Int8VarP(&memoryUsageLowWatermark, "memory_usage_low_watermark", "m", 10, "low usage watermark whether to issue or not GPU task")
	runCmd.Flags().Int16VarP(&gpuInfoRequestInterval, "request_interval", "r", 5, "interval to request gpu usage for GPU watcher agent")
}
