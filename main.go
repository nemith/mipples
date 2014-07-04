package main

import (
	irc "github.com/fluffle/goirc/client"
)

func main() {
	config := loadConfig()

	cfg := config.Network.IrcConfig()
	c := Irc{irc.Client(cfg)}
	c.EnableStateTracking()

	c.HandleFunc("connected", func(conn *irc.Conn, line *irc.Line) {
		for _, cmd := range config.Network.OnConnectCmds {
			conn.Raw(cmd)
		}
		for channel, _ := range config.Network.Channels {
			conn.Join(channel)
		}
	})

	quit := make(chan bool)
	c.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	module.InitModules(&c, config)

	if err := c.ConnectTo(config.Network.Server); err != nil {
		panic(err)
	}

	// Wait for disconnect
	<-quit

}
