package capture

import (
	"fmt"
	"github.com/mattn/go-mjpeg"
	"gocv.io/x/gocv"
	"os"
	"reflect"
	"time"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/log"
)

const interval = 50 * time.Millisecond

type Channel struct {
	created        time.Time
	cleanup        bool
	name           int
	init           bool
	cap            *gocv.VideoCapture
	image          gocv.Mat
	writer         *gocv.VideoWriter
	Record         bool
	Recording      bool
	rules          []core.Rule
	path           string
	Stream         *mjpeg.Stream
	fps            int
	lastImage      time.Time
	startRecording time.Time
}

type Recording struct {
	isRecording bool
	startTime   time.Time
}

func CreateChannel(channel int, rules []core.Rule, fps int) *Channel {
	return &Channel{
		name:    channel,
		Stream:  mjpeg.NewStreamWithInterval(interval),
		rules:   rules,
		created: time.Now(),
		fps:     fps,
	}
}

func (c *Channel) Init() error {
	vc, err := gocv.OpenVideoCapture(c.name)
	if err != nil {
		return fmt.Errorf("Init failed to capture video %v: ", err)
	}
	vc.Set(gocv.VideoCaptureFOURCC, vc.ToCodec("mjpg"))
	vc.Set(gocv.VideoCaptureFPS, float64(c.fps))
	vc.Set(gocv.VideoCaptureFrameWidth, 1920)
	vc.Set(gocv.VideoCaptureFrameHeight, 1080)
	img := gocv.NewMat()

	ok := vc.Read(&img)
	if !ok {
		return fmt.Errorf("Init failed to read")
	}

	c.cap = vc
	c.image = img
	c.init = true

	return nil
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

func (c *Channel) Read() (*gocv.NativeByteBuffer, error) {
	imageRecord := c.checkImageRules()
	videoRecord := c.checkVideoRules()

	if !c.init {
		err := c.Init()
		if err != nil {
			return nil, fmt.Errorf("read init failed to close: %v", err)
		}
	}

	ok := c.cap.Read(&c.image)
	if !ok {
		return nil, fmt.Errorf("read encountered channel closed %v\n", c.name)
	}

	if c.image.Empty() {
		return nil, nil
	}

	if imageRecord {
		now := time.Now()
		saveFileName := c.path + "/" + now.Format("2006-01-02--15-04-05") + "-image.jpg"
		ok := gocv.IMWrite(saveFileName, c.image)
		if !ok {
			return nil, fmt.Errorf("read failed to write image")
		}
	}

	if videoRecord {
		if !c.Recording {
			c.Recording = true
			now := time.Now()
			path, err := c.createSavePath()
			if err != nil {
				return nil, fmt.Errorf("failed to create path: %v", err)
			}

			saveFileName := path + "/" + now.Format("2006-01-02--15-04-05") + ".avi"

			c.writer, err = gocv.VideoWriterFile(saveFileName, "MJPG", float64(c.fps), c.image.Cols(), c.image.Rows(), true)
			if err != nil {
				return nil, fmt.Errorf("failed to create writer", err)
			}
		}

		err := c.writer.Write(c.image)
		if err != nil {
			return nil, fmt.Errorf("read failed to write: %v", err)
		}
	} else {
		c.Recording = false
	}

	quality := 50
	buffer, err := gocv.IMEncodeWithParams(".jpg", c.image, []int{gocv.IMWriteJpegQuality, quality})
	if err != nil {
		return nil, fmt.Errorf("read failed to encode: %v", err)
	}

	err = c.Stream.Update(buffer.GetBytes())
	if err != nil {
		return nil, fmt.Errorf("failed to update stream in read: %v", err)
	}

	return buffer, nil
}

func (c *Channel) createSavePath() (string, error) {
	now := time.Now()
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

	if c.cleanup && now.Sub(c.created) >= time.Hour*24 {
		err := os.RemoveAll(path)
		if err != nil {
			log.Error("couldnt remove folder", path)
		}
		c.created = now
	}

	return path, nil
}

func (c *Channel) checkImageRules() bool {
	now := time.Now()
	for _, rule := range c.rules {
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

	if len(c.rules) == 0 {
		return false
	}

	for _, rule := range c.rules {

		if rule.Type != "video" {
			return false
		}

		if rule.Duration == 0 {
			return false
		}

		bar := GetTimeField(rule.Recurring, now)

		if rule.Start == bar {
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

func GetTimeField(s string, now time.Time) int64 {
	r := reflect.ValueOf(now).MethodByName(s)
	f := r.Call(nil)
	return int64(f[0].Interface().(int))
}
