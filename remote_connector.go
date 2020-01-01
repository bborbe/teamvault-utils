package teamvault

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/bborbe/http/header"
	"github.com/bborbe/http/rest"
)

type remoteConnector struct {
	url  Url
	user User
	pass Password
	rest rest.Rest
}

func NewRemoteConnector(
	executeRequest func(req *http.Request) (resp *http.Response, err error),
	url Url,
	user User,
	pass Password,
) Connector {
	return &remoteConnector{
		rest: rest.New(executeRequest),
		url:  url,
		user: user,
		pass: pass,
	}
}

func (t *remoteConnector) Password(ctx context.Context, key Key) (Password, error) {
	currentRevision, err := t.CurrentRevision(ctx, key)
	if err != nil {
		return "", err
	}
	var response struct {
		Password Password `json:"password"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", currentRevision.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Password, nil
}

func (t *remoteConnector) User(ctx context.Context, key Key) (User, error) {
	var response struct {
		User User `json:"username"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.User, nil
}

func (t *remoteConnector) Url(ctx context.Context, key Key) (Url, error) {
	var response struct {
		Url Url `json:"url"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.Url, nil
}

func (t *remoteConnector) CurrentRevision(ctx context.Context, key Key) (TeamvaultCurrentRevision, error) {
	var response struct {
		CurrentRevision TeamvaultCurrentRevision `json:"current_revision"`
	}
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/%s/", t.url.String(), key.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.CurrentRevision, nil
}

func (t *remoteConnector) File(ctx context.Context, key Key) (File, error) {
	rev, err := t.CurrentRevision(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get current revision failed: %v", err)
	}
	var response struct {
		File File `json:"file"`
	}
	if err := t.rest.Call(fmt.Sprintf("%sdata", rev.String()), nil, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return "", err
	}
	return response.File, nil
}

func (t *remoteConnector) createHeader() http.Header {
	httpHeader := make(http.Header)
	httpHeader.Add("Authorization", fmt.Sprintf("Basic %s", header.CreateAuthorizationToken(t.user.String(), t.pass.String())))
	httpHeader.Add("Content-Type", "application/json")
	return httpHeader
}

func (t *remoteConnector) Search(ctx context.Context, search string) ([]Key, error) {
	var response struct {
		Results []struct {
			ApiUrl TeamvaultApiUrl `json:"api_url"`
		} `json:"results"`
	}
	values := url.Values{}
	values.Add("search", search)
	if err := t.rest.Call(fmt.Sprintf("%s/api/secrets/", t.url.String()), values, http.MethodGet, nil, &response, t.createHeader()); err != nil {
		return nil, err
	}
	var result []Key
	for _, re := range response.Results {
		key, err := re.ApiUrl.Key()
		if err != nil {
			return nil, err
		}
		result = append(result, key)
	}
	return result, nil
}
