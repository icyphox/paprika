package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"strings"

	"git.icyphox.sh/paprika/config"
	"git.icyphox.sh/paprika/database"
	"git.icyphox.sh/paprika/plugins"
	"gopkg.in/irc.v3"
)

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
		plugins.SeenDoing(c, m)
		plugins.GetIntro(c, m)
	case "PART", "QUIT":
		plugins.SeenDoing(c, m)
	// TODO: Generalize this
	case "NOTICE":
		plugins.CTCPReply(c, m)
	case "PRIVMSG":
		// Trim leading and trailing spaces to not trip up our
		// plugins.
		m.Params[1] = strings.TrimSpace(m.Params[1])
		plugins.ProcessTrigger(c, m)
	// TODO: Generalize this
	case "401":
		plugins.NoSuchUser(c, m)
	}
}

func main() {
	var err error
	var conn io.ReadWriteCloser
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
