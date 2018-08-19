package connector_test

import (
	"testing"

	. "github.com/bborbe/assert"
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/connector"
)

func TestDummyUser(t *testing.T) {
	key := teamvault.TeamvaultKey("key123")
	du := connector.NewDummy()
	user, err := du.User(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(user, Is(teamvault.TeamvaultUser("key123"))); err != nil {
		t.Fatal(err)
	}
}

func TestDummyPassword(t *testing.T) {
	key := teamvault.TeamvaultKey("key123")
	du := connector.NewDummy()
	password, err := du.Password(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(password, Is(teamvault.TeamvaultPassword("LgIWz7BC2r68P9WTtVJdfFOYrpT2tv_yw95BzhzECiU="))); err != nil {
		t.Fatal(err)
	}
}

func TestDummyURL(t *testing.T) {
	key := teamvault.TeamvaultKey("key123")
	du := connector.NewDummy()
	url, err := du.Url(key)
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(url, Is(teamvault.TeamvaultUrl("dk9kTUjDqGcvPlvF0ZOovq3sBE-0_-Y62i8mlTX_g1M="))); err != nil {
		t.Fatal(err)
	}
}
