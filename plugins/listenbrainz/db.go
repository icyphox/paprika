package listenbrainz

import (
	"fmt"

	"git.icyphox.sh/paprika/database"
)

// Store the Listenbrainz username against the nick.
func Setup(lbzUser, nick string) error {
	err := database.DB.Set(
		[]byte(fmt.Sprintf("lbz/%s", nick)),
		[]byte(lbzUser),
	)
	return err
}

// Gets the Listenbrainz username from the DB.
func GetUser(nick string) (string, error) {
	nick = fmt.Sprintf("lbz/%s", nick)
	user, err := database.DB.Get([]byte(nick))
	if err != nil {
		return "", err
	}
	return string(user), nil
}
