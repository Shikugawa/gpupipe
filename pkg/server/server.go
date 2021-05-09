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

package server

import (
	"encoding/json"
	"net/http"

	"github.com/Shikugawa/gpupipe/pkg/scheduler"
	"github.com/Shikugawa/gpupipe/pkg/types"
)

type Server struct {
	schedular *scheduler.Scheduler
}

func (e *Server) HandlePublish(w http.ResponseWriter, r *http.Request) {
	var request types.ProcessPublishRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusInternalServerError)
	}

	if err := e.schedular.Publish(&request); err != nil {
		http.Error(w, "Failed to publish process", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

func (e *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	b, err := e.schedular.List()
	if err != nil {
		http.Error(w, "Failed to fetch process", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write(b)
}

func NewServer(s *scheduler.Scheduler) *Server {
	return &Server{
		schedular: s,
	}
}
