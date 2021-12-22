package capture

import (
	"bytes"
	"fmt"
	"github.com/hybridgroup/mjpeg"
	"gocv.io/x/gocv"
	"image/jpeg"
	"os"
	"reflect"
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
	cap            *gocv.VideoCapture
	image          gocv.Mat
	writer         *gocv.VideoWriter
	Queue          chan []byte
	Record         bool
	Recording      bool
	Rules          []core.Rule
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
		Rules:   rules,
		created: time.Now(),
		fps:     fps,
		Queue:   make(chan []byte),
	}

	return channel
}

func (c *Channel) Init() error {
	vc, err := gocv.OpenVideoCaptureWithAPI(c.name, gocv.VideoCaptureV4L2)
	if err != nil {
		return fmt.Errorf("Init failed to capture video %v: ", err)
	}
	vc.Set(gocv.VideoCaptureFOURCC, vc.ToCodec("mjpg"))
	vc.Set(gocv.VideoCaptureFPS, float64(c.fps))
	vc.Set(gocv.VideoCaptureFrameWidth, 1920)
	vc.Set(gocv.VideoCaptureFrameHeight, 1080)
	vc.Set(gocv.VideoCaptureBufferSize, 5)
	img := gocv.NewMat()

	ok := vc.Read(&img)
	if !ok {
		return fmt.Errorf("Init failed to read")
	}

	c.cap = vc
	c.image = img
	c.init = true
	c.run = true

	return nil
}

func (c *Channel) InitStreamer() {
	c.streamer = CreateStreamer(c.port, c.Queue)
}

func (c *Channel) close() error {
	err := c.cap.Close()
	if err != nil {
		return fmt.Errorf("failed to close capture: %v", err)
	}
	err = c.image.Close()
	if err != nil {
		return fmt.Errorf("failed to close image: %v", err)
	}

	err = c.writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
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
	ok := c.cap.Read(&c.image)
	if !ok {
		return fmt.Errorf("read encountered channel closed %v\n", c.name)
	}

	if c.image.Empty() {
		return fmt.Errorf("Empty Image")
	}

	return nil
}

func (c *Channel) Read() {
	imageRecord := c.checkImageRules()
	videoRecord := c.checkVideoRules()

	if !c.init {
		err := c.Init()
		if err != nil {
			log.Warn(fmt.Sprintf("read init failed to close: %v", err))
			c.run = false
		}
	}

	if imageRecord {
		now := time.Now()
		saveFileName := c.path + "/" + now.Format("2006-01-02--15-04-05") + "-image.jpg"

		err := c.getImage()
		if err != nil {
			log.Warn(fmt.Sprintf("failed to read image: %v", err))
			return
		}

		ok := gocv.IMWrite(saveFileName, c.image)
		if !ok {
			log.Warn(fmt.Sprintf("failed to write image"))
		}
	}

	if videoRecord {
		err := c.getImage()
		if err != nil {
			log.Warn(fmt.Sprintf("failed to read image: %v", err))
			return
		}

		err = c.doRecord()
		if err != nil {
			log.Warn(fmt.Sprintf("fauled to record: %v", err))
		}
	}

	if true {
		err := c.getImage()
		if err != nil {
			log.Warn(fmt.Sprintf("failed to read image: %v", err))
			return
		}
		c.Queue <- c.encodeImage()
		gocv.WaitKey(1)
	}
}

//func (c *Channel) doStream() error {
//	if c.Recording {
//		log.V5("STOP RECORD")
//	}
//
//	c.Recording = false
//
//	buffer, err := gocv.IMEncode(".jpg", c.image)
//	if err != nil {
//		return fmt.Errorf("read failed to encode: %v", err)
//	}
//
//	c.Stream.UpdateJPEG(buffer.GetBytes())
//	if err != nil {
//		return fmt.Errorf("failed to update stream in read: %v", err)
//	}
//
//	buffer.Close()
//	return nil
//}

func (c *Channel) doRecord() error {
	if !c.Recording {
		err := c.createVWriter()
		if err != nil {
			return fmt.Errorf("faield to create writer: %v", err)
		}
	}

	err := c.writer.Write(c.image)
	if err != nil {
		return fmt.Errorf("read failed to write: %v", err)
	}

	return nil
}

func (c *Channel) createVWriter() error {
	log.V5("START RECORD")
	c.Recording = true
	now := time.Now()
	path, err := c.createSavePath()
	if err != nil {
		return fmt.Errorf("failed to create path: %v", err)
	}

	saveFileName := path + "/" + now.Format("2006-01-02--15-04-05") + ".avi"

	c.writer, err = gocv.VideoWriterFile(saveFileName, "MJPG", float64(c.fps), c.image.Cols(), c.image.Rows(), true)
	if err != nil {
		return fmt.Errorf("failed to create writer", err)
	}

	return nil
}

func (c *Channel) createSavePath() (string, error) {
	_, err := os.Stat("videos")

	if os.IsNotExist(err) {
		log.V5("videos directory doesnt exist. creating it now!")
		err := os.Mkdir("videos", 0777)
		if err != nil {
			log.Error("couldnt create images directory", err)
			return "", err
		}
	}

	path := fmt.Sprintf("videos/channel-%v", c.name)
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		log.V5("creating file direcotry!")
		err = os.Mkdir(path, 0777)
		if err != nil {
			log.Error("couldnt create images directory", err)
			return "", err
		}
	}

	return path, nil
}

func (c *Channel) checkImageRules() bool {
	now := time.Now()
	for _, rule := range c.Rules {
		if rule.Type != "image" {
			return false
		}

		if rule.Duration == 0 {
			return false
		}

		var t float64
		if rule.Recurring == "Second" {
			t = time.Minute.Seconds()
		} else if rule.Recurring == "Minute" {
			t = time.Hour.Seconds()
		} else {
			t = time.Hour.Seconds() * 24
		}

		interval := time.Duration(t / float64(rule.Duration))

		if now.Sub(c.lastImage) >= interval {
			c.lastImage = now
			return true
		}
	}
	return false
}

func (c *Channel) checkVideoRules() bool {
	now := time.Now()
	if c.Record {
		return true
	}

	if len(c.Rules) == 0 {
		return false
	}

	for _, rule := range c.Rules {

		if rule.Type != "video" {
			return false
		}

		if rule.Duration == 0 {
			return false
		}

		bar := GetTimeField(rule.Recurring, now)

		if rule.Start == uint(bar) {
			c.startRecording = now
		}

		if c.startRecording.IsZero() || now.Sub(c.startRecording) > time.Second*time.Duration(rule.Duration) {
			c.startRecording = time.Time{}
			return false
		}

		return true
	}
	return false
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

func GetTimeField(s string, now time.Time) int64 {
	r := reflect.ValueOf(now).MethodByName(s)
	f := r.Call(nil)
	return int64(f[0].Interface().(int))
}
