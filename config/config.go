package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Nick       string
	Pass       string
	Host       string
	Sasl       string
	Tls        bool
	Channels   []string
	DbPath     string            `yaml:"db-path"`
	ApiKeys    map[string]string `yaml:"api-keys"`
	IgnoreBots bool              `yaml:"ignore-bots"`
}

var C Config

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

	data, err := io.ReadAll(file)
	err = yaml.Unmarshal(data, &C)
	if err != nil {
		log.Fatal(err)
	}

	if C.DbPath == "" {
		C.DbPath = getDbPath()
	}
}
