package generator

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestImplementsConfigWriter(t *testing.T) {
	object := New()
	var expected *ConfigWriter
	err := AssertThat(object, Implements(expected))
	if err != nil {
		t.Fatal(err)
	}
}
