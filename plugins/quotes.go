package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"git.icyphox.sh/paprika/database"
	"github.com/dgraph-io/badger/v3"
	"gopkg.in/irc.v3"
)

type Quotes struct{}

func init() {
	Register(Quotes{})
}

func (Quotes) Triggers() []string {
	return []string{".q", ".quote"}
}

func result(quoteNum int, total int, nick string, quote string) string {
	return fmt.Sprintf("[%d/%d] <%s> %s", quoteNum, total, nick, quote)
}

var found = errors.New("Found")
var KeyEncodingError = errors.New("Unexpected key encoding")
var TooManyQuotes = errors.New("Too many quotes")

func getQuoteTotal(txn *badger.Txn, keyPrefix []byte) (int, error) {
	item, err := txn.Get([]byte(keyPrefix))
	if err != nil {
		return 0, err
	}
	it, err := item.ValueCopy(nil)
	if err != nil {
		return 0, err
	}
	res, err := strconv.Atoi(string(it))
	if _, ok := err.(*strconv.NumError); ok {
		log.Printf("quotes.go: Warning: Something is wrong with the value in key: %s", keyPrefix)
		return 0, nil // return 0 in hopes of it being overwritten
	} else if err != nil {
		return 0, err
	}
	return res, nil
}

func getAndIncrementQuoteTotal(txn *badger.Txn, keyPrefix []byte) (int, error) {
	total, err := getQuoteTotal(txn, keyPrefix)
	if err == badger.ErrKeyNotFound {
		total = 0
	} else if err != nil {
		return 0, err
	}
	total++
	return total, txn.Set(keyPrefix, []byte(strconv.Itoa(total)))
}

func findQuotes(nick string, keyPrefix, search []byte) (string, error) {
	var (
		num   int
		total int
		quote string
	)

	err := database.DB.DB.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		var err error
		prefix := append(keyPrefix, ' ')
		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			item := iter.Item()
			key := item.Key()
			err = item.Value(func(val []byte) error {
				if bytes.Contains(val, search) {
					quote = string(val)
					keys := bytes.SplitN(key, []byte{' '}, 3)
					if len(keys) != 3 {
						log.Printf("quotes.go: Key Error: %s is not in expected format", key)
						return KeyEncodingError
					}
					num, err = database.DecodeNumber(keys[2])
					if err != nil {
						return err
					} else {
						return found
					}
				} else {
					return nil
				}
			})

			if err != nil {
				break
			}
		}

		if err == found {
			total, err = getQuoteTotal(txn, []byte(keyPrefix))
			if err != nil {
				return err
			} else {
				return found
			}
		} else {
			return err
		}
	})

	if err == nil {
		return "No quote found.", nil
	} else if err == found {
		return result(num, total, nick, quote), nil
	} else {
		return "", err
	}
}

