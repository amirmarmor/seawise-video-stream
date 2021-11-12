package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/backend/log"
)

type Config = struct {
	Host string
	Port int
}

var StreamerConfig Config

func InitFlags() {
	flag.StringVar(&StreamerConfig.Host, "streamer-host", "localhost", "The streamer host")
	flag.IntVar(&StreamerConfig.Port, "streamer-port", 8000, "The streamer port")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(StreamerConfig)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
