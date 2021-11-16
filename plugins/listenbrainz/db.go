package listenbrainz

import (
	"fmt"

	"git.icyphox.sh/taigobot/db"
	"github.com/dgraph-io/badger/v3"
)

// Store the Listenbrainz username against the nick.
func Setup(lbzUser, nick string) error {
	err := db.DB.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(fmt.Sprintf("lbz/%s", nick)), []byte(lbzUser))
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// Gets the Listenbrainz username from the DB.
func GetUser(nick string) (string, error) {
	var user string
	err := db.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fmt.Sprintf("lbz/%s", nick)))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			user = string(val)
			return nil
		})
		return nil
	})

	if err != nil {
		return "", err
	}
	return user, nil
}
