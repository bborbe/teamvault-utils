package model

import (
	"os"
	"testing"

	. "github.com/bborbe/assert"

	"github.com/golang/glog"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestAddress(t *testing.T) {
	address, err := ParseAddress("172.16.60.123/24")
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(address, NotNilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(address.String(), Is("172.16.60.123/24")); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(address.Ip.String(), Is("172.16.60.123")); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(address.Mask.String(), Is("24")); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(address.Network(), Is("172.16.60.0/24")); err != nil {
		t.Fatal(err)
	}
}
