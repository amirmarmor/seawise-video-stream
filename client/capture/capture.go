package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"www.seawise.com/client/core"
	"www.seawise.com/common/log"
)

type Capture struct {
	counter    int
	manager    *core.ConfigManager
	Channels   []*Channel
	lastUpdate time.Time
	rules      []core.Rule
	timer      *time.Ticker
	stop       chan struct{}
	attempts   int
}

type ShowRecord struct {
	Type    string
	Channel int
}

func Create(config *core.ConfigManager, attempts int) *Capture {
	return &Capture{
		manager:    config,
		lastUpdate: time.Now(),
		timer:      time.NewTicker(10 * time.Second),
		attempts:   attempts,
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

	c.startStreamers()

	go c.updateConfig()
	return nil
}

func (c *Capture) startStreamers() {
	for _, channel := range c.Channels {
		channel.InitStreamer()
	}
}

func (c *Capture) updateConfig() {
	for {
		select {
		case <-c.timer.C:
			c.updateChannels()
		case <-c.stop:
			c.timer.Stop()
			return
		}
	}
}

func (c *Capture) updateChannels() {
	err := c.manager.GetConfig()
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to update configuration: %v", err))
		return
	}
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
			log.V5(vid.Name())
			vidNum, err := strconv.Atoi(re.FindAllString(vid.Name(), -1)[0])
			if err != nil {
				return fmt.Errorf("failed to convert video filename to int: %v", err)
			}
			vids = append(vids, vidNum)
		}
	}

	log.V5(fmt.Sprintf("Done checking vid - %v", vids))

	i := 0
	count := 0
	for i < c.attempts {
		log.V5(fmt.Sprintf("Attempting to start channel - %v / %v", i, c.attempts))
		c.Channels = make([]*Channel, 0)
		for _, num := range vids {
			if num >= c.manager.Config.Offset {
				channel := CreateChannel(c.manager.Config.Id, count, num, c.rules, c.manager.Config.Fps)
				err := channel.Init()
				if err != nil {
					continue
				} else {
					c.Channels = append(c.Channels, channel)
					count++
				}
			}
		}

		if len(c.Channels) > 0 {
			i = 99
		}

		i++
	}

	log.V5(fmt.Sprintf("Initiated all channels - %v", c.Channels))
	return nil
}

func (c *Capture) Start() {
	for _, ch := range c.Channels {
		go ch.Start()
	}
	log.V5(fmt.Sprintf("Started channels"))
}
