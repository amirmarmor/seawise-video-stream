package core

import (
	"encoding/json"
	"fmt"
	"www.seawise.com/shrimps/backend/exposed"
	"www.seawise.com/shrimps/backend/persistance"
)

type Configuration struct {
	Offset  int    `json:"offset,string"`
	Cleanup bool   `json:"bool,string"`
	Rules   []Rule `json:"rules"`
}

type ConfigManager struct {
	Persist *persistance.Persist
	Config  *Configuration
}

type Rule struct {
	Id        int64  `json:"id,string"`
	Recurring string `json:"recurring"`
	Start     int64  `json:"start,string"`
	Duration  int64  `json:"duration,string"`
	Type      string `json:"type"`
}

func Produce(persistanceApi *persistance.Persist) (*ConfigManager, error) {
	InitFlags()

	manager := ConfigManager{
		Persist: persistanceApi,
	}

	err := manager.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}

	return &manager, nil
}

func (cm *ConfigManager) GetConfig() error {
	configJson, err := cm.Persist.Get(exposed.ConfigKey)
	if err != nil {
		return err
	}

	config := &Configuration{}
	if configJson != "" {
		err = json.Unmarshal([]byte(configJson), &config)
		if err != nil {
			return err
		}
		cm.Config = config
		return nil
	}

	config.Offset = Defaults.Offset

	rules := make([]Rule, 0)
	err = json.Unmarshal([]byte(Defaults.Rules), &rules)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	show := make([]int, 0)
	err = json.Unmarshal([]byte(Defaults.Show), &show)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	record := make([]int, 0)
	err = json.Unmarshal([]byte(Defaults.Record), &record)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}

	config.Rules = rules

	cm.Config = config

	return nil
}

func (cm *ConfigManager) SetConfig(configJson string) error {
	err := cm.Persist.Set(exposed.ConfigKey, configJson)
	if err != nil {
		return fmt.Errorf("failed to set: %v", err)
	}

	config := &Configuration{}
	err = json.Unmarshal([]byte(configJson), config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal: %v", err)
	}
	cm.Config = config
	return nil
}
