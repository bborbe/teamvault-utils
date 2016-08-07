package generator

import (
	"testing"

	. "github.com/bborbe/assert"
)

func TestImplementsConfigWriter(t *testing.T) {
	object := New()
	var expected *ConfigWriter
	if err := AssertThat(object, Implements(expected)); err != nil {
		t.Fatal(err)
	}
}
