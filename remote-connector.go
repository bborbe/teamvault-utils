// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bborbe/time"
	"github.com/golang/glog"
)

// Url represents a TeamVault URL or secret URL value.
type Url string

// String returns the string representation of the Url.
func (u Url) String() string {
	return string(u)
}

// User represents a TeamVault username.
type User string

// String returns the string representation of the User.
func (u User) String() string {
	return string(u)
}

// UnmarshalJSON implements json.Unmarshaler to handle both string and number types.
func (u *User) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*u = User(str)
		return nil
	}

	// If that fails, try as number and convert to string
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*u = User(num.String())
		return nil
	}

	return fmt.Errorf("username must be string or number")
}

// Password represents a TeamVault password.
type Password string

// String returns the string representation of the Password.
func (t Password) String() string {
	return string(t)
}

// UnmarshalJSON implements json.Unmarshaler to handle both string and number types.
func (t *Password) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*t = Password(str)
		return nil
	}

	// If that fails, try as number and convert to string
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		*t = Password(num.String())
		return nil
	}

	return fmt.Errorf("password must be string or number")
}

// NewRemoteConnector creates a new Connector that connects to a remote TeamVault instance.
func NewRemoteConnector(
	httpClient *http.Client,
	url Url,
	user User,
	pass Password,
	currentDateTime time.CurrentDateTime,
) Connector {
	return &remoteConnector{
		httpClient:      httpClient,
		url:             url,
		user:            user,
		pass:            pass,
		currentDateTime: currentDateTime,
	}
}

type remoteConnector struct {
	url             Url
	user            User
	pass            Password
	httpClient      *http.Client
	currentDateTime time.CurrentDateTime
}

func (r *remoteConnector) Password(ctx context.Context, key Key) (Password, error) {
	currentRevision, err := r.CurrentRevision(ctx, key)
	if err != nil {
		return "", err
	}
	var response struct {
		Password Password `json:"password"`
	}
	if err := r.call(ctx, fmt.Sprintf("%sdata", currentRevision.String()), nil, http.MethodGet, nil, &response, r.createHeader()); err != nil {
		return "", err
	}
	return response.Password, nil
}

func (r *remoteConnector) User(ctx context.Context, key Key) (User, error) {
	var response struct {
		User User `json:"username"`
	}
	if err := r.call(ctx, fmt.Sprintf("%s/api/secrets/%s/", r.url.String(), key.String()), nil, http.MethodGet, nil, &response, r.createHeader()); err != nil {
		return "", err
	}
	return response.User, nil
}

func (r *remoteConnector) Url(ctx context.Context, key Key) (Url, error) {
	var response struct {
		Url Url `json:"url"`
	}
	if err := r.call(ctx, fmt.Sprintf("%s/api/secrets/%s/", r.url.String(), key.String()), nil, http.MethodGet, nil, &response, r.createHeader()); err != nil {
		return "", err
	}
	return response.Url, nil
}

func (r *remoteConnector) CurrentRevision(ctx context.Context, key Key) (CurrentRevision, error) {
	var response struct {
		CurrentRevision CurrentRevision `json:"current_revision"`
	}
	if err := r.call(ctx, fmt.Sprintf("%s/api/secrets/%s/", r.url.String(), key.String()), nil, http.MethodGet, nil, &response, r.createHeader()); err != nil {
		return "", err
	}
	return response.CurrentRevision, nil
}

func (r *remoteConnector) File(ctx context.Context, key Key) (File, error) {
	rev, err := r.CurrentRevision(ctx, key)
	if err != nil {
		return "", fmt.Errorf("get current revision failed: %v", err)
	}
	var response struct {
		File File `json:"file"`
	}
	if err := r.call(ctx, fmt.Sprintf("%sdata", rev.String()), nil, http.MethodGet, nil, &response, r.createHeader()); err != nil {
		return "", err
	}
	return response.File, nil
}

func (r *remoteConnector) createHeader() http.Header {
	httpHeader := make(http.Header)
	httpHeader.Add(
		"Authorization",
		fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", r.user.String(), r.pass.String())),
			),
		),
	)
	httpHeader.Add("Content-Type", "application/json")
	return httpHeader
}

func (r *remoteConnector) Search(ctx context.Context, search string) ([]Key, error) {
	var response struct {
		Results []struct {
			ApiUrl ApiUrl `json:"api_url"`
		} `json:"results"`
	}
	values := url.Values{}
	values.Add("search", search)
	if err := r.call(ctx, fmt.Sprintf("%s/api/secrets/", r.url.String()), values, http.MethodGet, nil, &response, r.createHeader()); err != nil {
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

func (r *remoteConnector) call(
	ctx context.Context,
	url string,
	values url.Values,
	method string,
	request interface{},
	response interface{},
	headers http.Header,
) error {
	if values != nil {
		url = fmt.Sprintf("%s?%s", url, values.Encode())
	}
	glog.V(4).Infof("rest %s to %s", method, url)
	start := r.currentDateTime.Now()
	defer glog.V(8).
		Infof("create completed in %dms", r.currentDateTime.Now().Sub(start)/time.Millisecond)
	glog.V(8).Infof("send message to %s", url)

	var body io.Reader
	if request != nil {
		content, err := json.Marshal(request)
		if err != nil {
			glog.V(2).Infof("marhal request failed: %v", err)
			return err
		}
		if glog.V(8) {
			glog.Infof("send request to %s: %s", url, string(content))
		}
		body = bytes.NewBuffer(content)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		glog.V(2).Infof("build request failed: %v", err)
		return err
	}
	req.Header.Set("ContentType", "application/json")
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := r.httpClient.Do(
		req,
	) // #nosec G704 -- URLs are constructed from configured base URL and API paths, not user input
	if err != nil {
		glog.V(2).Infof("execute request failed: %v", err)
		return err
	}
	if resp.StatusCode/100 != 2 {
		glog.V(2).Infof("request to %s failed with status: %d", url, resp.StatusCode)
		return fmt.Errorf("request to %s failed with status: %d", url, resp.StatusCode)
	}
	if response != nil {
		if err = json.NewDecoder(resp.Body).Decode(response); err != nil {
			glog.V(2).Infof("decode response failed: %v", err)
			return err
		}
	}
	glog.V(8).Infof("rest call successful")
	return nil
}
