package entrypoint

import (
	"net/http"
	"strconv"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/server"
	"www.seawise.com/common/log"
)

type EntryPoint struct {
	servers []*server.Server
	devices *core.Devices
}

func (p *EntryPoint) Run() {
	core.InitFlags()
	log.InitFlags()

	log.ParseFlags()
	log.Info("Starting")

	p.buildBlocks()
	p.addHandlers()
	cleanSigTerm := Produce()
	err := http.ListenAndServe("127.0.0.1:8080", nil)
	if err != nil {
		panic(err)
	}
	cleanSigTerm.WaitForTermination()
}

func (p *EntryPoint) buildBlocks() {
	var err error
	p.devices, err = core.Produce()
	if err != nil {
		panic(err)
	}

	p.servers, err = server.Create(p.devices)
	if err != nil {
		panic(err)
	}
}

func (p *EntryPoint) addHandlers() {
	for _, s := range p.servers {
		path := "/stream/" + strconv.Itoa(s.Port)
		http.HandleFunc(path, s.HandleOutbound)
	}
}
