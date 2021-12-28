package plugins

import (
	"bytes"
	"crypto/rand"
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
func (t *Tell) saveTell() error {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(t); err != nil {
		return err
	}
	// Store key as 'tell/nick/randbytes'; should help with
	// easy prefix scans for tells.
	rnd := make([]byte, 8)
	rand.Read(rnd)

	key := []byte(fmt.Sprintf("tell/%s/", t.To))
	key = append(key, rnd...)
	err := database.DB.Set(key, data.Bytes())
	if err != nil {
		return err
	}
	return nil
}

// Decodes tell data from encoding/gob into a Tell.
func getTell(data io.Reader) (*Tell, error) {
	dec := gob.NewDecoder(data)
	t := Tell{}
	if err := dec.Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
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
			prefix := []byte("tell/" + m.Prefix.Name)
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				k := item.Key()
				err := item.Value(func(v []byte) error {
					var r bytes.Reader
					tell, err := getTell(&r)
					if err != nil {
						return err
					}
					tells = append(tells, *tell)
					return nil
				})
				if err != nil {
					return err
				}
				err = txn.Delete(k)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return "", err
		}

		// Sort tells by time.
		sort.Slice(tells, func(i, j int) bool {
			return tells[i].Time.Before(tells[j].Time)
		})

		// Formatted tells in a slice, for joining into a string
		// later.
		tellsFmtd := []string{}
		for _, tell := range tells {
			tellsFmtd = append(
				tellsFmtd,
				fmt.Sprintf(
					"%s sent you a message %s: %s",
					tell.From, humanize.Time(tell.Time), tell.Message,
				),
			)
		}

		return strings.Join(tellsFmtd, "\n"), &IsPrivateNotice{To: t.To}
	}
}
