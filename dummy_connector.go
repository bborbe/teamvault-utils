package teamvault

import (
	"crypto/sha256"
	"encoding/base64"
)

type dummyConnector struct {
}

func NewDummyConnector() Connector {
	return &dummyConnector{}
}

func (t *dummyConnector) Password(key Key) (Password, error) {
	h := sha256.New()
	h.Write([]byte(key + "-password"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return Password(result), nil
}

func (t *dummyConnector) User(key Key) (User, error) {
	return User(key.String()), nil
}

func (t *dummyConnector) Url(key Key) (Url, error) {
	h := sha256.New()
	h.Write([]byte(key + "-url"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return Url(result), nil
}

func (t *dummyConnector) File(key Key) (File, error) {
	result := base64.URLEncoding.EncodeToString([]byte(key + "-file"))
	return File(result), nil
}

func (t *dummyConnector) Search(search string) ([]Key, error) {
	return nil, nil
}
