package entrypoint

import (
	"fmt"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/server"
	"www.seawise.com/common/log"
)

type EntryPoint struct {
	servers []*server.Server
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
	for i := 0; i < 3; i++ {
		s, err := server.NewServer(i)
		if err != nil {
			panic(fmt.Errorf("Failed to create server %v: %v", i, err))
		}
		s.Run()
		p.servers = append(p.servers, s)

	}
}

