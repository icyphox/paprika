package config

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func SplitChannelList(channels []string) string {
	lineSize := 0
	first := true

	var ret strings.Builder

	// Splits configured list of channels into a safe set of commands
	for _, channel := range channels {
		if lineSize+len(channel) > 510 {
			lineSize = 0
			first = true
			ret.WriteByte('\r')
			ret.WriteByte('\n')
		}

		if !first {
			ret.WriteByte(',')
			lineSize += 1
		} else {
			ret.WriteString("JOIN ")
			lineSize += 5
			first = false
		}
		ret.WriteString(channel)
		lineSize += len(channel)
	}
	ret.WriteByte('\r')
	ret.WriteByte('\n')

	return ret.String()
}

func getDbPath() string {
	// https://www.freedesktop.org/software/systemd/man/systemd.exec.html#RuntimeDirectory=
	systemdStateDir := os.Getenv("STATE_DIRECTORY")
	xdgDataDir := os.Getenv("XDG_DATA_HOME")
	if systemdStateDir == "" && xdgDataDir == "" {
		log.Print("Warning: Nowhere to store state!")
		log.Print("- Please add StateDirectory= to your systemd unit or add db-path to your config.")

		// TODO: change this to os.MkdirTemp when 1.17 is more common
		dir, err := ioutil.TempDir("", "paprika-bot")
		if err != nil {
			log.Fatalf("Failed making temporary area: %s", err)
		}
		log.Printf("Warning: Generated a temporary path: %s", dir)
		return dir
	} else if xdgDataDir != "" {
		return path.Join(xdgDataDir, "paprika")
	} else {
		return systemdStateDir
	}
}

func usage() {
	println("Usage: paprika [init] [-c config]\n")
	println("  init       Initialize configuration for the bot.")
	println("  -c config  Use config given on the command line.\n")
	os.Exit(1)
}
