package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
	"www.seawise.com/shrimps/backend/capture"
	"www.seawise.com/shrimps/backend/core"
)

type Executor struct {
	Manager *core.ConfigManager
	Capt    *capture.Capture
	Wait    bool
}

type ActionRequest struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

type SetResponse struct {
	Message string
}

type ConfigResponse struct {
	Offset   int         `json:"offset"`
	Cleanup  bool        `json:"cleanup"`
	Rules    []core.Rule `json:"rules"`
	Show     []int       `json:"show"`
	Record   []int       `json:"record"`
	Channels int         `json:"channels"`
}

func Produce(configManager *core.ConfigManager, capt *capture.Capture) *Executor {
	executor := Executor{
		Manager: configManager,
		Capt:    capt,
	}
	return &executor
}

func (executor *Executor) prepareConfigResponse() *ConfigResponse {
	show := make([]int, 0)
	record := make([]int, 0)
	for i, channel := range executor.Capt.Channels {
		if channel.Show {
			show = append(show, i)
		}
		if channel.Record {
			record = append(record, i)
		}
	}
	return &ConfigResponse{
		executor.Manager.Config.Offset,
		executor.Manager.Config.Cleanup,
		executor.Manager.Config.Rules,

		show,
		record,
		len(executor.Capt.Channels),
	}
}

func (executor *Executor) GetConfig(c echo.Context) error {
	response := executor.prepareConfigResponse()
	return c.JSON(http.StatusOK, response)
}

func (executor *Executor) SetConfig(c echo.Context) error {
	config := &core.Configuration{}

	err := c.Bind(config)
	if err != nil {
		return fmt.Errorf("failed to bind: %v", err)
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal: %v", err)
	}

	temp := string(configJson)
	err = executor.Manager.SetConfig(temp)
	if err != nil {
		return fmt.Errorf("failed to set configuration: %v", err)
	}

	action := &capture.ShowRecord{
		Type:    "config",
		Channel: 0,
	}

	executor.Capt.Action <- action

	response := executor.prepareConfigResponse()
	//go executor.Capt.Update() //TODO: fix update when offset changes - restart issue

	return c.JSON(http.StatusOK, response)
}

func (executor *Executor) DoAction(c echo.Context) error {
	actionRequest := &ActionRequest{}
	err := c.Bind(actionRequest)
	if err != nil {
		return fmt.Errorf("failed to bind: %v", err)
	}

	channel, err := strconv.Atoi(actionRequest.Channel)
	if err != nil {
		return fmt.Errorf("failed to convert channel: %v", err)
	}

	action := &capture.ShowRecord{
		Type:    actionRequest.Type,
		Channel: channel,
	}

	executor.Capt.Action <- action

	time.Sleep(time.Millisecond * 10)
	response := executor.prepareConfigResponse()
	return c.JSON(http.StatusOK, response)
}
