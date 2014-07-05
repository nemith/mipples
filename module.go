package main

import (
	"encoding/json"
	irc "github.com/fluffle/goirc/client"
)

var module ModuleStore

func init() {
	module = NewModuleStore()
}

type Module interface {
	Init(*irc.Conn, json.RawMessage)
}

type ModuleStore struct {
	modules map[string]Module
}

func NewModuleStore() ModuleStore {
	return ModuleStore{
		modules: make(map[string]Module),
	}
}

func (m *ModuleStore) Register(name string, module Module) {
	m.modules[name] = module
}

func (m *ModuleStore) InitModules(conn *irc.Conn, config *Config) {
	for name, module := range m.modules {
		modConfig := config.ModuleConfig[name]
		module.Init(conn, modConfig)
	}
}
