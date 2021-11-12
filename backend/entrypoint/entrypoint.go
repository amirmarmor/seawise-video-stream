package entrypoint

import (
	"www.seawise.com/backend/capture"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/log"
	captureCore "www.seawise.com/backend/capture/core"
)

type EntryPoint struct {
	manager  *core.ConfigManager
	capt     *capture.Capture
}

func (p *EntryPoint) Run() {
	core.InitFlags()
	log.InitFlags()
	captureCore.InitFlags()

	log.ParseFlags()
	log.Info("Starting")

	p.buildBlocks()

	cleanSigTerm := Produce()
	go p.capt.Start()


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
