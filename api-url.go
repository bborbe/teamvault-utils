package teamvault

import (
	"fmt"
	"strings"
)

type ApiUrl string

func (a ApiUrl) String() string {
	return string(a)
}

func (a ApiUrl) Key() (Key, error) {
	parts := strings.Split(a.String(), "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("parse key form api-url failed")
	}
	return Key(parts[len(parts)-2]), nil
}
