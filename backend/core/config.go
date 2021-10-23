package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"www.seawise.com/shrimps/backend/log"
)

type DeviceInfo struct {
	Sn       string `json:"sn"`
	Owner    string `json:"owner"`
	Id       int    `json:"id"`
	Ip       string `json:"ip"`
	Channels int    `json:"channels"`
}

type RegisterResponse struct {
	RegistrationId int `json:"id"`
}

type MessageResponse struct {
	Msg string `json:"msg"`
}

type Configuration struct {
	Id        int    `json:"id"`
	Offset    int    `json:"offset"`
	Cleanup   bool   `json:"cleanup"`
	Fps       int    `json:"fps"`
	RecordNow bool   `json:"record"`
	Rules     string `json:"rules"`
}

type ConfigManager struct {
	Info     *DeviceInfo
	Config   *Configuration
	Platform string
}

type Rule struct {
	Id        int64  `json:"id"`
	Recurring string `json:"recurring"`
	Start     int64  `json:"start,string"`
	Duration  int64  `json:"duration,string"`
	Type      string `json:"type"`
}

func Produce() (*ConfigManager, error) {
	manager := ConfigManager{}

	log.V5("REGISTERING DEVICE - " + Api.Host)

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
	cm.Config = &Configuration{
		Offset: 0,
		Id:     cm.Info.Id,
	}

	resp, err := http.Get("http://" + Api.Host + "/api/device/" + strconv.Itoa(cm.Info.Id))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, cm.Config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal: %v", err)
		}
	} else {
		err = cm.SetConfig()
		if err != nil {
			return fmt.Errorf("failed to set default config: %v", err)
		}
	}

	defer resp.Body.Close()
	return nil
}

func (cm *ConfigManager) SetConfig() error {

	cm.Config.RecordNow = false
	cm.Config.Cleanup = true
	cm.Config.Fps = 30
	cm.Config.Rules = "[]"

	return nil
}

func (cm *ConfigManager) UpdateDeviceInfo(channels int) error {
	ip, err := cm.getIp()
	if err != nil {
		return fmt.Errorf("failed to update registration: %v", err)
	}

	cm.Info.Channels = channels
	cm.Info.Ip = ip

	postBody, err := json.Marshal(cm.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal register requets: %v", err)
	}

	body, err := cm.post("http://"+Api.Host+"/api/registration/update", postBody)
	if err != nil {
		return fmt.Errorf("failed to update registration: %v", err)
	}

	response := &MessageResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("failed to update registration: %v", err)
	}

	log.V5(response.Msg)
	return nil
}

func (cm *ConfigManager) Register() error {
	err := cm.getPlatform()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	ip, err := cm.getIp()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	sn, err := cm.getSN()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	cm.Info = &DeviceInfo{
		Sn:    sn,
		Ip:    ip,
		Owner: "echo",
	}

	postBody, err := json.Marshal(cm.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal register requets: %v", err)
	}

	apiUrl := "http://" + Api.Host + "/api/register"
	log.V5(apiUrl)
	body, err := cm.post(apiUrl, postBody)
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	response := &RegisterResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal register response: %v", err)
	}

	cm.Info.Id = response.RegistrationId

	return nil
}

func (cm *ConfigManager) post(url string, postBody []byte) ([]byte, error) {
	respBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to post: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to post: %v", err)
	}

	return body, nil
}

func (cm *ConfigManager) getPlatform() error {
	log.V5("HERE")
	out, err := exec.Command("/bin/sh", "-c", "uname -m").Output()
	if err != nil {
		return fmt.Errorf("failed to identify platform: %v", err)
	}
	platform := strings.ReplaceAll(string(out), "\n", "")
	if strings.Contains(platform, "arm"){
		cm.Platform = "pi"
	} else {
		cm.Platform = "other"
	}
	log.V5("MY PLATFORM IS", cm.Platform)
	return nil
}


func (cm *ConfigManager) getSN() (string, error) {
	log.V5("GETTING SERIAL NUMBER")
	var out *exec.Cmd
	if cm.Platform == "pi" {
		out = exec.Command("/bin/sh", "-c", "sudo cat /proc/cpuinfo | grep Serial | cut -d ' ' -f 2")
	} else {
		out = exec.Command("/bin/sh", "-c", "sudo cat /sys/class/dmi/id/board_serial")
	}
	res, err := out.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get S/N: %v", err)
	}
	sn := strings.ReplaceAll(string(res), "\n", "")
	log.V5("sn", sn)
	return sn, nil
}

func (cm *ConfigManager) getIp() (string, error) {
	if cm.Platform != "pi" {
		return "127.0.0.1", nil
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to get IP: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
