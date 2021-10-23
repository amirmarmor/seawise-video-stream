package entrypoint

import (
	"www.seawise.com/backend/capture"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/log"
	"www.seawise.com/backend/streamer"
)

type EntryPoint struct {
	manager  *core.ConfigManager
	capt     *capture.Capture
	streamer *streamer.Streamer
}

func (p *EntryPoint) Run() {
	core.InitFlags()
	log.InitFlags()

	log.ParseFlags()
	log.Info("Starting")

	p.buildBlocks()

	cleanSigTerm := Produce()
	go p.capt.Start()

	err := p.streamer.Server.ListenAndServe()
	if err != nil {
		log.Error("FAILED TO START SERVER", err)
	}

	p.streamer.Stop(p.capt)

	cleanSigTerm.WaitForTermination(*p.streamer)
}

func (p *EntryPoint) buildBlocks() {
	var err error
	p.manager, err = core.Produce()
	if err != nil {
		panic(err)
	}

	p.capt = capture.Create(p.manager)
	p.capt.Init()

	p.streamer = streamer.Create(p.capt)
}
