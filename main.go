package main

import (
	"log"
	"net"

	"git.icyphox.sh/taigobot/plugins"
	"gopkg.in/irc.v3"
)

func ircHandler(c *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001":
		// TODO: load this from config
		c.Write("JOIN #taigobot-test")
	case "PRIVMSG":
		if m.Trailing()[:1] == "." {
			err := plugins.ProcessCommands(m.Trailing())
			if err != nil {
				c.Writef("error: %v", err)
			}
		}
	}
}

func main() {
	// TODO: load this from config
	conn, err := net.Dial("tcp", "irc.rizon.net:6667")
	if err != nil {
		log.Fatal(err)
	}

	config := irc.ClientConfig{
		Nick:    "taigobot",
		Pass:    "",
		User:    "taigobot",
		Name:    "taigobot",
		Handler: irc.HandlerFunc(ircHandler),
	}

	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Fatal(err)
	}
}
