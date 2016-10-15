package config_parser

import (
	"testing"

	. "github.com/bborbe/assert"
	"github.com/golang/glog"
	"os"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}

func TestImplementsConfigParser(t *testing.T) {
	object := New()
	var expected *ConfigParser
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}
