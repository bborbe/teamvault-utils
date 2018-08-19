package connector

import (
	"fmt"
	"net/http"
	"net/url"

	http_header "github.com/bborbe/http/header"
	"github.com/bborbe/http/rest"
	"github.com/bborbe/teamvault-utils"
)

type Remote struct {
	url  teamvault.Url
	user teamvault.User
	pass teamvault.Password
	rest rest.Rest
}

func NewRemote(
	executeRequest func(req *http.Request) (resp *http.Response, err error),
	url teamvault.Url,
	user teamvault.User,
	pass teamvault.Password,
) *Remote {
	t := new(Remote)
	t.rest = rest.New(executeRequest)
	t.url = url
	t.user = user
	t.pass = pass
	return t
}

func (t *Remote) Password(key teamvault.Key) (teamvault.Password, error) {
	currentRevision, err := t.CurrentRevision(key)
	if err != nil {
		return "", err
	}
	var response struct {
		Password teamvault.Password `json:"password"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", currentRevision.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Password, nil
}

func (t *Remote) User(key teamvault.Key) (teamvault.User, error) {
	var response struct {
		User teamvault.User `json:"username"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.User, nil
}

func (t *Remote) Url(key teamvault.Key) (teamvault.Url, error) {
	var response struct {
		Url teamvault.Url `json:"url"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Url, nil
}

func (t *Remote) CurrentRevision(key teamvault.Key) (teamvault.TeamvaultCurrentRevision, error) {
	var response struct {
		CurrentRevision teamvault.TeamvaultCurrentRevision `json:"current_revision"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.CurrentRevision, nil
}

func (t *Remote) File(key teamvault.Key) (teamvault.File, error) {
	rev, err := t.CurrentRevision(key)
	if err != nil {
		return "", fmt.Errorf("get current revision failed: %v", err)
	}
	var response struct {
		File teamvault.File `json:"file"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", rev.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.File, nil
}

func (t *Remote) createHeader() http.Header {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("Basic %s", http_header.CreateAuthorizationToken(t.user.String(), t.pass.String())))
	header.Add("Content-Type", "application/json")
	return header
}

func (t *Remote) Search(search string) ([]teamvault.Key, error) {
	var response struct {
		Results []struct {
			ApiUrl teamvault.TeamvaultApiUrl `json:"api_url"`
		} `json:"results"`
	}
	values := url.Values{}
	values.Add("search", search)
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/", t.url.String()), values, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return nil, err
	}
	var result []teamvault.Key
	for _, re := range response.Results {
		key, err := re.ApiUrl.Key()
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}
	return result, nil
}
