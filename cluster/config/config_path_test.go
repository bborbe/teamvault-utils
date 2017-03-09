package config

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

func TestNormalizePath(t *testing.T) {
	var configPath ConfigPath = "/tmp/cluster.config"
	n, err := configPath.NormalizePath()
	if err := AssertThat(err, NilValue()); err != nil {
		t.Fatal(err)
	}
	if err := AssertThat(n, Is(configPath)); err != nil {
		t.Fatal(err)
	}
}
