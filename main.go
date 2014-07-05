package main

import (
	irc "github.com/fluffle/goirc/client"
	irclog "github.com/fluffle/goirc/logging"
	"time"
)

func main() {
	config := loadConfig()

	irclog.SetLogger(LogrusAdapter{*log})

	cfg := config.Network.IrcConfig()
	c := irc.Client(cfg)
	c.EnableStateTracking()

	module.InitModules(c, config)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		for _, cmd := range config.Network.OnConnectCmds {
			conn.Raw(cmd)
		}
		time.Sleep(2 * time.Second)
		for channel, _ := range config.Network.Channels {
			conn.Join(channel)
		}
	})

	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	if err := c.ConnectTo(config.Network.Server); err != nil {
		panic(err)
	}

	// Wait for disconnect
	<-quit

}
