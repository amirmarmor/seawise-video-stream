package entrypoint

import (
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

	cleanSigTerm := Produce()

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

func (p *EntryPoint) startListening() {

}
