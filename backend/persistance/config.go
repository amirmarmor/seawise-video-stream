package persistance

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/shrimps/backend/log"
)

type RedisConfiguration struct {
	Host string
}

var RedisConfig RedisConfiguration

func InitFlags() {
	flag.StringVar(&RedisConfig.Host, "host", "localhost", "redis host")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(RedisConfig)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
