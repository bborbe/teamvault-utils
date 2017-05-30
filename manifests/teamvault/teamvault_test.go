package teamvault

import (
	"bytes"
	"fmt"
	. "github.com/bborbe/assert"
	"github.com/bborbe/io/reader_nop_close"
	"github.com/seibert-media/kubernetes_tools/manifests/model"
	"github.com/golang/glog"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestPassword(t *testing.T) {
	key := model.TeamvaultKey("key123")
	tv := New(func(req *http.Request) (resp *http.Response, err error) {

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
