package server

import (
	"bytes"
	"fmt"
	"gocv.io/x/gocv"
	"net"
	"sync"
	"www.seawise.com/backend/core"
)

type Server struct {
	TCPListener      net.Listener
	TCPListenerMutex sync.Mutex
	Frame            *bytes.Buffer
	FrameMutex       sync.RWMutex
}

func Create(channels int) ([]*Server, error){
	sockets := make([]*Server, 0)
	for i := 0; i < channels; i++ {
		socket, err := NewServer(i)
		if err != nil {
			return nil, fmt.Errorf("failed to create new server for channel %v: %v", i, err)
		}

		sockets = append(sockets, socket)
	}

	return sockets, nil
}

func NewServer(channel int) (*Server, error) {
	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(core.StreamerConfig.Host),
		Port: core.StreamerConfig.Port + channel,
	})

	if err != nil {
		return nil, fmt.Errorf("generate tcp server failed! - %v", err)
	}

	buf := new(bytes.Buffer)

	server := &Server{
		TCPListener:      tcpListener,
		TCPListenerMutex: sync.Mutex{},
		Frame:            buf,
		FrameMutex:       sync.RWMutex{},
	}

	return server, nil
}

func (s *Server) Run() {
	defer func(tcpListener net.Listener) {
		err := tcpListener.Close()
		if err != nil {
			panic(err)
		}
	}(s.TCPListener)

	go s.runListener()

	window := gocv.NewWindow("Streaming")
	for {
		s.FrameMutex.RLock()
		mat, err := gocv.IMDecode(s.Frame.Bytes(), gocv.IMReadUnchanged)
		if err != nil {
			s.FrameMutex.RUnlock()
			continue
		}
		s.FrameMutex.RUnlock()
		window.IMShow(mat)
		window.WaitKey(1)
		_ = mat.Close()
	}
}
