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
	
	if errors.Is(err, plugins.NoReply) {
		return
	}

	cmd := "PRIVMSG"
	if errors.Is(err, plugins.IsNotice) {
		err = nil
		cmd = "NOTICE"
	}
	msg := irc.Message {Command: cmd}
	target := m.Params[0]
	split := strings.Split(response, "\n")

	if errors.Is(err, plugins.IsRaw) {
		for _, line := range split {
			c.Write(line)
		}
	} else if err != nil {
		log.Printf("error: %v", err)
	} else {
		for _, line := range split {
			msg.Params = []string{target, line}
			c.WriteMessage(&msg)
		}
		return
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
