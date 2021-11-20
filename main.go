package main

import (
	"log"
	"net"

	"git.icyphox.sh/paprika/database"
	"git.icyphox.sh/paprika/plugins"
	"gopkg.in/irc.v3"
)

func ircHandler(c *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001":
		// TODO: load this from config
		c.Write("JOIN #taigobot-test")
	case "PRIVMSG":
		response, err := plugins.ProcessTrigger(m)
		if err != nil {
			log.Printf("error: %v", err)
		}
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				response,
			},
		})
	}
}

func main() {
	// TODO: load this from config
	conn, err := net.Dial("tcp", "irc.rizon.net:6667")
	if err != nil {
		log.Fatal(err)
	}

	config := irc.ClientConfig{
		Nick:    "paprika112",
		Pass:    "",
		User:    "paprika112",
		Name:    "paprika",
		Handler: irc.HandlerFunc(ircHandler),
	}

	database.DB.DB, err = database.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Fatal(err)
	}
}
