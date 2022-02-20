package plugins

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"gopkg.in/irc.v3"
)

// This plugin is an example of how to make something that will
// respond (or just have read access to) every message that comes in.
// The plugins.go file has a special case for handling an 'empty' Triggers string.
// on such a case, it will simply run Execute on every message that it sees.
func init() {
	RegisterMatcher(LinkHandler{})
}

type LinkHandler struct{}

func (LinkHandler) Triggers() []string {
	return nil
}

func (LinkHandler) Matches(_ *irc.Client, m *irc.Message) (string, bool) {
	msg := m.Trailing()
	if strings.HasPrefix(msg, ".") {
		return "", false
	}

	for _, word := range strings.Split(msg, " ") {
		u, err := url.Parse(word)
		if err == nil && u != nil && (u.Scheme == "http" || u.Scheme == "https") {
			return word, true
		}
	}

	return "", false
}

func (LinkHandler) Execute(match, msg string, c *irc.Client, m *irc.Message) {
	u, err := url.Parse(match)
	if err != nil {
		panic("UNREACHABLE: We parse this in Matches, wtf?")
	}

	// just a test check for the time being.
	// this if statement block will be used for content that is
	// non-generic. I.e it belongs to a specific website, like
	// stackoverflow or youtube.
	if u.Hostname() == "www.youtube.com" || u.Hostname() == "youtube.com" || u.Hostname() == "youtu.be" {
		yt, err := YoutubeDescriptionFromUrl(u)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, yt))
		}
	} else {
		desc, err := getDescriptionFromURL(match)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, fmt.Sprintf("[URL] %s (%s)\n", desc, u.Hostname())))
		}
	}
}

// the three funcs below are taken from:
// https://siongui.github.io/2016/05/10/go-get-html-title-via-net-html/
func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) string {
	if isTitleElement(n) {
		return n.FirstChild.Data
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := traverse(c)
		if result != "" {
			return result
		}
	}

	return ""
}

func getHtmlTitle(r io.Reader, contentType string) string {
	readable, err := charset.NewReader(r, contentType)
	if err != nil {
		return err.Error()
	}

	doc, err := html.Parse(readable)
	if err != nil {
		return err.Error()
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
	// try to handle crap like blah (fdasfdsafdsaf https://example.org/)
	if resp != nil && resp.StatusCode == 404 && url[len(url)-1] == ')' {
		resp, err = http.Get(url[:len(url)-1])
	}

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	contentType := resp.Header.Get("content-type")
	if contentType == "" {
		return "- Unknown Content -", nil
	}

	media, _, _ := mime.ParseMediaType(contentType)

	if strings.Contains(media, "html") ||
		strings.Contains(media, "xml") || strings.Contains(media, "xhtml") {

		if title := getHtmlTitle(resp.Body, contentType); title != "" {
			return title, nil
		}
		return "- No Title -", nil
	}

	return fmt.Sprintf("%s, size: %s", media, byteCountSI(resp.ContentLength)), nil
}
