package teamvault_test

import (
	"context"
	"testing"

	. "github.com/bborbe/assert"

	"github.com/bborbe/teamvault-utils"
)

func TestDummyConnctorImplementsConnector(t *testing.T) {
	c := teamvault.NewDummyConnector()
	var i *teamvault.Connector
	if err := AssertThat(c, Implements(i)); err != nil {
		t.Fatal(err)
	}
}

func TestDummyUser(t *testing.T) {
	key := teamvault.Key("key123")
	du := teamvault.NewDummyConnector()
	user, err := du.User(context.Background(), key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(user, Is(teamvault.User("key123"))); err != nil {
		t.Fatal(err)
	}
}

func TestDummyPassword(t *testing.T) {
	key := teamvault.Key("key123")
	du := teamvault.NewDummyConnector()
	password, err := du.Password(context.Background(), key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(password, Is(teamvault.Password("LgIWz7BC2r68P9WTtVJdfFOYrpT2tv_yw95BzhzECiU="))); err != nil {
		t.Fatal(err)
	}
}

func TestDummyURL(t *testing.T) {
	key := teamvault.Key("key123")
	du := teamvault.NewDummyConnector()
	url, err := du.Url(context.Background(), key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(url, Is(teamvault.Url("dk9kTUjDqGcvPlvF0ZOovq3sBE-0_-Y62i8mlTX_g1M="))); err != nil {
		t.Fatal(err)
	}
}
