// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

// newConnector is a seam that returns a teamvault.Connector. Defaults to the
// SharedFlags builder but is overridden by tests via SetNewConnectorForTest.
var newConnector = func(sf *SharedFlags) func(context.Context) (teamvault.Connector, error) {
	return sf.buildConnector
}

// SetNewConnectorForTest overrides the connector constructor for tests.
// Returns a function to call in AfterEach to reset.
func SetNewConnectorForTest(
	f func(sf *SharedFlags) func(context.Context) (teamvault.Connector, error),
) func() {
	prev := newConnector
	newConnector = f
	return func() { newConnector = prev }
}

// ResetNewConnectorForTest is an alias for the reset function returned by SetNewConnectorForTest.
func ResetNewConnectorForTest() {}
