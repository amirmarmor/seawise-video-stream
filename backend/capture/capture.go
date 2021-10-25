package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"www.seawise.com/backend/core"
	"www.seawise.com/backend/log"
)

type Capture struct {
	counter     int
	run         bool
	manager     *core.ConfigManager
	Channels    []*Channel
	Action      chan *ShowRecord
	StopChannel chan string
	Errors      chan error
	lastUpdate  time.Time
	rules       []core.Rule
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
		lastUpdate:  time.Now(),
	}
}

func (c *Capture) Init() error {
	err := json.Unmarshal([]byte(c.manager.Config.Rules), &c.rules)
	if err != nil {
		return fmt.Errorf("failed to unmarshal rules: %v", err)
	}

	err = c.detectCameras()
	if err != nil {
		return err
	}

	err = c.manager.UpdateDeviceInfo(len(c.Channels))
	if err != nil {
		return fmt.Errorf("failed to update registration: %v", err)
	}

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
	for i, num := range vids {
		if i < 10 {
			if num >= c.manager.Config.Offset {
				channel := CreateChannel(num, c.rules, 30)
				err := channel.Init()
				if err != nil {
					continue
				} else {
					c.Channels = append(c.Channels, channel)
				}
			}
		}
	}

	return nil
}

func (c *Capture) Start() {
	for _, ch := range c.Channels {
		go ch.Start()
	}
}

func (c *Capture) stop(s string) {
	log.V5(fmt.Sprintf("capture - %s", s))
	c.run = false
}

func (c *Capture) restart() error {
	log.V5(fmt.Sprintf("restarting"))
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

func (c *Capture) capture() error {
	now := time.Now()
	if now.Sub(c.lastUpdate) > time.Second*10 {
		err := c.manager.GetConfig()
		if err == nil {
			err = json.Unmarshal([]byte(c.manager.Config.Rules), &c.rules)
			if err != nil {
				fmt.Println("failed to unmarshal config update: ", err)
			} else {
				log.V5("updated successful")
			}
		} else {
			fmt.Println("failed to get config update: ", err)
		}

		c.lastUpdate = now
	}

	for _, channel := range c.Channels {
		channel.Record = c.manager.Config.RecordNow
		channel.rules = c.rules
		go channel.Read()
		//if err != nil {
		//	return fmt.Errorf("capture failhttps://www.dynamsoft.com/codepool/opencv-webcam-golangapp-desktop-web.htmled: %v", err)
		//}
		//
		//if buf != nil {
		//	buf.Close()
		//}
	}
	return nil

}
