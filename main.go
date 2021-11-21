package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"git.icyphox.sh/paprika/config"
	"git.icyphox.sh/paprika/database"
	"git.icyphox.sh/paprika/plugins"
	"gopkg.in/irc.v3"
)

// A simple check func to find out if an incoming irc message
// is supposed to be a CTCP message
func isCTCPmessage(m *irc.Message) bool {
	// CTCP messages are identified by their first character byte
	// being equal to 0x01.
	return m.Params[1][0] == 0x01
}

// CTCP responses are sent with a NOTICE, instead of a PRIVMSG
func sendCTCPResponse(c *irc.Client, m *irc.Message, command string, message string) {
	c.WriteMessage(&irc.Message{
		Command: "NOTICE",
		Params: []string{
			m.Prefix.Name,
			fmt.Sprintf("\x01%s %s\x01", command, message),
		},
	})
}

// for future use. Perhaps move all of this CTCP stuff out another file?
func sendCTCPRequest(c *irc.Client, m *irc.Message, command string, message string) {
	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			m.Prefix.Name,
			fmt.Sprintf("\x01%s %s\x01", command, message),
		},
	})
}

func handleCTCPMessage(c *irc.Client, m *irc.Message) {

	// Refer to for further commands to implement and as a general
	// guide on how CTCP works:
	// https://tools.ietf.org/id/draft-oakley-irc-ctcp-01.html

	var ctcpCommand = m.Params[1] // var might be named wrong

	// start of the main if/else tree for CTCP checking. idk why, but
	// for some reason a straight switch/case comparison simply did not work
	if strings.Contains(ctcpCommand, "VERSION") {
		sendCTCPResponse(c, m, "VERSION", "Paprika v0.0.1")
	} else if strings.Contains(ctcpCommand, "PING") {
		// lotsa ugly string processing here, but w/e

		// split the incoming ping by word, strip out the first word (the command)
		output := strings.Split(m.Params[1], " ")[1:]
		outputStr := strings.Join(output, " ") // re-join

		// send response while stripping out the last char,
		// somehwere deep in the wirting a random char gets added
		// to the start and end of the incomming message.
		// the first char gets stripped out along with the command word.
		// the last char gets stripped out here.
		sendCTCPResponse(c, m, "PING", outputStr[0:len(outputStr)-1])
	} else if strings.Contains(ctcpCommand, "CLIENTINFO") {

		// UPDATE THIS WITH ANY NEW CTCP COMMAND YOU IMPLEMENT
		sendCTCPResponse(
			c, m,
			"CLIENTINFO",
			"VERSION PING",
		)
	}
}

func handleChatMessage(c *irc.Client, m *irc.Message) {
	response, err := plugins.ProcessTrigger(m)
	if err != nil {
		log.Printf("error: %v", err)
	}

	// split the plugin output by it's newlines, send every line
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
		c.Write(config.SplitChannelList(config.C.Channels))
	case "PRIVMSG":
		if isCTCPmessage(m) {
			handleCTCPMessage(c, m)
		} else {
			handleChatMessage(c, m)
		}
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
