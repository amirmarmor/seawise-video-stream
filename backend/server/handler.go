package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"www.seawise.com/common/log"
)

func (s *Server) runListener() {
	for {
		s.TCPListenerMutex.Lock()
		s.FrameMutex.Lock()
		conn, err := s.TCPListener.Accept()
		if err != nil {
			log.Warn(fmt.Sprintf("broken connection: %v", err))
			continue
		}
		//go sendTimeStamp(conn)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Warn(fmt.Sprintf("tcp conn close failed! - %v", err))
		}
	}(conn)

	defer s.TCPListenerMutex.Unlock()

	s.FrameMutex.Unlock()

	reader := bufio.NewReader(conn)

	timeStamp := make([]byte, s.timeStampPacketSize)
	contentLength := make([]byte, s.contentLengthPacketSize)

	for {
		_, err := io.ReadFull(reader, timeStamp)
		if err != nil {
			log.Warn(fmt.Sprintf("%v Down! - %v", conn.RemoteAddr().String(), err))
			break
		}

		_, err = io.ReadFull(reader, contentLength)
		if err != nil {
			log.Warn(fmt.Sprintf("%v Down! - %v", conn.RemoteAddr().String(), err))
			break
		}

		length := int64(binary.LittleEndian.Uint64(contentLength))

		buf := make([]byte, length)
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			log.Warn(fmt.Sprintf("%v Down! - %v", conn.RemoteAddr().String(), err))
			break
		}

		s.FrameMutex.Lock()
		s.Frame = new(bytes.Buffer)
		s.Frame.Write(buf)
		s.FrameMutex.Unlock()
	}
}
