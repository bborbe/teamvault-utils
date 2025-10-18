// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "context"

// Connector provides access to TeamVault secrets including passwords, users, URLs, and files.
//
//counterfeiter:generate -o  mocks/connector.go --fake-name Connector . Connector
type Connector interface {
	Password(ctx context.Context, key Key) (Password, error)
	User(ctx context.Context, key Key) (User, error)
	Url(ctx context.Context, key Key) (Url, error)
	File(ctx context.Context, key Key) (File, error)
	Search(ctx context.Context, name string) ([]Key, error)
}
