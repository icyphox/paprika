package config

import (
	"log"
	"os"
	"strings"

	"github.com/adedomin/indenttext"
)

var (
	Nick = "paprika"
	Pass = ""
	Host = "irc.rizon.net:6667"
	Sasl = ""
	Tls = false
	ChanJoinStr = "JOIN #taigobot-test"
	DbPath = ""
	ApiKeys = make(map[string]string)
)

func init() {
	configPath := ""
	inC := false
	initMode := false
	for _, v := range os.Args {
		switch v {
		case "-c", "--config":
			inC = true
		case "-h", "--help":
			usage()
		default:
			if inC {
				configPath = v
			} else if v == "init" {
				initMode = true
			} else if v == "help" {
				usage()
			}
		}
	}

	if initMode {
		createConfig(configPath)
	}

	var file *os.File
	var err error
	if configPath == "" {
		file = findConfig()
	} else {
		file, err = os.Open(configPath)
		if err != nil {
			log.Printf("Error: Could not open config path.")
			log.Fatalf("- %s", err)
		}
	}
	defer file.Close()

	firstChannel := true
	var chanList strings.Builder
	err = indenttext.Parse(file, func (parents []string, item string, typeof indenttext.ItemType) bool {
		if len(parents) == 1 && typeof == indenttext.Value {
			switch parents[0] {
			case "nick":
				Nick = item
			case "pass":
				Pass = item
			case "host":
				Host = item
			case "sasl":
				Sasl = item
			case "tls":
				if item == "true" {
					Tls = true
				} else {
					Tls = false
				}
			case "channels":
				if firstChannel {
					chanList.WriteString(item)
					firstChannel = false
				} else {
					chanList.WriteByte(',')
					chanList.WriteString(item)
				}
			case "db-path":
				DbPath = item
			}
		} else if len(parents) == 2 && typeof == indenttext.Value {
			if parents[0] == "api-keys" {
				ApiKeys[parents[1]] = item
			}
		}
		return false
	})
	if err != nil {
		log.Fatal(err)
	}

	if chanList.Len() > 0 {
		ChanJoinStr = splitChannelList(chanList.String())
	}

	if DbPath == "" {
		DbPath = getDbPath()
	}
}

