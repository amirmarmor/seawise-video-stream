package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/common/log"
)

type HostsInfo = struct {
	Stream      string
	StreamPort  int
	Backend     string
	BackendPort int
}

var Hosts HostsInfo

func InitFlags() {
	flag.StringVar(&Hosts.Backend, "backend-host", "localhost", "The backend host")
	flag.IntVar(&Hosts.BackendPort, "backend-port", 5000, "The backend port")
	flag.StringVar(&Hosts.Stream, "stream-host", "localhost", "The stream host")
	flag.IntVar(&Hosts.StreamPort, "stream-port", 8000, "The stream port")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(Hosts)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
