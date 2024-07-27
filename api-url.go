package teamvault

import (
	"fmt"
	"strings"
)

type TeamvaultApiUrl string

func (t TeamvaultApiUrl) String() string {
	return string(t)
}

func (t TeamvaultApiUrl) Key() (Key, error) {
	parts := strings.Split(t.String(), "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("parse key form api-url failed")
	}
	return Key(parts[len(parts)-2]), nil
}
