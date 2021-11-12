package capture

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"www.seawise.com/backend/capture/core"
	"www.seawise.com/backend/log"
)

type Streamer struct {
	TCPConn                 *net.TCPConn
	offset                  int
	queue                   chan []byte
	TimeStampPacketSize     int
	ContentLengthPacketSize int
	HeadPacketSize          int
}

func CreateStreamer(offset int, queue chan []byte) *Streamer {
	streamer := &Streamer{
		TimeStampPacketSize:     8,
		ContentLengthPacketSize: 8,
		HeadPacketSize:          64,
		queue:                   queue,
		offset:                  offset,
	}

	streamer.connect()
	return streamer
}

func (s *Streamer) connect() {
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.ParseIP(core.StreamerConfig.Host),
		Port: core.StreamerConfig.Port + s.offset,
	})

	if err != nil {
		log.Warn(fmt.Sprintf("generate udp client failed! - %v", err))
		time.Sleep(time.Second * 3)
		go s.connect()
		return
	}

	s.TCPConn = conn
	go s.handleSend()
	return
}

func (s *Streamer) handleSend() {
	writer := bufio.NewWriter(s.TCPConn)
	for pkt := range s.queue {
		_, err := writer.Write(s.pack(pkt))
		if err != nil {
			log.Warn(fmt.Sprintf("Packet Send Failed! - %v", err))
			go s.connect()
			return
		}
	}
}

func (s *Streamer) pack(frame []byte) []byte {
	// ------ Packet ------
	// timestamp (8 bytes)
	// content-length (8 bytes)
	// content (content-length bytes)
	// ------  End   ------

	timePkt := make([]byte, s.TimeStampPacketSize)
	binary.LittleEndian.PutUint64(timePkt, uint64(time.Now().UnixNano()))

	contentLengthPkt := make([]byte, s.ContentLengthPacketSize)
	binary.LittleEndian.PutUint64(contentLengthPkt, uint64(len(frame)))

	var pkt []byte
	pkt = append(pkt, timePkt...)
	pkt = append(pkt, contentLengthPkt...)
	pkt = append(pkt, frame...)

	return pkt
}

//func (s *Streamer) Stop(capture *capture.Capture) {
//	for _, channel := range capture.Channels {
//		//channel.Stream.Close()
//	}
//	s.Cancel()
//}
