package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/common/log"
)

type ApiInfo = struct {
	ApiHost                 string
	ApiPort                 string
	StreamerHost            string
	StreamerPort            int
	TimeStampPacketSize     int
	ContentLengthPacketSize int
	HeadPacketSize          int
}

var Api ApiInfo

func InitFlags() {
	flag.StringVar(&Api.ApiHost, "api-host", "localhost", "The api host")
	flag.StringVar(&Api.ApiPort, "api-port", "5000", "The api host")
	flag.StringVar(&Api.StreamerHost, "streamer-host", "localhost:5000", "The api host")
	flag.IntVar(&Api.StreamerPort, "streamer-port", 8000, "The streamer port")
	flag.IntVar(&Api.TimeStampPacketSize, "timestamp-packet-size", 8, "Timestamp packet size")
	flag.IntVar(&Api.ContentLengthPacketSize, "content-length-packet-size", 8, "content length packet size")
	flag.IntVar(&Api.HeadPacketSize, "head-packet-size", 64, "head packet size")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(Api)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
