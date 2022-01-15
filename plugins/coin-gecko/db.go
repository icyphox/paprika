package coingecko

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
)

var (
	canary = []byte("coin-gecko/list-up-to-date")
	expire = 72 * time.Hour
)

func upsertCoinList() error {
	coins, err := GetCoinList()
	if err != nil {
		return err
	}
	err = database.DB.DB.Update(func(txn *badger.Txn) error {
		newCanary := badger.
			NewEntry(canary, []byte{}).
			WithTTL(time.Duration(expire))
		err := txn.SetEntry(newCanary)
		if err != nil {
			return err
		}

		for _, coin := range coins {
			var buf bytes.Buffer
			encoder := gob.NewEncoder(&buf)
			err := encoder.Encode(coin)
			if err != nil {
				return err
			}
			c := buf.Bytes()
			err = txn.Set([]byte("coin-gecko/id/"+coin.Id), c)
			if err != nil {
				return err
			}
			err = txn.Set([]byte("coin-gecko/symbol/"+coin.Symbol), c)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func GetCoinId(sym string) (string, error) {
	var ret string
	err := database.DB.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("coin-gecko/symbol/" + sym))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		} else if err != badger.ErrKeyNotFound {
			err = item.Value(func(val []byte) error {
				var coinId CoinId
				decoder := gob.NewDecoder(bytes.NewReader(val))
				err := decoder.Decode(&coinId)
				if err != nil {
					return err
				} else {
					ret = coinId.Id
					return nil
				}
			})
			if err != nil {
				return err
			}
		}

		if ret != "" {
			return nil
		}

		item, err = txn.Get([]byte("coin-gecko/id/" + sym))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		} else if err != badger.ErrKeyNotFound {
			err = item.Value(func(val []byte) error {
				var coinId CoinId
				decoder := gob.NewDecoder(bytes.NewReader(val))
				err := decoder.Decode(&coinId)
				if err != nil {
					return err
				} else {
					ret = coinId.Id
					return nil
				}
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return ret, err
}

func CheckUpdateCoinList() error {
	_, err := database.DB.Get(canary)
	if err == badger.ErrKeyNotFound {
		log.Print("Updating Coin Gecko Coin IDs List... ")
		err = upsertCoinList()
		log.Println("Done.")
	} else if err != nil {
		return err
	}
	return err
}
