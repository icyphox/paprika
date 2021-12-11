package lastfm

import (
	"fmt"

	"git.icyphox.sh/paprika/database"
)

// Store the Last.fm username against the nick.
func Setup(lfmUser, nick string) error {
	err := database.DB.Set(
		[]byte(fmt.Sprintf("lfm/%s", nick)),
		[]byte(lfmUser),
	)
	return err
}

// Gets the Last.fm username from the DB.
func GetUser(nick string) (string, error) {
	nick = fmt.Sprintf("lfm/%s", nick)
	user, err := database.DB.Get([]byte(nick))
	if err != nil {
		return "", err
	}
	return string(user), nil
}
