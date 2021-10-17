package streamer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"www.seawise.com/shrimps/backend/mjpeg"
)

type Streamer struct {
	server *http.Server
	client *http.ServeMux
}

func Create() *Streamer {
	c := http.NewServeMux()
	s := &http.Server{
		Addr:    ":8080",
		Handler: c,
	}
	return &Streamer{
		client: c,
		server: s,
	}
}

func (s *Streamer) Produce(channel int, stream *mjpeg.Stream) {
	path := "/stream/" + strconv.Itoa(channel)
	s.client.Handle(path, stream)
}

func (s *Streamer) Start() {
	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (s *Streamer) Stop() error {
	ctx := context.Background()
	err := s.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop stream server: %v", err)
	}
	return nil
}
