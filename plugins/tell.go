package plugins

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"github.com/dustin/go-humanize"
	"gopkg.in/irc.v3"
)

func init() {
	Register(Tell{})
}

type Tell struct {
	From    string
	To      string
	Message string
	Time    time.Time
}

func (Tell) Triggers() []string {
	return []string{".tell", ""}
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

	key := []byte(fmt.Sprintf("tell/%s/", t.To))
	key = append(key, hash...)
	err := database.DB.Set(key, data.Bytes())
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

func (t Tell) Execute(m *irc.Message) (string, error) {
	parts := strings.SplitN(m.Trailing(), " ", 3)

	if parts[0] == ".tell" {
		// No message passed.
		if len(parts) == 2 {
			return "Usage: .tell <nick> <message>", nil
		}

		t.From = strings.ToLower(m.Prefix.Name)
		t.To = strings.ToLower(parts[1])
		t.Message = parts[2]
		t.Time = time.Now()

		if err := t.saveTell(); err != nil {
			return "Error saving message", err
		}

		return "Your message will be sent!", &IsPrivateNotice{t.From}
	} else {
		// React to all other messages here.
		// Iterate over key prefixes to check if our tell
		// recepient has shown up. Then send his tell and delete
		// the keys.

		// All pending tells.
		tells := []Tell{}

		err := database.DB.Update(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("tell/" + strings.ToLower(m.Prefix.Name))
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				k := item.Key()
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
				err = txn.Delete(k)
				if err != nil {
					return fmt.Errorf("deleting key: %w", err)
				}
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("fetching tells: %w", err)
		}

		// No tells for this user.
		if len(tells) == 0 {
			return "", NoReply
		}

		// Sort tells by time.
		sort.Slice(tells, func(i, j int) bool {
			return tells[j].Time.Before(tells[i].Time)
		})

		// Formatted tells in a slice, for joining into a string
		// later.
		tellsFmtd := strings.Builder{}
		for _, tell := range tells {
			s := fmt.Sprintf(
				"%s sent you a message %s: %s\n",
				tell.From, humanize.Time(tell.Time), tell.Message,
			)
			tellsFmtd.WriteString(s)
		}
		return tellsFmtd.String(), &IsPrivateNotice{To: tells[0].To}
	}
}
