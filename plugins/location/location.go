package location

import (
	"fmt"

	"git.icyphox.sh/paprika/database"
)

func SetLocation(loc, nick string) error {
	err := database.DB.Set(
		[]byte(fmt.Sprintf("loc/%s", nick)),
		[]byte(loc),
	)

	return err
}

func GetLocation(nick string) (string, error) {
	loc, err := database.DB.Get([]byte(fmt.Sprintf("loc/%s", nick)))
	if err != nil {
		return "", err
	}

	return string(loc), nil
}
