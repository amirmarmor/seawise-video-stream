package core

import (
	"encoding/json"
	"github.com/namsral/flag"
	"www.seawise.com/shrimps/backend/log"
)

type Default = struct {
	Offset  int
	Rules   string
	Show		string
	Record 	string
}

var Defaults Default

func InitFlags() {
	defaultRule := `[]`
	defaultShow := "[]"
	defaultRecord := "[]"

	flag.IntVar(&Defaults.Offset, "offset", 0, "The offset of the first camera - if no webcam = 0")
	flag.StringVar(&Defaults.Rules, "rules", defaultRule, "recording schedules")
	flag.StringVar(&Defaults.Show, "show", defaultShow, "which cameras should show online")
	flag.StringVar(&Defaults.Record, "record", defaultRecord, "which cameras should record now")

	log.AddNotify(postParse)
}

func postParse() {
	marshal, err := json.Marshal(Defaults)
	if err != nil {
		log.Fatal("marshal config failed: %v", err)
	}

	log.V5("configuration loaded: %v", string(marshal))
}
