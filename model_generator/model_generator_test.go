package model_generator

import (
	"testing"

	"github.com/golang/glog"
	"os"
)

func TestMain(m *testing.M) {
	exit := m.Run()
	glog.Flush()
	os.Exit(exit)
}
