package streamer

import (
	"context"
	"net/http"
	"strconv"
	"www.seawise.com/backend/capture"
)

type Streamer struct {
	Server *http.Server
	Ctx    *context.Context
	Cancel context.CancelFunc
}

func Create(capture *capture.Capture) *Streamer {
	for i := 0; i < len(capture.Channels); i++ {
		path := "/stream/" + strconv.Itoa(i)
		http.HandleFunc(path, capture.Channels[i].Stream.ServeHTTP)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Streamer{
		Server: &http.Server{
			Addr: ":8080",
		},
		Ctx:    &ctx,
		Cancel: cancel,
	}
}

func (s *Streamer) Stop(capture *capture.Capture) {
	for _, channel := range capture.Channels {
		channel.Stream.Close()
	}
	s.Cancel()
}
