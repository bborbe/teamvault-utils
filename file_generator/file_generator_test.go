package file_generator

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

func TestImplementsConfigWriter(t *testing.T) {
	object := New()
	var expected *ConfigWriter
	if err := AssertThat(object, Implements(expected)); err != nil {
		t.Fatal(err)
	}
}
