package main

import (
	"encoding/json"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"regexp"
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

var karmaRegexp = regexp.MustCompile(`^([^\+]+)(\+\+|--)$`)

func (k *KarmaModule) Init(i *Irc, config json.RawMessage) {
	db.AutoMigrate(&Karma{})
	i.AddCommand("karma", k.karmaCmdHandler)
	i.AddMatch(karmaRegexp, k.collectorHandler)
}

func (k *KarmaModule) karmaCmdHandler(conn *irc.Conn, cmd *Cmd) {
	if len(cmd.Args) < 1 {
		cmd.RespondToNick(conn, "You must specifiy a nick to lookup karma on.")
		return
	}
	nick := cmd.Args[0]

	var karma Karma
	db.Where(Karma{Nick: nick}).FirstOrInit(&karma)

	if karma.Nick != "" {
		cmd.Respond(conn, fmt.Sprintf("%s has %d karma", nick, karma.Count))
	}
}

func (k *KarmaModule) collectorHandler(conn *irc.Conn, msg *Privmsg, match []string) {
	nickStr := match[0]
	if msg.Nick == nickStr {
		msg.RespondToNick(conn, "You can't karama yourself, you'll go blind!")
		return
	}

	nick := conn.StateTracker().GetNick(nickStr)
	if nick == nil {
		// not a valid nick
		return
	}
	_, isOnChan := nick.IsOnStr(msg.Channel)
	if !isOnChan {
		// nick is not in the current channel
		return
	}

	var karma Karma
	db.Where(Karma{Nick: nick.Nick}).FirstOrInit(&karma)
	if match[2] == "++" {
		karma.Count++
	} else {
		karma.Count--
	}
	db.Save(&karma)
}
