package main

import (
	"log"
	"net"
	"strings"

	"git.icyphox.sh/taigobot/db"
	"git.icyphox.sh/taigobot/plugins"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

// (nojusr): A simple check func to find out if an irc message
// is supposed to be a CTCP message
func isCTCPmessage(m *irc.Message) bool {
	// NOTE(nojusr): CTCP messages are identified by their first character byte
	// being equal to 0x01.
	if m.Params[1][0] == 0x01 {
		return true
	}
	return false
}

func handleCTCPMessage(c *irc.Client, m *irc.Message) {

	// (nojusr): Refer to for further commands to implement and as a general
	// guide on how CTCP works:
	// https://tools.ietf.org/id/draft-oakley-irc-ctcp-01.html

	var ctcpCommand = m.Params[1]

	// NOTE(nojusr): start of the main if/else tree for CTCP checking. idk why, but
	// for some reason a straight switch/case comparison simply did not work
	if strings.Contains(ctcpCommand, "VERSION") {

		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Prefix.Name,
				// TODO(nojusr): read version from config
				string(0x01) + "VERSION NONE OF YOUR BUISSNESS" + string(0x01),
			},
		})
	}
}

func handleChatMessage(c *irc.Client, m *irc.Message) {
	response, err := plugins.ProcessTrigger(m)
	if err != nil {
		log.Printf("error: %v", err)
	}

	// NOTE(nojusr): split the plugin output by it's newlines, send every line
	// as a separate PRIVMSG
	split := strings.Split(response, "\n")

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

func ircHandler(c *irc.Client, m *irc.Message) {
	switch m.Command {
	case "001":
		// TODO: load this from config
		c.Write("JOIN #taigobot-test")
	case "PRIVMSG":
		if isCTCPmessage(m) {
			handleCTCPMessage(c, m)
		} else {
			handleChatMessage(c, m)
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
		Nick:    "kaligobot",
		Pass:    "",
		User:    "kaligobot",
		Name:    "bitch",
		Handler: irc.HandlerFunc(ircHandler),
	}

	db.DB, err = badger.Open(badger.DefaultOptions("./badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	client := irc.NewClient(conn, config)
	err = client.Run()
	if err != nil {
		log.Fatal(err)
	}
}