func addQuote(keyPrefix, quote []byte) (string, error) {
	var id int
	err := database.DB.DB.Update(func(txn *badger.Txn) error {
		var err error
		id, err = getAndIncrementQuoteTotal(txn, keyPrefix)
		if err != nil {
			return err
		} else if id > 5_000 {
			return TooManyQuotes
		}

		encodedId, err := database.EncodeNumber(id)
		if err != nil {
			return err
		}

		key := append(keyPrefix, ' ')
		key = append(key, encodedId...)

		err = txn.Set(key, quote)
		if err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		return fmt.Sprintf("Quote %d added.", id), err
	} else {
		return "", err
	}
}

func getQuote(nick string, qnum int, keyPrefix []byte) (string, error) {
	var (
		num      int
		total    int
		quote    string
		negative bool
	)

	if qnum < 0 {
		qnum += 1
		negative = true
	} else {
		negative = false
	}

	err := database.DB.DB.View(func(txn *badger.Txn) error {
		var err error
		total, err = getQuoteTotal(txn, keyPrefix)
		if err != nil {
			return err
		}

		num = qnum
		if negative {
			num = total + qnum
			if num < 1 {
				return badger.ErrKeyNotFound
			}
		} else if num > total {
			return badger.ErrKeyNotFound
		} else if num == randomQuote {
			// [1, total+1)
			num = rand.Intn(total) + 1
		}

		encodeQnum, err := database.EncodeNumber(num)
		if err != nil {
			return err
		}
		encodedKey := append(keyPrefix, ' ')
		encodedKey = append(encodedKey, encodeQnum...)

		qItem, err := txn.Get(encodedKey)
		if err != nil {
			return err
		}
		quoteT, err := qItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		quote = string(quoteT)
		return nil
	})

	if err == badger.ErrKeyNotFound {
		return "No such quote for " + nick, nil
	} else if err != nil {
		return "", err
	} else {
		return result(num, total, nick, quote), nil
	}
}

const randomQuote = 0
const (
	addQ int = iota
	getQ
	start
	parseNick
	parseGetParam
)

func (Quotes) Execute(cmd, rest string, c *irc.Client, m *irc.Message) {
	params := strings.Split(m.Trailing(), " ")
	if len(params) == 1 {
		c.WriteMessage(NewRes(m, ".q [ add ] nickname [ quote | search | number ]"))
		return
	}

	pState := start
	cState := getQ

	var nick string
	keyPrefix := []byte(m.Params[0] + " ")
	for i := 1; i < len(params); i++ {
		word := params[i]
	back:
		if len(word) == 0 {
			continue
		}
		switch pState {
		case start:
			pState = parseNick
			if word == "add" {
				cState = addQ
			} else {
				goto back
			}
		case parseNick:
			if word == "<" || len(word) == 0 {
				break
			}

			// <xyz> -> xyz
			word = strings.TrimPrefix(word, "<")
			word = strings.TrimSuffix(word, ">")
			if len(word) == 0 {
				c.WriteMessage(NewRes(m, "Invalid nickname given"))
				return
			}
			// This is used elsewhere to check the prefix of a "target"
			// if it's true, then this word still has a prefix we can
			// remove.
			if likelyInvalidNickChr(word[0]) {
				word = word[1:]
			}
			if len(word) == 0 {
				c.WriteMessage(NewRes(m, "Invalid nickname given"))
				return
			}
			// we only allow "< " prefix, not "<" + 2*sym
			for j := 0; j < len(word); j++ {
				if likelyInvalidNickChr(word[j]) {
					c.WriteMessage(NewRes(m, fmt.Sprintf("Invalid nickname: %s", word)))
					return
				}
			}
			nick = word
			keyPrefix = append(keyPrefix, nick...)
			if cState == addQ {
				quote := strings.Join(params[i+1:], " ")
				if len(quote) == 0 {
					c.WriteMessage(NewRes(m, "Empty quote not allowed."))
					return
				}
				res, err := addQuote(keyPrefix, []byte(quote))
				if err != nil {
					log.Println(err)
				} else {
					c.WriteMessage(NewRes(m, res))
				}
				return
			} else {
				pState = parseGetParam
			}
		case parseGetParam:
			if i+1 == len(params) {
				qnum, err := strconv.Atoi(word)
				if err != nil {
					res, err := findQuotes(nick, keyPrefix, []byte(word))
					if err != nil {
						log.Println(err)
					} else {
						c.WriteMessage(NewRes(m, res))
					}
					return
				} else {
					res, err := getQuote(nick, qnum, keyPrefix)
					if err != nil {
						log.Println(err)
					} else {
						c.WriteMessage(NewRes(m, res))
					}
					return
				}
			} else {
				quote := strings.Join(params[i+1:], " ")
				res, err := findQuotes(nick, keyPrefix, []byte(quote))
				if err != nil {
					log.Println(err)
				} else {
					c.WriteMessage(NewRes(m, res))
				}
				return
			}
		}
	}
	// If no number given, use 0 to indicate random quote.
	if pState == parseGetParam {
		res, err := getQuote(nick, randomQuote, keyPrefix)
		if err != nil {
			log.Println(err)
		} else {
			c.WriteMessage(NewRes(m, res))
		}
	} else {
		c.WriteMessage(NewRes(m, "Invalid number of parameters."))
	}
}
