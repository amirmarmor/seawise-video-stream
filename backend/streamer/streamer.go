package streamer

import (
	"net/http"
	"strconv"
	"www.seawise.com/shrimps/backend/capture"
)

type Streamer struct {
	Client *http.ServeMux
	Capt   *capture.Capture
}

func Create(capt *capture.Capture) *Streamer {
	s := http.NewServeMux()

	return &Streamer{
		Client: s,
		Capt: capt,
	}
}

func (s *Streamer) Produce(){
	for i:=0; i < len(s.Capt.Channels); i++ {
		path := "/stream/" + strconv.Itoa(i)
		s.Client.Handle(path, s.Capt.Channels[i].Stream)
	}
}

func (s *Streamer) Start() {
	err := http.ListenAndServe(":8080", s.Client)
	if err != nil {
		panic(err)
	}
}

