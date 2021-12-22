package server

import (
	"bytes"
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"net"
	"net/http"
	"sync"
	"www.seawise.com/backend/core"
	"www.seawise.com/common/log"
)

type Server struct {
	TCPListener             net.Listener
	TCPListenerMutex        sync.Mutex
	Frame                   *bytes.Buffer
	FrameMutex              sync.RWMutex
	timeStampPacketSize     uint
	contentLengthPacketSize uint
	Stream                  *mjpeg.Stream
	Port                    int
}

func Create(devices *core.Devices) ([]*Server, error) {
	servers := make([]*Server, 0)
	for _, device := range devices.List {
		for ch := 0; ch < device.Channels; ch++ {
			port := core.Config.Port + (device.Id * 10) + ch
			server, err := NewServer(port)
			if err != nil {
				return nil, fmt.Errorf("failed to create listener - %v", err)
			}

			go server.Run()

			servers = append(servers, server)
		}
	}

	return servers, nil
}

func NewServer(port int) (*Server, error) {
	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})

	if err != nil {
		return nil, fmt.Errorf("generate tcp server failed! - %v", err)
	}

	buf := new(bytes.Buffer)

	server := &Server{
		TCPListener:             tcpListener,
		TCPListenerMutex:        sync.Mutex{},
		Frame:                   buf,
		FrameMutex:              sync.RWMutex{},
		timeStampPacketSize:     8,
		contentLengthPacketSize: 8,
		Stream:                  mjpeg.NewStream(),
		Port:                    port,
	}
	log.V5(fmt.Sprintf("Listening on 127.0.0.1:%v", port))
	return server, nil
}

func (s *Server) Run() {
	defer func(tcpListener net.Listener) {
		err := tcpListener.Close()
		if err != nil {
			panic(err)
		}
	}(s.TCPListener)

	go s.runListener()

	for {
		s.FrameMutex.RLock()
		s.Stream.UpdateJPEG(s.Frame.Bytes())
		s.FrameMutex.RUnlock()
	}
}

func (s *Server) HandleOutbound(w http.ResponseWriter, r *http.Request) {
	log.V5("CONNN")
	s.Stream.ServeHTTP(w, r)
}

//if imageRecord {
//now := time.Now()
//saveFileName := c.path + "/" + now.Format("2006-01-02--15-04-05") + "-image.jpg"
//
//err := c.getImage()
//if err != nil {
//log.Warn(fmt.Sprintf("failed to read image: %v", err))
//return
//}
//
//ok := gocv.IMWrite(saveFileName, c.image)
//if !ok {
//log.Warn(fmt.Sprintf("failed to write image"))
//}
//}

//if videoRecord {
//err := c.getImage()
//if err != nil {
//log.Warn(fmt.Sprintf("failed to read image: %v", err))
//return
//}
//
//err = c.doRecord()
//if err != nil {
//log.Warn(fmt.Sprintf("fauled to record: %v", err))
//}
//}
//

//func (c *Channel) doRecord() error {
//	if !c.Recording {
//		err := c.createVWriter()
//		if err != nil {
//			return fmt.Errorf("faield to create writer: %v", err)
//		}
//	}
//
//	err := c.writer.Write(c.image)
//	if err != nil {
//		return fmt.Errorf("read failed to write: %v", err)
//	}
//
//	return nil
//}
//
//func (c *Channel) createVWriter() error {
//	log.V5("START RECORD")
//	c.Recording = true
//	now := time.Now()
//	path, err := c.createSavePath()
//	if err != nil {
//		return fmt.Errorf("failed to create path: %v", err)
//	}
//
//	saveFileName := path + "/" + now.Format("2006-01-02--15-04-05") + ".avi"
//
//	c.writer, err = gocv.VideoWriterFile(saveFileName, "MJPG", float64(c.fps), c.image.Cols(), c.image.Rows(), true)
//	if err != nil {
//		return fmt.Errorf("failed to create writer", err)
//	}
//
//	return nil
//}
//
//func (c *Channel) createSavePath() (string, error) {
//	_, err := os.Stat("videos")
//
//	if os.IsNotExist(err) {
//		log.V5("videos directory doesnt exist. creating it now!")
//		err := os.Mkdir("videos", 0777)
//		if err != nil {
//			log.Error("couldnt create images directory", err)
//			return "", err
//		}
//	}
//
//	path := fmt.Sprintf("videos/channel-%v", c.name)
//	_, err = os.Stat(path)
//
//	if os.IsNotExist(err) {
//		log.V5("creating file direcotry!")
//		err = os.Mkdir(path, 0777)
//		if err != nil {
//			log.Error("couldnt create images directory", err)
//			return "", err
//		}
//	}
//
//	return path, nil
//}
//
//func (c *Channel) checkImageRules() bool {
//	now := time.Now()
//	for _, rule := range c.Rules {
//		if rule.Type != "image" {
//			return false
//		}
//
//		if rule.Duration == 0 {
//			return false
//		}
//
//		var t float64
//		if rule.Recurring == "Second" {
//			t = time.Minute.Seconds()
//		} else if rule.Recurring == "Minute" {
//			t = time.Hour.Seconds()
//		} else {
//			t = time.Hour.Seconds() * 24
//		}
//
//		interval := time.Duration(t / float64(rule.Duration))
//
//		if now.Sub(c.lastImage) >= interval {
//			c.lastImage = now
//			return true
//		}
//	}
//	return false
//}
//
//func (c *Channel) checkVideoRules() bool {
//	now := time.Now()
//	if c.Record {
//		return true
//	}
//
//	if len(c.Rules) == 0 {
//		return false
//	}
//
//	for _, rule := range c.Rules {
//
//		if rule.Type != "video" {
//			return false
//		}
//
//		if rule.Duration == 0 {
//			return false
//		}
//
//		bar := GetTimeField(rule.Recurring, now)
//
//		if rule.Start == uint(bar) {
//			c.startRecording = now
//		}
//
//		if c.startRecording.IsZero() || now.Sub(c.startRecording) > time.Second*time.Duration(rule.Duration) {
//			c.startRecording = time.Time{}
//			return false
//		}
//
//		return true
//	}
//	return false
//}

//func GetTimeField(s string, now time.Time) int64 {
//	r := reflect.ValueOf(now).MethodByName(s)
//	f := r.Call(nil)
//	return int64(f[0].Interface().(int))
//}
