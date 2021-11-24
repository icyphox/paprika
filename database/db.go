package database

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"path"
	"strconv"
	"time"

	"git.icyphox.sh/paprika/config"
	"github.com/dgraph-io/badger/v3"
)

// Use this as a global DB handle.
var DB database

type database struct {
	*badger.DB
}

// see: https://dgraph.io/docs/badger/get-started/#garbage-collection
func gc() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := DB.DB.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
	panic("Unreachable!")
}

func Open() (*badger.DB, error) {
	db, err := badger.Open(
		badger.DefaultOptions(path.Join(config.C.DbPath, "badger")),
	)
	if err != nil {
		return nil, err
	}
	go gc()
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

var NumTooBig = errors.New("Number Too Big")
var InvalidNumber = errors.New("Invalid Number")

// encode number so it sorts lexicographically, while being semi readable.
func EncodeNumber(n int) ([]byte, error) {
	neg := false
	num := n
	if n < 0 {
		neg = true
		num = -num
	}

	digits := int(math.Trunc(math.Log10(float64(num))))+1
	if digits > 93 {
		return []byte{}, NumTooBig
	}

	if !neg {
		lenCode := 33 + digits
		return []byte(fmt.Sprintf("%c %d", lenCode, n)), nil
	} else {
		lenCode := 127 - digits
		return []byte(fmt.Sprintf("!%c %d", lenCode, n)), nil
	}
}

// encode number so it sorts lexicographically, while being semi readable.
func DecodeNumber(n []byte) (int, error) {
	if len(n) < 3 {
		return 0, InvalidNumber
	}

	// No digit padding
	if n[0] < 33 || n[0] > 126 {
		return 0, InvalidNumber
	}

	num := bytes.SplitN(n, []byte{' '}, 2)
	if len(num) != 2 {
		return 0, InvalidNumber
	}

	number, err := strconv.Atoi(string(num[1]))
	if err != nil {
		return number, err
	}

	if n[0] == '!' {
		return -number, nil
	} else {
		return number, nil
	}
}
