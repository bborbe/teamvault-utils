package connector

import (
	"fmt"
	"net/http"
	"net/url"

	http_header "github.com/bborbe/http/header"
	"github.com/bborbe/http/rest"
	"github.com/bborbe/teamvault-utils"
)

type teamvaultPasswordProvider struct {
	url  teamvault.TeamvaultUrl
	user teamvault.TeamvaultUser
	pass teamvault.TeamvaultPassword
	rest rest.Rest
}

func New(
	executeRequest func(req *http.Request) (resp *http.Response, err error),
	url teamvault.TeamvaultUrl,
	user teamvault.TeamvaultUser,
	pass teamvault.TeamvaultPassword,
) *teamvaultPasswordProvider {
	t := new(teamvaultPasswordProvider)
	t.rest = rest.New(executeRequest)
	t.url = url
	t.user = user
	t.pass = pass
	return t
}

func (t *teamvaultPasswordProvider) Password(key teamvault.TeamvaultKey) (teamvault.TeamvaultPassword, error) {
	currentRevision, err := t.CurrentRevision(key)
	if err != nil {
		return "", err
	}
	var response struct {
		Password teamvault.TeamvaultPassword `json:"password"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", currentRevision.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Password, nil
}

func (t *teamvaultPasswordProvider) User(key teamvault.TeamvaultKey) (teamvault.TeamvaultUser, error) {
	var response struct {
		User teamvault.TeamvaultUser `json:"username"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.User, nil
}

func (t *teamvaultPasswordProvider) Url(key teamvault.TeamvaultKey) (teamvault.TeamvaultUrl, error) {
	var response struct {
		Url teamvault.TeamvaultUrl `json:"url"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Url, nil
}

func (t *teamvaultPasswordProvider) CurrentRevision(key teamvault.TeamvaultKey) (teamvault.TeamvaultCurrentRevision, error) {
	var response struct {
		CurrentRevision teamvault.TeamvaultCurrentRevision `json:"current_revision"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.CurrentRevision, nil
}

func (t *teamvaultPasswordProvider) File(key teamvault.TeamvaultKey) (teamvault.TeamvaultFile, error) {
	rev, err := t.CurrentRevision(key)
	if err != nil {
		return "", fmt.Errorf("get current revision failed: %v", err)
	}
	var response struct {
		File teamvault.TeamvaultFile `json:"file"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", rev.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.File, nil
}

func (t *teamvaultPasswordProvider) createHeader() http.Header {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("Basic %s", http_header.CreateAuthorizationToken(t.user.String(), t.pass.String())))
	header.Add("Content-Type", "application/json")
	return header
}

func (t *teamvaultPasswordProvider) Search(search string) ([]teamvault.TeamvaultKey, error) {
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
	var result []teamvault.TeamvaultKey
	for _, re := range response.Results {
		key, err := re.ApiUrl.Key()
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}
	return result, nil
}
