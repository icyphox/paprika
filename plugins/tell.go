package plugins

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(TellPlug{})
	RegisterMatcher(TellPlug{})
}

type Tell struct {
	From    string
	To      string
	Message string
	Time    time.Time
}

type TellPlug struct{}

func (TellPlug) Triggers() []string {
	return []string{".tell", ".showtells"}
}

func (TellPlug) Matches(c *irc.Client, m *irc.Message) (string, bool) {
	// always match
	return "", true
}

// Encodes message into encoding/gob for storage.
// We use a hash of the message in the key to ensure
// we don't queue dupes.
func (t *Tell) saveTell() error {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(t); err != nil {
		return err
	}
	// Store key as 'tell/nick/hash'; should help with
	// easy prefix scans for tells.
	err := database.DB.Set(t.getKeyId(), data.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// Encodes message into encoding/gob for storage.
// We use a hash of the message in the key to ensure
// we don't queue dupes.
func (t Tell) getKeyId() []byte {
	// Hash (SHA1) a message for use as key.
	// Helps ensure we don't queue the same
	// message over and over.
	h := sha1.New()
	io.WriteString(h, t.Message)
	hash := h.Sum(nil)
	// Store key as 'tell/nick/hash'; should help with
	// easy prefix scans for tells.
	Key := []byte(fmt.Sprintf("tell/%s/", t.To))
	return append(Key, hash...)
}

// Decodes tell data from encoding/gob into a Tell.
func getTell(data []byte) (*Tell, error) {
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	t := Tell{}
	if err := dec.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (TellPlug) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	if cmd == ".tell" {
		target := strings.SplitN(rest, " ", 2)
		// No message passed.
		if rest == "" || len(target) == 1 {
			c.WriteMessage(NewRes(m, "Usage: .tell <nick> <message>"))
			return
		}

		to := target[0]
		msg := target[1]

		t := Tell{
			From:    strings.ToLower(m.Prefix.Name),
			To:      strings.ToLower(to),
			Message: msg,
			Time:    time.Now(),
		}

		if err := t.saveTell(); err != nil {
			log.Println(err)
			return
		}

		c.WriteMessage(&irc.Message{
			Command: "NOTICE",
			Params:  []string{t.From, "Your message will be sent!"},
		})
		return
	} else {
		// React to all other messages here.
		// Iterate over key prefixes to check if our tell
		// recepient has shown up. Then send his tell and delete
		// the keys.

		// All pending tells.
		tells := []Tell{}

		err := database.DB.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("tell/" + strings.ToLower(m.Prefix.Name))
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				err := item.Value(func(v []byte) error {
					tell, err := getTell(v)
					if err != nil {
						return fmt.Errorf("degobbing: %w", err)
					}
					tells = append(tells, *tell)
					return nil
				})
				if err != nil {
					return fmt.Errorf("iterating: %w", err)
				}

			}
			return nil
		})
		if err != nil {
			log.Println("fetching tells: %w", err)
			return
		} else if len(tells) == 0 {
			return
		}

		// Sort tells by time.
		sort.Slice(tells, func(i, j int) bool {
			return tells[j].Time.After(tells[i].Time)
		})

		for _, tell := range tells {
			resp := &irc.Message{
				Command: "NOTICE",
				Params: []string{tell.To, fmt.Sprintf(
					"%s sent you a message %s: %s\n",
					tell.From, humanize.Time(tell.Time), tell.Message,
				)},
			}

			err = database.DB.Update(func(txn *badger.Txn) error {
				return txn.Delete(tell.getKeyId())
			})

			if err != nil {
				log.Println(err)
			} else {
				c.WriteMessage(resp)
			}

			// only show one at a time unless .showtells
			if cmd != ".showtells" {
				break
			}
		}
	}
}
