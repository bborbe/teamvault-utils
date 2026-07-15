// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "context"

//counterfeiter:generate -o mocks/connector.go --fake-name Connector . Connector

// Connector provides access to TeamVault secrets including passwords, users, URLs, and files.
type Connector interface {
	Password(ctx context.Context, key Key) (Password, error)
	User(ctx context.Context, key Key) (User, error)
	Url(ctx context.Context, key Key) (Url, error)
	File(ctx context.Context, key Key) (File, error)
	// Search returns the secrets whose name matches name as SearchResult values
	// (key + name + username + url), following the server's pagination up to an
	// internal safety cap. NOTE: returns []SearchResult (was []Key in v5.9.x and
	// earlier) — a breaking change for library consumers.
	Search(ctx context.Context, name string) ([]SearchResult, error)
}
