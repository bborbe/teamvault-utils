package config_parser

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestImplementsConfigParser(t *testing.T) {
	object := New()
	var expected *ConfigParser
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}
