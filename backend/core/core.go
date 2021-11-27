package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/common/log"
)

type Config = struct {
	Host                    string
	Port                    int
	TimeStampPacketSize     int
	ContentLengthPacketSize int
	HeadPacketSize          int
}

var StreamerConfig Config

func InitFlags() {
	flag.StringVar(&StreamerConfig.Host, "streamer-host", "localhost", "The streamer host")
	flag.IntVar(&StreamerConfig.Port, "streamer-port", 8000, "The streamer port")
	flag.IntVar(&StreamerConfig.TimeStampPacketSize, "timestamp-packet-size", 8, "Timestamp packet size")
	flag.IntVar(&StreamerConfig.ContentLengthPacketSize, "content-length-packet-size", 8, "content length packet size")
	flag.IntVar(&StreamerConfig.HeadPacketSize, "head-packet-size", 64, "head packet size")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(StreamerConfig)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
