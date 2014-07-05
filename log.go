package main

import (
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Level = logrus.Debug // #Temporary
}

type LogrusAdapter struct {
	logrus.Logger
}

func (a LogrusAdapter) Debug(format string, args ...interface{}) {
	a.Debugf(format, args...)
}

func (a LogrusAdapter) Info(format string, args ...interface{}) {
	a.Infof(format, args...)
}

func (a LogrusAdapter) Warn(format string, args ...interface{}) {
	a.Warnf(format, args...)
}

func (a LogrusAdapter) Error(format string, args ...interface{}) {
	a.Errorf(format, args...)
}
