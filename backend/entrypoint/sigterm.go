package entrypoint

import (
	"os"
	"os/signal"
	"syscall"
	"www.seawise.com/shrimps/backend/log"
)

type Cleanup func()

type CleanSigTerm struct {
	signalsChannel chan os.Signal
}

func Produce() *CleanSigTerm {
	s := CleanSigTerm{}
	s.signalsChannel = make(chan os.Signal, 1)
	signal.Notify(s.signalsChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	return &s
}

func (s *CleanSigTerm) WaitForTermination() {
	<-s.signalsChannel
	log.Info("Termination starting")
	//err := streamer.Server.Shutdown(*streamer.Ctx)
	//if err != nil {
	//	log.Error("failed to terminate")
	//}
	log.Info("Termination complete")
}
