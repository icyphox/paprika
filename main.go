package main

import (
	"errors"
	"log"
	"net"
	"strings"

	"git.icyphox.sh/paprika/config"
	"git.icyphox.sh/paprika/database"
	"git.icyphox.sh/paprika/plugins"
	"gopkg.in/irc.v3"
)

func handleChatMessage(c *irc.Client, m *irc.Message) {
	response, err := plugins.ProcessTrigger(m)
	split := strings.Split(response, "\n")

	if plugins.IsReplyT(err) {
		r := err.(*plugins.ReplyT)
		r.ApplyFlags(m)

		for _, line := range split {
			c.WriteMessage(&irc.Message{
				Command: m.Command,
				Params: []string {
					m.Params[0],
					line,
				},
			})
		}
	} else if err != nil {
		if !errors.Is(err, plugins.NoReply) {
			log.Printf("error: %v", err)
		}
	} else {
		for _, line := range split {
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					m.Params[0],
					line,
				},
			})
		}
	}
}

func ircHandler(c *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001":
		c.Write(config.SplitChannelList(config.C.Channels))
	case "PRIVMSG":
		handleChatMessage(c, m)
	}
}

func main() {
	conn, err := net.Dial("tcp", config.C.Host)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ircConfig := irc.ClientConfig{
		Nick:    config.C.Nick,
		Pass:    config.C.Pass,
		User:    "paprikabot",
		Name:    config.C.Nick,
		Handler: irc.HandlerFunc(ircHandler),
	}

	database.DB.DB, err = database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	client := irc.NewClient(conn, ircConfig)
	err = client.Run()
	if err != nil {
		log.Fatal(err)
	}
}
