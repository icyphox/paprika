package database

import (
	"path"

	"git.icyphox.sh/paprika/config"
	"github.com/dgraph-io/badger/v3"
)

// Use this as a global DB handle.
var DB database

type database struct {
	*badger.DB
}

func Open() (*badger.DB, error) {
	db, err := badger.Open(
		badger.DefaultOptions(path.Join(config.DbPath, "badger")),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Wrapper function to simplify setting a key/val
// in badger.
func (d *database) Set(key, val []byte) error {
	err := d.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		return err
	})

	return err
}

// Wrapper function to simplify getting a key from badger.
func (d *database) Get(key []byte) ([]byte, error) {
	var val []byte
	err := d.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(v []byte) error {
			val, err = item.ValueCopy(nil)
			return err
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}
