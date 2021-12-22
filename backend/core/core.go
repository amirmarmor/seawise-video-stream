package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/common/log"
)

type Configuration = struct {
	Port        int
	BackendHost string
	BackendPort int
	DevicesPort int
	//TimeStampPacketSize     int
	//ContentLengthPacketSize int
	//HeadPacketSize          int
}

var Config Configuration

func InitFlags() {
	flag.IntVar(&Config.Port, "port", 8000, "The stream port start")
	flag.IntVar(&Config.BackendPort, "backendport", 5000, "The backend port")
	flag.IntVar(&Config.DevicesPort, "devicesport", 3000, "The backend port")
	flag.StringVar(&Config.BackendHost, "backend-Host", "localhost", "The backend port")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(Config)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
