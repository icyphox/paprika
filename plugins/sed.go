package plugins

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"gopkg.in/irc.v3"
)

func init() {
	Register(Sed{})
	RegisterMatcher(Sed{})
}

type Sed struct{}

func (Sed) Triggers() []string {
	return []string{".s", ".sed"}
}

type SedCommand struct {
	Delim   rune
	Target  string
	Replace string
	Nick    string
}

type historyEntry struct {
	Channel string
	Nick    string
	Message string
}

type historyBuffer struct {
	head int
	buff []historyEntry
	lock sync.Mutex
}

func (h *historyBuffer) addMessage(m *irc.Message) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.buff[h.head].Channel = m.Params[0]
	h.buff[h.head].Nick = m.Prefix.Name
	h.buff[h.head].Message = m.Trailing()

	h.head++
	if h.head >= len(h.buff) {
		h.head = 0
	}
}

func (h *historyBuffer) findMessage(e *historyEntry) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Head is start of next, head-1 is the message that triggered this.
	for start := h.head - 2; start != h.head; start-- {
		if start < 0 {
			start = len(h.buff) - 1
		}

		hist := h.buff[start]

		if e.Channel == hist.Channel && e.Nick == hist.Nick && strings.Contains(hist.Message, e.Message) {
			e.Message = hist.Message
			return true
		}
	}
	return false
}

var histbuf = historyBuffer{
	head: 0,
	buff: make([]historyEntry, 500, 500),
	lock: sync.Mutex{},
}

var (
	ErrSedNotSubstitute = errors.New("not a substitute command")
	ErrSedTooShort      = errors.New("sed command too short")
	ErrSedIncomplete    = errors.New("too many or too few divisions")
)

func fromString(s string) (*SedCommand, error) {
	myS := s
	comm := &SedCommand{}
	if !strings.HasPrefix(myS, "s") {
		return nil, ErrSedNotSubstitute

	}
	// handle s@@@ and substitute@@@
	if strings.HasPrefix(myS, "substitute") {
		myS = strings.Replace(myS, "substitute", "s", 1)
	}

	/// s@@@ has a length of 4 but does nothing
	if len(myS) < 5 {
		return nil, ErrSedTooShort
	}

	comm.Delim, _ = utf8.DecodeRuneInString(myS[1:])
	parts := strings.Split(myS, string(comm.Delim))
	if len(parts) != 4 {
		return nil, ErrSedIncomplete
	}

	comm.Target = parts[1]
	comm.Replace = fmt.Sprintf("\x02%s\x02", parts[2])
	comm.Nick = parts[3]

	return comm, nil
}

func (Sed) Matches(c *irc.Client, m *irc.Message) (string, bool) {
	histbuf.addMessage(m)

	msg := m.Trailing()
	if _, err := fromString(msg); err != nil {
		return "", false
	}

	return msg, true
}

func (Sed) Execute(sedstr, rest string, c *irc.Client, m *irc.Message) {
	if strings.HasPrefix(sedstr, ".s") {
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				"usage: s/string(s) to replace/replacement/g <- you can use numbers from 1-9 for specific position"},
		})
	} else {
		s, err := fromString(sedstr)
		if err != nil {
			panic("This was supposed to be parsed in the matcher!")
		}

		rewriteNick := m.Prefix.Name
		if s.Nick != "" {
			rewriteNick = s.Nick
		}
		searcher := historyEntry{
			Channel: m.Params[0],
			Nick:    rewriteNick,
			Message: s.Target,
		}

		if ok := histbuf.findMessage(&searcher); ok {
			newMsg := strings.Replace(searcher.Message, s.Target, s.Replace, -1)
			c.WriteMessage(NewRes(m, fmt.Sprintf("<%s> %s", rewriteNick, newMsg)))
		}
	}
}
