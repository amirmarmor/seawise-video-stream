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

func (l *Listener) runListener() {
	for {
		l.TCPListenerMutex.Lock()
		l.FrameMutex.Lock()
		conn, err := l.TCPListener.Accept()
		if err != nil {
			log.Warn(fmt.Sprintf("broken connection: %v", err))
			continue
		}
		//go sendTimeStamp(conn)
		go l.handleConn(conn)
	}
}

func (l *Listener) handleConn(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Warn(fmt.Sprintf("tcp conn close failed! - %v", err))
		}
	}(conn)

	defer l.TCPListenerMutex.Unlock()

	l.FrameMutex.Unlock()

	reader := bufio.NewReader(conn)

	timeStamp := make([]byte, l.timeStampPacketSize)
	contentLength := make([]byte, l.contentLengthPacketSize)

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

		l.FrameMutex.Lock()
		l.Frame = new(bytes.Buffer)
		l.Frame.Write(buf)
		l.FrameMutex.Unlock()
	}
}
