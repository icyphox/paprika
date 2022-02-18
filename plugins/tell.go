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
	Key     []byte
	From    string
	To      string
	Message string
	Time    time.Time
}

type TellPlug struct{}

func (TellPlug) Triggers() []string {
	return []string{".tell"}
}

func (TellPlug) Matches(m *irc.Message) (string, error) {
	// always match
	return "", nil
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
	hash := hashMessage(t.Message)

	t.Key = []byte(fmt.Sprintf("tell/%s/", t.To))
	t.Key = append(t.Key, hash...)
	err := database.DB.Set(t.Key, data.Bytes())
	if err != nil {
		return err
	}
	return nil
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

// Hash (SHA1) a message for use as key.
// Helps ensure we don't queue the same
// message over and over.
func hashMessage(msg string) []byte {
	h := sha1.New()
	io.WriteString(h, msg)
	return h.Sum(nil)
}

func (TellPlug) Execute(cmd, rest string, m *irc.Message) (*irc.Message, error) {
	if cmd == ".tell" {
		// No message passed.
		if rest == "" {
			return NewRes(m, "Usage: .tell <nick> <message>"), nil
		}

		t := Tell{
			From:    strings.ToLower(m.Prefix.Name),
			To:      strings.ToLower(rest),
			Message: rest,
			Time:    time.Now(),
		}

		if err := t.saveTell(); err != nil {
			return nil, err
		}

		return &irc.Message{
			Command: "NOTICE",
			Params:  []string{t.From, "Your message will be sent!"},
		}, nil
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
			return nil, fmt.Errorf("fetching tells: %w", err)
		}

		// No tells for this user.
		if len(tells) == 0 {
			return nil, NoReply
		}

		// Sort tells by time.
		sort.Slice(tells, func(i, j int) bool {
			return tells[j].Time.Before(tells[i].Time)
		})

		tell := tells[0]

		resp := &irc.Message{
			Command: "NOTICE",
			Params: []string{tell.To, fmt.Sprintf(
				"%s sent you a message %s: %s\n",
				tell.From, humanize.Time(tell.Time), tell.Message,
			)},
		}

		err = database.DB.Update(func(txn *badger.Txn) error {
			return txn.Delete(tell.Key)
		})
		if err != nil {
			log.Println(err)
		}

		return resp, nil
	}
}
