package main

// Usage: go run new-plugin.go <name>

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var name string
	if len(os.Args) == 2 {
		name = os.Args[1]
	} else {
		fmt.Println("usage: new-plugin <name>")
		os.Exit(1)
	}

	// make plugin name Title case
	tc := strings.Title(name)

	template := `package plugins

import (
	"gopkg.in/irc.v3"
)

func init() {
	Register(%s{})
}

type %s struct{}

func (%s) Triggers() []string {
	return []string{}
}

func (%s) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	return "", nil
}`
	out := fmt.Sprintf(template, tc, tc, tc, tc)
	if err := os.WriteFile(filepath.Join("plugins", name+".go"), []byte(out), 0644); err != nil {
		fmt.Println(err)
	}

}
