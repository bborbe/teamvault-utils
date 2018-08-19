package connector_test

import (
	"testing"

	. "github.com/bborbe/assert"
	"github.com/bborbe/teamvault-utils"
	"github.com/bborbe/teamvault-utils/connector"
)

func TestDiskFallbackConnctorImplementsConnector(t *testing.T) {
	c := &connector.DiskFallback{}
	var i *teamvault.Connector
	if err := AssertThat(c, Implements(i)); err != nil {
		t.Fatal(err)
	}
}
