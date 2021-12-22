package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Device struct {
	Id       int    `json:"id,string"`
	Channels int    `json:"channels,string"`
	Ip       string `json:"ip"`
}

type Devices struct {
	List []*Device
}

func Produce() (*Devices, error) {
	list := make([]*Device, 0)

	devices := &Devices{
		List: list,
	}

	err := devices.get()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices - %v", err)
	}

	return devices, nil
}

func (d *Devices) get() error {
	backend := "http://" + Config.BackendHost + ":" + strconv.Itoa(Config.BackendPort) + "/api/devices"
	resp, err := http.Get(backend)
	if err != nil || resp.StatusCode != 200 {
		return fmt.Errorf("failed to get Configuration from remote using local: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Invalid response from server EXITING: %v", err)
	}

	err = json.Unmarshal(body, &d.List)
	if err != nil {
		return fmt.Errorf("failed to unmarshall devices - %v", err)
	}

	return nil
}
