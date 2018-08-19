package connector

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/bborbe/teamvault-utils"
)

type Dummy struct {
}

func NewDummy() *Dummy {
	t := new(Dummy)
	return t
}

func (t *Dummy) Password(key teamvault.Key) (teamvault.Password, error) {
	h := sha256.New()
	h.Write([]byte(key + "-password"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return teamvault.Password(result), nil
}

func (t *Dummy) User(key teamvault.Key) (teamvault.User, error) {
	return teamvault.User(key.String()), nil
}

func (t *Dummy) Url(key teamvault.Key) (teamvault.Url, error) {
	h := sha256.New()
	h.Write([]byte(key + "-url"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return teamvault.Url(result), nil
}

func (t *Dummy) File(key teamvault.Key) (teamvault.File, error) {
	result := base64.URLEncoding.EncodeToString([]byte(key + "-file"))
	return teamvault.File(result), nil
}

func (t *Dummy) Search(search string) ([]teamvault.Key, error) {
	return nil, nil
}
