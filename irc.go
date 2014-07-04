package main

import (
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"regexp"
	"strings"
	"unicode/utf8"
)

type CommandHandler func(*irc.Conn, *Cmd)
type MatchHandler func(*irc.Conn, *Privmsg, []string)

type Irc struct {
	*irc.Conn
}

func (c *Irc) AddCommand(command string, handler CommandHandler) {
	c.HandleFunc("PRIVMSG", func(c *irc.Conn, line *irc.Line) {
		command := command
		msg := parsePrivmsg(line)
		if msg.isToChannel() {
			command = fmt.Sprintf("!%s", command)
		}
		if strings.HasPrefix(msg.Text, command) {
			handler(c, &Cmd{
				Command: command,
				Args:    strings.Split(msg.Text, " ")[1:],
				Privmsg: msg})
		}
	})
}

func (c *Irc) AddMatch(matcher *regexp.Regexp, handler MatchHandler) {
	c.HandleFunc("PRIVMSG", func(c *irc.Conn, line *irc.Line) {
		msg := parsePrivmsg(line)

		matches := matcher.FindStringSubmatch(msg.Text)
		if matches == nil {
			return
		}
		handler(c, msg, matches)
	})
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
	if m.isToChannel() {
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

func (m *Privmsg) isToChannel() bool {
	fmt.Println(m.Channel)
	chr, _ := utf8.DecodeRuneInString(m.Channel)
	if chr == '#' || chr == '&' {
		return true
	}
	return false
}

type Cmd struct {
	Command string
	Args    []string
	*Privmsg
}
