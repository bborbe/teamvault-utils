package dummy

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/seibert-media/kubernetes_tools/manifests/model"
)

type dummyPasswordProvider struct {
}

func New() *dummyPasswordProvider {
	t := new(dummyPasswordProvider)
	return t
}

func (t *dummyPasswordProvider) Password(key model.TeamvaultKey) (model.TeamvaultPassword, error) {
	h := sha256.New()
	h.Write([]byte(key + "-password"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return model.TeamvaultPassword(result), nil
}

func (t *dummyPasswordProvider) File(key model.TeamvaultKey) (model.TeamvaultPassword, error) {
	h := sha256.New()
	h.Write([]byte(key + "-file"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return model.TeamvaultPassword(result), nil
}

func (t *dummyPasswordProvider) User(key model.TeamvaultKey) (model.TeamvaultUser, error) {
	return model.TeamvaultUser(key.String()), nil
}

func (t *dummyPasswordProvider) URL(key model.TeamvaultKey) (model.TeamvaultUrl, error) {
	h := sha256.New()
	h.Write([]byte(key + "-url"))
	result := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return model.TeamvaultUrl(result), nil
}
