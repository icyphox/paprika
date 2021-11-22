// Plugin that looks for links and provides additional info whenever
// someone posts an URL.
// Author: nojusr
package plugins

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"gopkg.in/irc.v3"
)

// This plugin is an example of how to make something that will
// respond (or just have read access to) every message that comes in.
// The plugins.go file has a special case for handling an 'empty' Triggers string.
// on such a case, it will simply run Execute on every message that it sees.
func init() {
	Register(LinkHandler{})
}

type LinkHandler struct{}

func (LinkHandler) Triggers() []string {
	return []string{""}
}

func (LinkHandler) Execute(m *irc.Message) (string, error) {

	var output string

	// in PRIVMSG's case, the second (first, if counting from 0) parameter
	// is the string that contains the complete message.
	for _, value := range strings.Split(m.Params[1], " ") {
		u, err := url.Parse(value)

		if err != nil {
			continue
		}

		// just a test check for the time being.
		// this if statement block will be used for content that is
		// non-generic. I.e it belongs to a specific website, like
		// stackoverflow or youtube.
		if u.Hostname() == "youtube.com" || u.Hostname() == "youtu.be" {
			// TODO finish this
			output += "[Youtube] yeah you definitely posted a youtube link\n"
		} else if len(u.Hostname()) > 0 {
			desc, err := getDescriptionFromURL(value)
			if err != nil {
				log.Printf("Failed to get title from http URL")
				fmt.Println(err)
				continue
			}
			output += fmt.Sprintf("[URL] %s (%s)\n", desc, u.Hostname())
		}
	}

	if len(output) > 0 {
		return output, nil
	} else {
		return "", NoReply // We need to NoReply so we don't consume all messages.
	}
}

// the three funcs below are taken from:
// https://siongui.github.io/2016/05/10/go-get-html-title-via-net-html/
func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func getHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", false
	}

	return traverse(doc)
}

// yoinkies from
// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// provides a basic, short description of whatever is inside
// a posted URL.
func getDescriptionFromURL(url string) (string, error) {
	resp, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	mime := resp.Header.Get("content-type")

	switch mime {
	case "image/jpeg":
		return fmt.Sprintf("JPEG image, size: %s", byteCountSI(resp.ContentLength)), nil
	case "image/png":
		return fmt.Sprintf("PNG image, size: %s", byteCountSI(resp.ContentLength)), nil
	default:
		output, ok := getHtmlTitle(resp.Body)

		if !ok {
			return "", errors.New("Failed to find <title> in html")
		}

		return output, nil
	}
}
