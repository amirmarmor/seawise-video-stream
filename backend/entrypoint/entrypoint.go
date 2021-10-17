package entrypoint

import (
	"www.seawise.com/shrimps/backend/capture"
	"www.seawise.com/shrimps/backend/core"
	"www.seawise.com/shrimps/backend/log"
	"www.seawise.com/shrimps/backend/streamer"
)

type EntryPoint struct {
	manager *core.ConfigManager
	capt    *capture.Capture
	stream  *streamer.Streamer
}

func (p *EntryPoint) Run() {
	log.InitFlags()
	log.ParseFlags()
	log.Info("Starting")

	p.buildBlocks()
	cleanSigTerm := Produce()
	//
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
	go p.capt.Start()
}

