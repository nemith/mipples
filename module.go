package main

import (
	"encoding/json"
)

var module ModuleStore

func init() {
	module = NewModuleStore()
}

type Module interface {
	Init(*Irc, json.RawMessage)
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

func (m *ModuleStore) InitModules(irc *Irc, config *Config) {
	for name, module := range m.modules {
		modConfig := config.ModuleConfig[name]
		module.Init(irc, modConfig)
	}
}
