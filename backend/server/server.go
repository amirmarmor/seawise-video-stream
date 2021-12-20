package server

import (
	"bytes"
	"fmt"
	"gocv.io/x/gocv"
	"net"
	"sync"
	"www.seawise.com/backend/core"
	"www.seawise.com/common/log"
)

type Server struct {
	TCPListener             net.Listener
	TCPListenerMutex        sync.Mutex
	Frame                   *bytes.Buffer
	FrameMutex              sync.RWMutex
	timeStampPacketSize     uint
	contentLengthPacketSize uint
}

func Create(devices *core.Devices) ([]*Server, error) {
	servers := make([]*Server, 0)
	for _, device := range devices.List {
		for ch := 0; ch < device.Channels; ch++ {
			port := core.Config.Port + (device.Id * 10) + ch
			server, err := NewServer(port)
			if err != nil {
				return nil, fmt.Errorf("failed to create server - %v", err)
			}
			go server.Run()
			servers = append(servers, server)
		}
	}

	return servers, nil
}

func NewServer(port int) (*Server, error) {

	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})

	if err != nil {
		return nil, fmt.Errorf("generate tcp server failed! - %v", err)
	}

	buf := new(bytes.Buffer)

	server := &Server{
		TCPListener:             tcpListener,
		TCPListenerMutex:        sync.Mutex{},
		Frame:                   buf,
		FrameMutex:              sync.RWMutex{},
		timeStampPacketSize:     8,
		contentLengthPacketSize: 8,
	}
	log.V5(fmt.Sprintf("Listening on 127.0.0.1:%v", port))
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
