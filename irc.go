package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"regexp"
	"strings"
)

type CommandHandler func(*irc.Conn, *Cmd)
type MatchHandler func(*irc.Conn, *Privmsg, []string)

type Cmd struct {
	Command string
	Args    []string
	*Privmsg
}

type BangCmdHandler struct {
	command     string
	handlerFunc CommandHandler
}

func NewBangCmd(command string, handler CommandHandler) *BangCmdHandler {
	log.WithFields(logrus.Fields{
		"command": command,
	}).Debug("Registering new irc command")
	return &BangCmdHandler{
		command:     command,
		handlerFunc: handler,
	}

}

func (bch BangCmdHandler) Handle(c *irc.Conn, line *irc.Line) {
	msg := parsePrivmsg(line)
	if msg.Public() {
		bch.command = fmt.Sprintf("!%s", bch.command)
	}
	if strings.HasPrefix(msg.Text, bch.command) {
		log.WithFields(logrus.Fields{
			"nick":    msg.Nick,
			"channel": msg.Channel,
			"text":    msg.Text,
			"command": bch.command,
		}).Debug("Executing irc command")
		bch.handlerFunc(c, &Cmd{
			Command: bch.command,
			Args:    strings.Split(msg.Text, " ")[1:],
			Privmsg: msg})
	}
}

type RegexpMatchHandler struct {
	matcher     *regexp.Regexp
	handlerFunc MatchHandler
}

func NewRegexpMatch(matcher *regexp.Regexp, handler MatchHandler) *RegexpMatchHandler {
	log.WithFields(logrus.Fields{
		"matcher": matcher,
	}).Debug("Registering new irc regex matcher")
	return &RegexpMatchHandler{
		matcher:     matcher,
		handlerFunc: handler,
	}
}

func (rmh RegexpMatchHandler) Handle(c *irc.Conn, line *irc.Line) {
	msg := parsePrivmsg(line)

	matches := rmh.matcher.FindStringSubmatch(msg.Text)
	if matches == nil {
		return
	}
	log.WithFields(logrus.Fields{
		"nick":    msg.Nick,
		"channel": msg.Channel,
		"text":    msg.Text,
		"matcher": rmh.matcher,
		"matches": matches,
	}).Debug("Executing irc regex matcher")
	rmh.handlerFunc(c, msg, matches)
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

func (m *Privmsg) respond(conn *irc.Conn, msg string, includeToNick bool) {
	if m.Public() {
		if includeToNick {
			msg = fmt.Sprintf("%s: %s", m.Nick, msg)
		}
		conn.Privmsg(m.Channel, msg)
	} else {
		conn.Privmsg(m.Nick, msg)
	}
}

func (m *Privmsg) Respond(conn *irc.Conn, msg string) {
	m.respond(conn, msg, false)
}

func (m *Privmsg) RespondToNick(conn *irc.Conn, msg string) {
	m.respond(conn, msg, true)
}
