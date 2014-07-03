package main

import (
	irc "github.com/fluffle/goirc/client"
)

func main() {
	config := loadConfig()

	c := Irc{irc.SimpleClient(config.Network.Nick)}
	c.SSL = config.Network.SSL
	c.EnableStateTracking()

	c.AddHandler(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		for _, cmd := range config.Network.OnConnectCmds {
			conn.Raw(cmd)
		}
		for channel, _ := range config.Network.Channels {
			conn.Join(channel)
		}
	})

	quit := make(chan bool)
	c.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	module.InitModules(&c, config)

	if err := c.Connect(config.Network.Host); err != nil {
		panic(err)
	}

	// Wait for disconnect
	<-quit

}
