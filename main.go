package gpiped

import (
	"net/http"

	"github.com/Shikugawa/gpupipe/gpiped"
)

func main() {
	s := gpiped.NewGreedyScheduler()
	r := gpiped.NewReceiver(s)

	http.HandleFunc("/publish", r.HandlePublish)

	http.ListenAndServe("0.0.0.0:8000", nil)

	// TODO: cleanup all pending processes
}
