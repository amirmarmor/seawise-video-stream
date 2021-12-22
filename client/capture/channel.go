package capture

import (
	"bytes"
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"image/jpeg"
	"time"
	"www.seawise.com/client/core"
	"www.seawise.com/common/log"
)

type Channel struct {
	run            bool
	created        time.Time
	name           int
	port           int
	init           bool
	capture        *gocv.VideoCapture
	image          gocv.Mat
	Queue          chan []byte
	path           string
	Stream         *mjpeg.Stream
	fps            int
	lastImage      time.Time
	startRecording time.Time
	StopChannel    chan string
	streamer       *Streamer
}

type Recording struct {
	isRecording bool
	startTime   time.Time
}

func CreateChannel(device int, count int, channelName int, rules []core.Rule, fps int) *Channel {
	channel := &Channel{
		name:    channelName,
		port:    core.Hosts.StreamPort + (device * 10) + count,
		Stream:  mjpeg.NewStream(),
		created: time.Now(),
		fps:     fps,
		Queue:   make(chan []byte),
	}

	return channel
}

func (c *Channel) Init() error {
	vc, err := gocv.OpenVideoCapture(c.name)
	if err != nil {
		return fmt.Errorf("Init failed to capture video %v: ", err)
	}
	vc.Set(gocv.VideoCaptureFPS, float64(c.fps))
	vc.Set(gocv.VideoCaptureFrameWidth, 1920)
	vc.Set(gocv.VideoCaptureFrameHeight, 1080)
	vc.Set(gocv.VideoCaptureBufferSize, 5)
	img := gocv.NewMat()

	ok := vc.Read(&img)
	if !ok {
		return fmt.Errorf("Init failed to read")
	}

	c.capture = vc
	c.image = img
	c.init = true
	c.run = true

	return nil
}

func (c *Channel) InitStreamer() {
	c.streamer = CreateStreamer(c.port, c.Queue)
}

func (c *Channel) close() error {
	err := c.capture.Close()
	if err != nil {
		return fmt.Errorf("failed to close capture: %v", err)
	}
	err = c.image.Close()
	if err != nil {
		return fmt.Errorf("failed to close image: %v", err)
	}

	c.init = false
	return nil
}

func (c *Channel) Start() {
	for c.run {
		select {
		case <-c.StopChannel:
			c.close()
		default:
			c.Read()
		}
	}
	c.StopChannel <- "restarting"
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
		err := c.Init()
		if err != nil {
			log.Warn(fmt.Sprintf("read init failed to close: %v", err))
			c.run = false
		}
	}

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
