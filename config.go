package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	irc "github.com/fluffle/goirc/client"
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
	Pass          string                   `json:"pass"`
	SSL           bool                     `json:"ssl,omitempty"`
	Channels      map[string]ChannelConfig `json:"channels"`
	OnConnectCmds []string                 `json:"on_connect_commands"`
}

func (cn *ConfigNetworks) GoIrcConfig() *irc.Config {
	cfg := irc.NewConfig(cn.Nick)
	cfg.SSL = cn.SSL
	cfg.Server = cn.Server
	cfg.Pass = cn.Pass

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
		syntaxErr, ok := err.(*json.SyntaxError)
		if !ok {
			log.Fatalf("Cannot read config: %s", err)
		}

		// We have a syntax error. Extract out the line number and friends.
		// https://groups.google.com/forum/#!topic/golang-nuts/fizimmXtVfc
		newline := []byte{'\x0a'}

		// Calculate the start/end position of the line where the error is
		start := bytes.LastIndex(file[:syntaxErr.Offset], newline) + 1
		end := len(file)
		if idx := bytes.Index(file[start:], newline); idx >= 0 {
			end = start + idx
		}

		// Count the line number we're on plus the offset in the line
		line := bytes.Count(file[:start], newline) + 1
		pos := int(syntaxErr.Offset) - start - 1

		log.Fatalf("Cannot read config. Error in line %d, char %d: %s\n%s",
			line, pos, syntaxErr, file[start:end])
	}

	return config
}
