package dummy

import (
	"os"
	"testing"

	. "github.com/bborbe/assert"
	"github.com/golang/glog"
	"github.com/seibert-media/kubernetes_tools/manifests/model"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestUser(t *testing.T) {
	key := model.TeamvaultKey("key123")
	du := New()
	user, err := du.User(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(user, Is(model.TeamvaultUser("key123"))); err != nil {
		t.Fatal(err)
	}
}

func TestPassword(t *testing.T) {
	key := model.TeamvaultKey("key123")
	du := New()
	password, err := du.Password(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(password, Is(model.TeamvaultPassword("LgIWz7BC2r68P9WTtVJdfFOYrpT2tv_yw95BzhzECiU="))); err != nil {
		t.Fatal(err)
	}
}

func TestFile(t *testing.T) {
	key := model.TeamvaultKey("key123")
	du := New()
	password, err := du.File(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(password, Is(model.TeamvaultPassword("YTJWNU1USXpMV1pwYkdVSwo="))); err != nil {
		t.Fatal(err)
	}
}

func TestURL(t *testing.T) {
	key := model.TeamvaultKey("key123")
	du := New()
	url, err := du.URL(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(url, Is(model.TeamvaultUrl("dk9kTUjDqGcvPlvF0ZOovq3sBE-0_-Y62i8mlTX_g1M="))); err != nil {
		t.Fatal(err)
	}
}
