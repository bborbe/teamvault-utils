package connector_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	. "github.com/bborbe/assert"
	"github.com/bborbe/io/reader_nop_close"
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/connector"
)

func TestRemoteConnctorImplementsConnector(t *testing.T) {
	c := connector.NewRemote(nil, "", "", "")
	var i *teamvault.Connector
	if err := AssertThat(c, Implements(i)); err != nil {
		t.Fatal(err)
	}
}

func TestTeamvaultPassword(t *testing.T) {
	key := teamvault.Key("key123")
	tv := connector.NewRemote(func(req *http.Request) (resp *http.Response, err error) {

		user, pass, _ := req.BasicAuth()
		if user != "user" && pass != "pass" {
			return &http.Response{StatusCode: 403}, fmt.Errorf("invalid user/pass")
		}

		if req.URL.String() == "http://teamvault.example.com/api/secrets/key123/" {
			return &http.Response{
				StatusCode: 200,
				Body:       reader_nop_close.New(bytes.NewBufferString(`{"current_revision":"https://teamvault.example.com/api/secret-revisions/ref123/"}`)),
			}, nil
		}
		if req.URL.String() == "https://teamvault.example.com/api/secret-revisions/ref123/data" {
			return &http.Response{
				StatusCode: 200,
				Body:       reader_nop_close.New(bytes.NewBufferString(`{"password":"S3CR3T"}`)),
			}, nil
		}
		return &http.Response{StatusCode: 404}, fmt.Errorf("invalid url %v", req.URL.String())
	}, "http://teamvault.example.com", "user", "pass")
	password, err := tv.Password(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(password.String(), Is("S3CR3T")); err != nil {
		t.Fatal(err)
	}
}

func TestTeamvaultUser(t *testing.T) {
	key := teamvault.Key("key123")
	tv := connector.NewRemote(createRequest(`{"username":"user"}`, "http://teamvault.example.com/api/secrets/key123/"), "http://teamvault.example.com", "user", "pass")
	user, err := tv.User(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(user.String(), Is("user")); err != nil {
		t.Fatal(err)
	}
}

func TestTeamvaultUrl(t *testing.T) {
	key := teamvault.Key("key123")
	tv := connector.NewRemote(createRequest(`{"url":"https://example.com"}`, "http://teamvault.example.com/api/secrets/key123/"), "http://teamvault.example.com", "user", "pass")
	url, err := tv.Url(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(url.String(), Is("https://example.com")); err != nil {
		t.Fatal(err)
	}
}

func TestSearch(t *testing.T) {
	tv := connector.NewRemote(createRequest(`{
  "count": 1,
  "next": null,
  "previous": null,
  "results": [
    {
      "access_policy": "request",
      "allowed_groups": [],
      "allowed_users": [],
      "api_url": "https://teamvault.example.com/api/secrets/key123/",
      "content_type": "password",
      "created": "2017-08-21T12:29:53.252282Z",
      "created_by": "skegel",
      "current_revision": "https://teamvault.example.com/api/secret-revisions/rKp1x5/",
      "data_readable": [],
      "description": "",
      "filename": null,
      "last_read": "2017-08-30T08:37:02.189161Z",
      "name": "SearchString",
      "needs_changing_on_leave": true,
      "status": "ok",
      "url": "https://example.com",
      "username": "foo",
      "web_url": "https://teamvault.example.com/secrets/key123"
    }
  ]
}`, "http://teamvault.example.com/api/secrets/?search=searchString"), "http://teamvault.example.com", "user", "pass")
	matches, err := tv.Search("searchString")
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(len(matches), Is(1)); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(matches[0].String(), Is("key123")); err != nil {
		t.Fatal(err)
	}
}

func createRequest(content string, validUrl string) func(req *http.Request) (resp *http.Response, err error) {
	return func(req *http.Request) (resp *http.Response, err error) {

		user, pass, _ := req.BasicAuth()
		if user != "user" && pass != "pass" {
			return &http.Response{StatusCode: 403}, fmt.Errorf("invalid user/pass")
		}

		if req.URL.String() == validUrl {
			return &http.Response{
				StatusCode: 200,
				Body:       reader_nop_close.New(bytes.NewBufferString(content)),
			}, nil
		}
		return &http.Response{StatusCode: 404}, fmt.Errorf("invalid url %v", req.URL.String())
	}
}
