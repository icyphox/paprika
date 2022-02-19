package main

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"strings"

	"git.icyphox.sh/paprika/config"
	"git.icyphox.sh/paprika/database"
	"git.icyphox.sh/paprika/plugins"
	"gopkg.in/irc.v3"
)

func handleChatMessage(c *irc.Client, responses []*irc.Message, err error) {
	if errors.Is(err, plugins.NoReply) {
		return
	}

	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	for _, resp := range responses {
		c.WriteMessage(resp)
	}
}

// GENERAL TODO: We need a way to have plugins send continuations or handlers
//               so we can dynamically react to IRC commands.
//               This should mean we can also populate stateful information generically
//               using only the plugin subsystem.
func ircHandler(c *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001":
		c.Write(config.SplitChannelList(config.C.Channels))
	// TODO: Generalize this
	case "JOIN":
		err := plugins.SeenDoing(m)
		if err != nil && err != plugins.NoReply {
			log.Printf("error: %v", err)
		}
		response, err := plugins.GetIntro(m)
		handleChatMessage(c, []*irc.Message{response}, err)
	case "PART", "QUIT":
		err := plugins.SeenDoing(m)
		if err != nil && err != plugins.NoReply {
			log.Printf("error: %v", err)
		}
	// TODO: Generalize this
	case "NOTICE":
		response, err := plugins.CTCPReply(m)
		handleChatMessage(c, []*irc.Message{response}, err)
	case "PRIVMSG":
		// Trim leading and trailing spaces to not trip up our
		// plugins.
		m.Params[1] = strings.TrimSpace(m.Params[1])
		response, err := plugins.ProcessTrigger(m)
		handleChatMessage(c, response, err)
	// TODO: Generalize this
	case "401":
		response, err := plugins.NoSuchUser(m)
		handleChatMessage(c, []*irc.Message{response}, err)
	}
}

func main() {
	var err error
	var conn io.ReadWriteCloser
	if !config.C.Tls {
		conn, err = net.Dial("tcp", config.C.Host)
	} else {
		conn, err = tls.Dial("tcp", config.C.Host, nil)
	}
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
