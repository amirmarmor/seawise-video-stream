package server

import (
	"bytes"
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"net"
	"sync"
	"www.seawise.com/backend/core"
	"www.seawise.com/common/log"
)

type Server struct {
	Streams   []*Stream
	Listeners []*Listener
}

type Stream struct{}
type Listener struct {
	TCPListener             net.Listener
	TCPListenerMutex        sync.Mutex
	Frame                   *bytes.Buffer
	FrameMutex              sync.RWMutex
	timeStampPacketSize     uint
	contentLengthPacketSize uint
}

func Create(devices *core.Devices) (*Server, error) {
	server := &Server{}

	for _, device := range devices.List {
		for ch := 0; ch < device.Channels; ch++ {
			port := core.Config.Port + (device.Id * 10) + ch
			stream := mjpeg.NewStream()
			listener, err := NewListener(port)
			if err != nil {
				return nil, fmt.Errorf("failed to create listener - %v", err)
			}
			go listener.Run(stream)
			server.Listeners = append(server.Listeners, listener)
		}
	}

	return servers, nil
}

func NewListener(port int) (*Listener, error) {
	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})

	if err != nil {
		return nil, fmt.Errorf("generate tcp server failed! - %v", err)
	}

	buf := new(bytes.Buffer)

	server := &Listener{
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

func (l *Listener) Run(stream *mjpeg.Stream) {
	defer func(tcpListener net.Listener) {
		err := tcpListener.Close()
		if err != nil {
			panic(err)
		}
	}(l.TCPListener)

	go l.runListener()

	//window := gocv.NewWindow("Streaming")
	for {
		l.FrameMutex.RLock()
		stream.UpdateJPEG(l.Frame.Bytes())
		l.FrameMutex.RUnlock()
	}
}
