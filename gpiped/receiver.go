package gpiped

import (
	"encoding/json"
	"net/http"
)

type ProcessRequest struct {
	Command   string `json:"command"`
	TargetGpu []int  `json:"target_gpu"`
}

type Receiver struct {
	schedular Scheduler
}

func (e *Receiver) HandlePublish(w http.ResponseWriter, r *http.Request) {
	var request ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusInternalServerError)
	}

	if err := e.schedular.Publish(request.Command, request.TargetGpu); err != nil {
		http.Error(w, "Failed to publish process", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func NewReceiver(s Scheduler) *Receiver {
	return &Receiver{
		schedular: s,
	}
}
