package main

import (
	"encoding/json"
	"net/http"

	"github.com/Shikugawa/gpupipe/pkg/scheduler"
)

type ProcessRequest struct {
	Command   string `json:"command"`
	TargetGpu []int  `json:"target_gpu"`
}

type Server struct {
	schedular scheduler.Scheduler
}

func (e *Server) HandlePublish(w http.ResponseWriter, r *http.Request) {
	var request ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusInternalServerError)
	}

	if err := e.schedular.Publish(request.Command, request.TargetGpu); err != nil {
		http.Error(w, "Failed to publish process", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func NewServer(s scheduler.Scheduler) *Server {
	return &Server{
		schedular: s,
	}
}
