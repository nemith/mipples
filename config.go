package main

import (
	"encoding/json"
	irc "github.com/fluffle/goirc/client"
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
	Server        string                   `json:"host"`
	SSL           bool                     `json:"ssl,omitempty"`
	Channels      map[string]ChannelConfig `json:"channels"`
	OnConnectCmds []string                 `json:"on_connect_commands"`
}

func (cn *ConfigNetworks) IrcConfig() *irc.Config {
	cfg := irc.NewConfig(cn.Nick)
	cfg.SSL = cn.SSL
	cfg.Server = cn.Server

	cfg.Me.Ident = "mipples"
	cfg.Me.Name = "Mipples bot (http://github.com/nemith/mipples)"
	cfg.QuitMessage = "I love dem mipples!"
	return cfg

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
