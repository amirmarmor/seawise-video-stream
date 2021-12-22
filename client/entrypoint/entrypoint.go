package entrypoint

import (
	"net/http"
	"strings"
	"www.seawise.com/client/capture"
	"www.seawise.com/client/core"
	"www.seawise.com/common/log"
)

type EntryPoint struct {
	manager *core.ConfigManager
	capt    *capture.Capture
}

func (p *EntryPoint) Run() {
	core.InitFlags()
	log.InitFlags()

	log.ParseFlags()
	log.Info("Starting")

	//p.buildBlocks()

	cleanSigTerm := Produce()
	//go p.capt.Start()

	cleanSigTerm.WaitForTermination()
}

func (p *EntryPoint) buildBlocks() {
	var err error
	p.manager, err = core.Produce()
	if err != nil {
		panic(err)
	}

	p.capt = capture.Create(p.manager, 5)
	p.capt.Init()
}

func (p *EntryPoint) addHandlers() {
	http.HandleFunc("/", p.handler)
}

func (p *EntryPoint) handler(w http.ResponseWriter, r *http.Request) {
	response := "ok"
	switch action := strings.TrimPrefix(r.URL.Path, "/"); action {
	case "start":
		go p.capt.Start()
		response = "starting..."
	case "stop":
		//go p.capt.Stop()
		response = "stopping..."
	}
	_, err := w.Write([]byte(response))
	if err != nil {
		panic(err)
	}
}
