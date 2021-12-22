package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"www.seawise.com/client/channels"
	"www.seawise.com/client/core"
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
	Id      int  `json:"id"`
	Offset  int  `json:"offset"`
	Cleanup bool `json:"cleanup"`
	Fps     int  `json:"fps"`
}

type IP struct {
	Local    string `json:"local"`
	External string `json:"external"`
}

type Server struct {
	Backend       string
	DeviceInfo    *DeviceInfo
	Configuration *Configuration
	Router        *mux.Router
	Channels      *channels.Channels
	Platform      string
}

func Produce(channels *channels.Channels) (*Server, error) {

	server := &Server{
		Backend:  "http://" + core.Hosts.Backend + ":" + strconv.Itoa(core.Hosts.BackendPort),
		Channels: channels,
	}

	log.V5("REGISTERING DEVICE - " + server.Backend)

	err := server.Register(len(channels.Array))
	if err != nil {
		return nil, err
	}

	err = server.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}

	server.Router = mux.NewRouter()
	server.Router.HandleFunc("/{action}", server.Handler)
	server.Router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.Warn(fmt.Sprintf("unable to write response on health check: %v", err))
		}
	})
	return server, nil
}

func (s *Server) GetConfig() error {
	s.Configuration = &Configuration{
		Offset: 0,
		Id:     s.DeviceInfo.Id,
	}

	var body []byte

	resp, err := http.Get(s.Backend + "/api/device/" + strconv.Itoa(s.DeviceInfo.Id))

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

	err = json.Unmarshal(body, s.Configuration)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	return nil
}

func (s *Server) Register(channels int) error {

	url := "http://" + core.Hosts.Backend + ":" + strconv.Itoa(core.Hosts.BackendPort) + "/api/register"
	err := s.getPlatform()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	ip, err := s.getIp()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	sn, err := s.getSN()
	if err != nil {
		return fmt.Errorf("failed to register: %v", err)
	}

	s.DeviceInfo = &DeviceInfo{
		Sn:       sn,
		Ip:       ip,
		Owner:    "echo",
		Channels: channels,
	}

	postBody, err := json.Marshal(s.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal register requets: %v", err)
	}

	log.V5(url)
	var body []byte
	body, err = s.post(url, postBody)
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

	s.DeviceInfo.Id = response.RegistrationId

	return nil
}

func (s *Server) post(url string, postBody []byte) ([]byte, error) {
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

func (s *Server) getPlatform() error {
	out, err := exec.Command("/bin/sh", "-c", "uname -m").Output()
	if err != nil {
		return fmt.Errorf("failed to identify platform: %v", err)
	}
	platform := strings.ReplaceAll(string(out), "\n", "")
	if platform == "aarch64" || platform == "armv7l" {
		s.Platform = "pi"
	} else {
		s.Platform = "other"
	}
	log.V5(fmt.Sprintf("MY PLATFORM IS - %v", s.Platform))
	return nil
}

func (s *Server) getSN() (string, error) {
	log.V5("GETTING SERIAL NUMBER")
	var out *exec.Cmd
	if s.Platform == "pi" {
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

func (s *Server) getIp() (*IP, error) {
	if s.Platform != "pi" {
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

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	response := "ok"
	switch action := vars["action"]; action {
	case "start":
		go s.Channels.Start(s.Configuration.Fps, s.Configuration.Offset, s.Configuration.Id)
		response = "starting..."
	case "stop":
		go s.Channels.Stop()
		response = "stopping..."
	}
	_, err := w.Write([]byte(response))
	if err != nil {
		panic(err)
	}
}
