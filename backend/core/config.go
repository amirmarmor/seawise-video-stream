package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	externalip "github.com/glendc/go-external-ip"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"www.seawise.com/shrimps/backend/exposed"
)

type RegisterRequest struct {
	Sn    string `json:"sn"`
	Ip    string `json:"ip"`
	Owner string `json:"owner"`
}

type RegisterResponse struct {
	RegistrationId string `json:"registration_id"`
}

type GetConfigResponse struct {
	Ip      string `json:"ip"`
	Offset  int    `json:"offset,string"`
	Cleanup string `json:"cleanup"`
	Fps     int    `json:"fps,string"`
	Rules   string `json:"rules"`
}

type Configuration struct {
	Offset  int
	Cleanup bool
	Fps     int
	Rules   []Rule `json:"rules"`
}

type ConfigManager struct {
	Info   *RegisterRequest
	Id     string
	Config *Configuration
}

type Rule struct {
	Id        int64  `json:"id,string"`
	Recurring string `json:"recurring"`
	Start     int64  `json:"start,string"`
	Duration  int64  `json:"duration,string"`
	Type      string `json:"type"`
}

func Produce() (*ConfigManager, error) {
	InitFlags()

	manager := ConfigManager{}
	err := manager.Register()
	if err != nil {
		return nil, err
	}

	err = manager.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}

	return &manager, nil
}

func (cm *ConfigManager) GetConfig() error {
	resp, err := http.Get(exposed.ApiUrl + "/device/" + cm.Id)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := &GetConfigResponse{}

	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	config := &Configuration{}

	config.Offset = response.Offset
	config.Fps = response.Fps

	config.Cleanup, err = strconv.ParseBool(response.Cleanup)
	if err != nil {
		return err
	}

	config.Rules = make([]Rule, 0)
	err = json.Unmarshal([]byte(response.Rules), &config.Rules)
	if err != nil {
		return err
	}

	cm.Config = config

	return nil
}

func (cm *ConfigManager) Register() error {
	sn := cm.getSN()
	ip, err := cm.getIp()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	cm.Info = &RegisterRequest{
		Sn:    sn,
		Ip:    ip,
		Owner: "echo",
	}

	postBody, err := json.Marshal(cm.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal register requets: %v", err)
	}

	respBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(exposed.ApiUrl+"/register", "application/json", respBody)
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	response := &RegisterResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal register response: %v", err)
	}

	cm.Id = response.RegistrationId
	return nil
}

func (cm *ConfigManager) getSN() string {
	out, _ := exec.Command("/bin/sh", "-c", "sudo cat /sys/class/dmi/id/board_serial").Output()
	sn := strings.ReplaceAll(string(out), "\n", "")
	return sn
}

func (cm *ConfigManager) getIp() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", fmt.Errorf("failed to get ip: %v", err)
	}
	return ip.String(), nil
}
