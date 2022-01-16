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

func handleChatMessage(c *irc.Client, m *irc.Message, response string, err error) {
	if errors.Is(err, plugins.NoReply) {
		return
	}

	cmd := "PRIVMSG"
	if errors.Is(err, plugins.IsNotice) {
		err = nil
		cmd = "NOTICE"
	}

	target := m.Params[0]
	if serr, ok := err.(*plugins.IsPrivateNotice); ok {
		target = serr.To
		cmd = "NOTICE"
		err = nil
	}

	msg := irc.Message{Command: cmd}
	split := strings.Split(response, "\n")

	if errors.Is(err, plugins.IsRaw) {
		for _, line := range split {
			c.Write(line)
		}
	} else if err != nil {
		msg.Params = []string{target, response}
		c.WriteMessage(&msg)
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
	// TODO: Generalize this
	case "JOIN":
		response, err := plugins.GetIntro(m)
		handleChatMessage(c, m, response, err)
	// TODO: Generalize this
	case "NOTICE":
		response, err := plugins.CTCPReply(m)
		handleChatMessage(c, m, response, err)
	case "PRIVMSG":
		// Trim leading and trailing spaces to not trip up our
		// plugins.
		m.Params[1] = strings.TrimSpace(m.Params[1])
		response, err := plugins.ProcessTrigger(m)
		handleChatMessage(c, m, response, err)
	// TODO: Generalize this
	case "401":
		response, err := plugins.NoSuchUser(m)
		handleChatMessage(c, m, response, err)
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
