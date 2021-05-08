package main

import (
	"net/http"

	"github.com/Shikugawa/gpupipe/pkg/scheduler"
)

func main() {
	sched := scheduler.NewGreedyScheduler(10)
	srv := NewServer(sched)

	http.HandleFunc("/publish", srv.HandlePublish)
	http.ListenAndServe("0.0.0.0:8000", nil)

	// TODO: cleanup all pending processes
}
