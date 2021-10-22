package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/shrimps/backend/log"
)

type ApiInfo = struct {
	Host string
}

var Api ApiInfo

func InitFlags() {
	flag.StringVar(&Api.Host, "apihost", "localhost:5000", "The api host")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(Api)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
