package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	irc "github.com/fluffle/goirc/client"
	"regexp"
	"strings"
)

func init() {
	module.Register("karma", &KarmaModule{})
}

type KarmaConfig struct {
	AllowDecrement      bool `json:"allow_decrement"`
	ShareAcrossChannels bool `json:"allow_decrement"`
}

// Database table format
type Karma struct {
	Id      int
	Nick    string `sql:"not null;unique"`
	Count   int    `sql:"not null"`
	Channel string
}

func (k Karma) TableName() string {
	return "karma"
}

type KarmaModule struct {
	config *KarmaConfig
}

var karmaRegexp = regexp.MustCompile(`^(` + NickRegexp + `)(\+\+|--)$`)

func (k *KarmaModule) Init(c *irc.Conn, config json.RawMessage) {
	db.AutoMigrate(&Karma{})
	c.HandleBG("PRIVMSG", NewCommandHandler("karma", k.karmaCmdHandler))
	c.HandleBG("PRIVMSG", NewMatchHandler(karmaRegexp, k.collectorHandler))
}

func (k *KarmaModule) karmaCmdHandler(conn *irc.Conn, cmd *Cmd) {
	if len(cmd.Args) < 1 {
		cmd.RespondToNick(conn, "You must specifiy a nick to lookup karma on.")
		return
	}
	nick := strings.TrimSpace(cmd.Args[0])

	log.WithFields(logrus.Fields{
		"srcNick": cmd.Nick,
		"dstNick": nick,
		"channel": cmd.Channel,
	}).Debug("Karma: Looking up nick")

	var karma Karma
	db.Where(Karma{Nick: nick}).FirstOrInit(&karma)

	if karma.Nick != "" {
		cmd.Respond(conn, fmt.Sprintf("%s has %d karma", nick, karma.Count))
	}
}

func (k *KarmaModule) collectorHandler(conn *irc.Conn, msg *Privmsg, match []string) {
	nickStr := strings.TrimSpace(match[1])
	if msg.Nick == nickStr {
		msg.RespondToNick(conn, "You can't karama yourself, you'll go blind!")
		return
	}

	nick := conn.StateTracker().GetNick(nickStr)
	if nick == nil {
		log.WithFields(logrus.Fields{
			"srcNick": msg.Nick,
			"dstNick": nickStr,
			"channel": msg.Channel,
		}).Debug("Karma: unknown nick")
		return
	}

	_, isOnChan := nick.IsOnStr(msg.Channel)
	if !isOnChan {
		log.WithFields(logrus.Fields{
			"srcNick": msg.Nick,
			"dstNick": nickStr,
			"channel": msg.Channel,
		}).Debug("Karma: nick not in the current channel")
		return
	}

	log.WithFields(logrus.Fields{
		"srcNick":   msg.Nick,
		"dstNick":   nickStr,
		"channel":   msg.Channel,
		"direction": match[2],
	}).Info("Karma: Updating nick")

	var karma Karma
	db.Where(Karma{Nick: nick.Nick}).FirstOrInit(&karma)
	if match[2] == "++" {
		karma.Count++
	} else {
		karma.Count--
	}
	db.Save(&karma)
}
