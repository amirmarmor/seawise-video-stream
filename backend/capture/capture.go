package capture

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"www.seawise.com/shrimps/backend/core"
	"www.seawise.com/shrimps/backend/log"
	"www.seawise.com/shrimps/backend/streamer"
)

type Capture struct {
	counter     int
	run         bool
	manager     *core.ConfigManager
	Channels    []*Channel
	Recording   map[string]time.Time
	Action      chan *ShowRecord
	StopChannel chan string
	Errors      chan error
	stream  		*streamer.Streamer
}

type ShowRecord struct {
	Type    string
	Channel int
}

func Create(config *core.ConfigManager) *Capture {
	return &Capture{
		manager:     config,
		run:         true,
		Action:      make(chan *ShowRecord, 0),
		StopChannel: make(chan string, 0),
	}
}

func (c *Capture) Init() error {
	err := c.detectCameras()
	if err != nil {
		return err
	}

	c.stream = streamer.Create()
	for i := 0; i < len(c.Channels); i++ {
		c.stream.Produce(i, c.Channels[i].Stream)
	}
	go c.stream.Start()
	return nil
}

func (c *Capture) detectCameras() error {
	devs, err := os.ReadDir("/dev")
	if err != nil {
		return fmt.Errorf("failed to read dir /dev: %v", err)
	}
	re := regexp.MustCompile("[0-9]+")

	var vids []int
	for _, vid := range devs {
		if strings.Contains(vid.Name(), "video") {
			vidNum, err := strconv.Atoi(re.FindAllString(vid.Name(), -1)[0])
			if err != nil {
				return fmt.Errorf("failed to convert video filename to int: %v", err)
			}
			vids = append(vids, vidNum)
		}
	}

	c.Channels = make([]*Channel, 0)
	for _, num := range vids {
		if num >= c.manager.Config.Offset {
			channel := CreateChannel(num, c.manager.Config.Rules)
			err := channel.Init()
			if err != nil {
				continue
			} else {
				c.Channels = append(c.Channels, channel)
			}
		}
	}
	return nil
}

func (c *Capture) Start() {
	for c.run {
		select {
		case action := <-c.Action:
			c.update(action)
		case s := <-c.StopChannel:
			c.stop(s)
		default:
			c.capture()
		}
	}
	c.StopChannel <- "restarting"
}

func (c *Capture) Update() {
	c.StopChannel <- "stopping"
	for !c.run {
		select {
		case s := <-c.StopChannel:
			c.restart(s)
		default:
			continue
		}
	}
}

func (c *Capture) stop(s string) {
	log.V5(fmt.Sprintf("capture - %s", s))
	c.run = false
}

func (c *Capture) restart(s string) error {
	log.V5(fmt.Sprintf("capture - %s", s))
	c.run = true

	for i := 0; i < len(c.Channels); i++ {
		c.Channels[i].close()
	}

	c.Channels = make([]*Channel, 0)

	err := c.Init()
	if err != nil {
		return err
	}

	go c.Start()
	return nil
}

func (c *Capture) update(action *ShowRecord) error {
	if action.Type == "record" {
		c.Channels[action.Channel].Record = !c.Channels[action.Channel].Record
	}

	if action.Type == "config" {
		err := c.manager.GetConfig()
		for _, channel := range c.Channels {
			channel.rules = c.manager.Config.Rules
		}
		if err != nil {
			return err
		}
	}

	err := c.capture()
	if err != nil {
		return err
	}
	return nil
}

func (c *Capture) capture() error {
	for _, channel := range c.Channels {
		err := channel.Read()
		if err != nil {
			return fmt.Errorf("capture failed: %v", err)
		}
	}
	return nil
}
