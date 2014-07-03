package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	// TODO: Support multiple networks
	Network      ConfigNetworks             `json:"network"`
	ModuleConfig map[string]json.RawMessage `json:"module_config,omitempty"`
}

type ChannelConfig struct {
	Modules []string `json:"modules,omitempty"`
}

type ConfigNetworks struct {
	Nick          string                   `json:"nick"`
	Host          string                   `json:"host"`
	Port          int                      `json:"port,omitempty"`
	SSL           bool                     `json:"ssl,omitempty"`
	Channels      map[string]ChannelConfig `json:"channels"`
	OnConnectCmds []string                 `json:"on_connect_command"`
}

func loadConfig() *Config {
	var config *Config
	file, err := ioutil.ReadFile("mipples.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		panic(err)
	}
	return config
}
