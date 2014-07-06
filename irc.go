package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"regexp"
	"strings"
)

const (
	NickRegexp = `(?i)[a-z;\[\]\\` + "`" + `_^{}|][a-z0-9;\[\]\\` + "`" + `_^{}|-]*`
)

type CommandHandlerFunc func(*irc.Conn, *Cmd)

type Cmd struct {
	Command string
	Args    []string
	*Privmsg
}

type CommandHandler struct {
	command     string
	handlerFunc CommandHandlerFunc
}

func NewCommandHandler(command string, handler CommandHandlerFunc) *CommandHandler {
	log.WithFields(logrus.Fields{
		"command": command,
	}).Debug("Registering new irc command")
	return &CommandHandler{
		command:     command,
		handlerFunc: handler,
	}

}

func (h CommandHandler) Handle(c *irc.Conn, line *irc.Line) {
	msg := parsePrivmsg(line)
	if msg.Public() {
		h.command = fmt.Sprintf("!%s", h.command)
	}
	if strings.HasPrefix(msg.Text, h.command) {
		log.WithFields(logrus.Fields{
			"nick":    msg.Nick,
			"channel": msg.Channel,
			"text":    msg.Text,
			"command": h.command,
		}).Debug("Executing irc command")
		h.handlerFunc(c, &Cmd{
			Command: h.command,
			Args:    strings.Split(msg.Text, " ")[1:],
			Privmsg: msg})
	}
}

type MatchHandlerFunc func(*irc.Conn, *Privmsg, []string)

type MatchHandler struct {
	matcher     *regexp.Regexp
	handlerFunc MatchHandlerFunc
}

func NewMatchHandler(matcher *regexp.Regexp, handler MatchHandlerFunc) *MatchHandler {
	log.WithFields(logrus.Fields{
		"matcher": matcher,
	}).Debug("Registering new irc regex matcher")
	return &MatchHandler{
		matcher:     matcher,
		handlerFunc: handler,
	}
}

func (h *MatchHandler) Handle(c *irc.Conn, line *irc.Line) {
	msg := parsePrivmsg(line)

	matches := h.matcher.FindStringSubmatch(msg.Text)
	if matches == nil {
		return
	}
	log.WithFields(logrus.Fields{
		"nick":    msg.Nick,
		"channel": msg.Channel,
		"text":    msg.Text,
		"matcher": h.matcher,
		"matches": matches,
	}).Debug("Executing irc regex matcher")
	h.handlerFunc(c, msg, matches)
}

type MatchAllHandlerFunc func(*irc.Conn, *Privmsg, [][]string)

type MatchAllHandler struct {
	matcher     *regexp.Regexp
	handlerFunc MatchAllHandlerFunc
}

func NewMatchAllHandler(matcher *regexp.Regexp, handler MatchAllHandlerFunc) *MatchAllHandler {
	log.WithFields(logrus.Fields{
		"matcher": matcher,
	}).Debug("Registering new irc regex all matcher")
	return &MatchAllHandler{
		matcher:     matcher,
		handlerFunc: handler,
	}
}

func (h *MatchAllHandler) Handle(c *irc.Conn, line *irc.Line) {
	msg := parsePrivmsg(line)

	matches := h.matcher.FindAllStringSubmatch(msg.Text, -1)
	if matches == nil {
		return
	}
	log.WithFields(logrus.Fields{
		"nick":    msg.Nick,
		"channel": msg.Channel,
		"text":    msg.Text,
		"matcher": h.matcher,
		"matches": matches,
	}).Debug("Executing irc regex all matcher")
	h.handlerFunc(c, msg, matches)
}

type Privmsg struct {
	Channel, Text string
	*irc.Line
}

func parsePrivmsg(line *irc.Line) *Privmsg {
	return &Privmsg{
		line.Args[0], // Channel
		line.Args[1], // Text
		line,
	}
}

func (m *Privmsg) respond(includeToNick bool, conn *irc.Conn, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if m.Public() {
		if includeToNick {
			msg = fmt.Sprintf("%s: %s", m.Nick, msg)
		}
		conn.Privmsg(m.Channel, msg)
	} else {
		conn.Privmsg(m.Nick, msg)
	}
}

func (m *Privmsg) Respond(conn *irc.Conn, format string, args ...interface{}) {
	m.respond(false, conn, format, args...)
}

func (m *Privmsg) RespondToNick(conn *irc.Conn, format string, args ...interface{}) {
	m.respond(true, conn, format, args...)
}
