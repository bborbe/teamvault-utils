package connector

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/bborbe/teamvault-utils"
)

type dummyPasswordProvider struct {
}

func NewDummy() *dummyPasswordProvider {
	t := new(dummyPasswordProvider)
	return t
}

func (t *dummyPasswordProvider) Password(key teamvault.TeamvaultKey) (teamvault.TeamvaultPassword, error) {
	h := sha256.New()
	h.Write([]byte(key + "-password"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return teamvault.TeamvaultPassword(result), nil
}

func (t *dummyPasswordProvider) User(key teamvault.TeamvaultKey) (teamvault.TeamvaultUser, error) {
	return teamvault.TeamvaultUser(key.String()), nil
}

func (t *dummyPasswordProvider) Url(key teamvault.TeamvaultKey) (teamvault.TeamvaultUrl, error) {
	h := sha256.New()
	h.Write([]byte(key + "-url"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return teamvault.TeamvaultUrl(result), nil
}

func (t *dummyPasswordProvider) File(key teamvault.TeamvaultKey) (teamvault.TeamvaultFile, error) {
	result := base64.URLEncoding.EncodeToString([]byte(key + "-file"))
	return teamvault.TeamvaultFile(result), nil
}

func (t *dummyPasswordProvider) Search(search string) ([]teamvault.TeamvaultKey, error) {
	return nil, nil
}
