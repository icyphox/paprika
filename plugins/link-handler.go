// Plugin that looks for links and provides additional info whenever
// someone posts an URL.
// Author: nojusr
package plugins

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"gopkg.in/irc.v3"
)

// NOTE(nojusr): This plugin is an example of how to make something that will
// respond (or just have read access to) every message that comes in.
// The plugins.go file has a special case for handling an 'empty' Triggers string.
// on such a case, it will simply run Execute on every message that it sees.

func init() {
	Register(LinkHandler{})
}

type LinkHandler struct{}

func (LinkHandler) Triggers() []string {
	return []string{""} // NOTE(nojusr): More than a single empty string here will probs get you undefined behaviour
}

func (LinkHandler) Execute(m *irc.Message) (string, error) {

	var output string

	// NOTE(nojusr): in PRIVMSG's case, the second (first, if counting from 0) parameter
	// is the string that contains the complete message.
	for _, value := range strings.Split(m.Params[1], " ") {
		u, err := url.Parse(value)

		if err != nil {
			continue
		}

		// NOTE(nojusr): just a test check for the time being.
		// this if statement block will be used for content that is
		// non-generic. I.e it belongs to a specific website, like
		// stackoverflow or youtube.
		if u.Hostname() == "youtube.com" || u.Hostname() == "youtu.be" {
			// TODO(nojusr): finish this
			output += "[Youtube] yeah you definitely posted a youtube link"
		} else if len(u.Hostname()) > 0 {
			output += "[URL] "
			output += getDescriptionFromURL(value)

		}
	}

	return output, nil
}

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

// NOTE(nojusr): yoinkies from
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

// NOTE(nojusr): provides a basic, short description of whatever is inside
// a posted URL.
func getDescriptionFromURL(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	mime := resp.Header.Get("content-type")

	switch mime {
	case "image/jpeg":
		return "JPEG image (" + byteCountSI(resp.ContentLength) + ")"
	case "image/png":
		return "PNG image (" + byteCountSI(resp.ContentLength) + ")"
	default:
		output, err := getHtmlTitle(resp.Body)
		if err == false {
			log.Printf("Failed to get title from http URL")
			return url // NOTE(nojusr): if you fuck up, just pretend that everythin is ok :)
		}
		return output + "(" + url + ")"
	}
}
