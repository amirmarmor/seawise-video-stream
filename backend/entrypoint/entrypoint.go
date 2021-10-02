package entrypoint

import (
	"www.seawise.com/shrimps/backend/capture"
	"www.seawise.com/shrimps/backend/core"
	"www.seawise.com/shrimps/backend/executors/configuration"
	"www.seawise.com/shrimps/backend/exposed"
	"www.seawise.com/shrimps/backend/log"
	"www.seawise.com/shrimps/backend/persistance"
	"www.seawise.com/shrimps/backend/streamer"
	"www.seawise.com/shrimps/backend/web"
)

type EntryPoint struct {
	persist *persistance.Persist
	web     *web.Web
	manager *core.ConfigManager
	capt    *capture.Capture
	stream  *streamer.Streamer
}

func (p *EntryPoint) Run() {
	log.InitFlags()
	persistance.InitFlags()
	log.ParseFlags()
	log.Info("Starting")

	p.buildBlocks()
	p.addExecutors()

	go p.stream.Start()
	p.web.Start()
}

func (p *EntryPoint) buildBlocks() {
	persist, err := persistance.Create()
	if err != nil {
		panic(err)
	}

	p.persist = persist
	p.web = web.Create()

	p.manager, err = core.Produce(persist)
	if err != nil {
		panic(err)
	}

	p.capt = capture.Create(p.manager)
	p.capt.Init()
	go p.capt.Start()

	p.stream = streamer.Create(p.capt)
	p.stream.Produce()
}

func (p *EntryPoint) addExecutors() {

	configurationExecutor := configuration.Produce(p.manager, p.capt)
	p.web.Client.GET(exposed.GetConfigUrl, configurationExecutor.GetConfig)
	p.web.Client.POST(exposed.SetConfigUrl, configurationExecutor.SetConfig)
	p.web.Client.POST(exposed.ActionUrl, configurationExecutor.DoAction)

}
