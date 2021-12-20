package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"www.seawise.com/common/log"
)

var home = os.Getenv("HOME") + "/seawise-video-stream/client/"
var deviceInfoFile = home + "core/saved/deviceInfo.conf"
var deviceConfigFile = home + "core/saved/deviceConfig.conf"

type DeviceInfo struct {
	Sn       string `json:"sn"`
	Owner    string `json:"owner"`
	Id       int    `json:"id"`
	Ip       *IP    `json:"ip"`
	Channels int    `json:"channels"`
}

type IPIFYResponse struct {
	Ip string `json:"ip"`
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
	Backend  string
	Stream   string
	Platform string
}

type Rule struct {
	Id        uint   `json:"id"`
	Recurring string `json:"recurring"`
	Start     uint   `json:"start"`
	Duration  uint   `json:"duration,string"`
	Type      string `json:"type"`
}

type IP struct {
	Local    string `json:"local"`
	External string `json:"external"`
}

func Produce() (*ConfigManager, error) {
	manager := ConfigManager{
		Backend: "http://" + Hosts.Backend + ":" + strconv.Itoa(Hosts.BackendPort),
		Stream:  "http://" + Hosts.Stream + ":" + strconv.Itoa(Hosts.StreamPort),
	}

	log.V5("REGISTERING DEVICE - " + manager.Backend)

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

	var body []byte

	resp, err := http.Get(cm.Backend + "/api/device/" + strconv.Itoa(cm.Info.Id))

	if err != nil || resp.StatusCode != 200 {
		log.Warn(fmt.Sprintf("failed to get Configuration from remote using local: %v", err))

		body, err = ioutil.ReadFile(deviceConfigFile)
		if err != nil {
			return fmt.Errorf("failed to read saved config EXITING: %v", err)
		}
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Invalid response from server EXITING: %v", err)
		}

		err = ioutil.WriteFile(deviceConfigFile, body, 0644)
		if err != nil {
			return fmt.Errorf("Failed to write config to local EXITING: %v", err)
		}

		defer resp.Body.Close()
	}

	err = json.Unmarshal(body, cm.Config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

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

	body, err := cm.post(cm.Backend+"/api/registration/update", postBody)
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

	log.V5(cm.Backend)
	var body []byte
	body, err = cm.post(cm.Backend+"/api/register", postBody)
	if err != nil {
		log.Warn(fmt.Sprintf("failed to register device no connectivity, looking to saved info: %v", err))
		body, err = ioutil.ReadFile(deviceInfoFile)
		if err != nil {
			return fmt.Errorf("failed to read info file and no connectivity: %v", err)
		}
	} else {
		err := os.WriteFile(deviceInfoFile, body, 0644)
		if err != nil {
			return fmt.Errorf("failed to write info EXITING: %v", err)
		}
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
	out, err := exec.Command("/bin/sh", "-c", "uname -m").Output()
	if err != nil {
		return fmt.Errorf("failed to identify platform: %v", err)
	}
	platform := strings.ReplaceAll(string(out), "\n", "")
	if platform == "aarch64" || platform == "armv7l" {
		cm.Platform = "pi"
	} else {
		cm.Platform = "other"
	}
	log.V5(fmt.Sprintf("MY PLATFORM IS - %v", cm.Platform))
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
	log.V5(fmt.Sprintf("SERIAL NUMBER IS - %v", sn))
	return sn, nil
}

func (cm *ConfigManager) getIp() (*IP, error) {
	if cm.Platform != "pi" {
		return &IP{"127.0.0.1", "127.0.0.1"}, nil
	}
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("failed to get IP: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return nil, fmt.Errorf("failed to get external IP: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ipJson := &IPIFYResponse{}
	err = json.Unmarshal(body, ipJson)

	return &IP{
		localAddr.IP.String(),
		ipJson.Ip,
	}, nil
}
