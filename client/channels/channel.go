package channels

import (
	"bytes"
	"fmt"
	"gocv.io/x/gocv"
	"image/jpeg"
	"time"
	"www.seawise.com/client/core"
	"www.seawise.com/common/log"
)

type Channel struct {
	fps         int
	name        int
	init        bool
	capture     *gocv.VideoCapture
	image       gocv.Mat
	Queue       chan []byte
	StopChannel chan string
	streamer    *Streamer
}

type Recording struct {
	isRecording bool
	startTime   time.Time
}

func CreateChannel(channelName int) *Channel {
	channel := &Channel{
		name:        channelName,
		Queue:       make(chan []byte),
		StopChannel: make(chan string),
	}

	return channel
}

func (c *Channel) Init(detectOnly bool) error {
	vc, err := gocv.OpenVideoCapture(c.name)
	if err != nil {
		return fmt.Errorf("Init failed to capture video %v: ", err)
	}

	vc.Set(gocv.VideoCaptureFrameWidth, 1920)
	vc.Set(gocv.VideoCaptureFrameHeight, 1080)
	vc.Set(gocv.VideoCaptureBufferSize, 5)
	img := gocv.NewMat()

	ok := vc.Read(&img)
	if !ok {
		return fmt.Errorf("Init failed to read")
	}

	c.init = true

	if detectOnly {
		err := vc.Close()
		if err != nil {
			return fmt.Errorf("failed to close vc in camera detection phase: %v", err)
		}

		err = img.Close()
		if err != nil {
			return fmt.Errorf("failed to close mat in camera detection phase: %v", err)
		}

		return nil
	}

	c.capture = vc
	c.image = img

	return nil
}

func (c *Channel) Ready(fps int, id int, count int) {
	port := core.Hosts.StreamPort + (id * 10) + count
	c.fps = fps

	if c.streamer == nil {
		c.streamer = CreateStreamer(port, c.Queue)
	}

	err := c.Init(false)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to init: %v", err))
	}
}

func (c *Channel) close() {
	err := c.capture.Close()
	if err != nil {
		log.Warn(fmt.Sprintf("failed to close capture: %v", err))
	}
	err = c.image.Close()
	if err != nil {
		log.Warn(fmt.Sprintf("failed to close image: %v", err))
	}

	c.init = false
	log.V5("stopped....")
}

func (c *Channel) Start() {
	for c.init {
		select {
		case <-c.StopChannel:
			c.close()
		default:
			c.Read()
		}
	}
}

func (c *Channel) getImage() error {
	ok := c.capture.Read(&c.image)
	if !ok {
		return fmt.Errorf("read encountered channel closed %v\n", c.name)
	}

	if c.image.Empty() {
		return fmt.Errorf("Empty Image")
	}

	return nil
}

func (c *Channel) Read() {
	if !c.init {
		err := c.Init(false)
		if err != nil {
			log.Warn(fmt.Sprintf("read init failed to close: %v", err))
		}
	}

	c.capture.Set(gocv.VideoCaptureFPS, float64(c.fps))

	err := c.getImage()
	if err != nil {
		log.Warn(fmt.Sprintf("failed to read image: %v", err))
		return
	}

	c.Queue <- c.encodeImage()
	gocv.WaitKey(1)
}

func (c *Channel) encodeImage() []byte {
	const jpegQuality = 50

	jpegOption := &jpeg.Options{Quality: jpegQuality}

	image, err := c.image.ToImage()
	if err != nil {
		return nil
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, image, jpegOption)
	if err != nil {
		return nil
	}

	return buf.Bytes()
}