package entrypoint

import (
	"www.seawise.com/shrimps/backend/capture"
	"www.seawise.com/shrimps/backend/core"
	"www.seawise.com/shrimps/backend/log"
	"www.seawise.com/shrimps/backend/streamer"
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

	p.streamer.Produce()
	p.streamer.Start()
	cleanSigTerm.WaitForTermination()
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