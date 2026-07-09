// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"

	"github.com/bborbe/errors"
)

//counterfeiter:generate -o mocks/keychain.go --fake-name Keychain . Keychain

// KeychainServiceName is the constant service name used for all teamvault-utils
// Keychain entries. The account key is the TeamVault URL, which keeps multi-vault
// setups isolated automatically.
const KeychainServiceName = "teamvault-utils"

// ErrKeychainNotSupported indicates the current platform has no supported
// credential store backend. Callers may match this with errors.Is to
// differentiate "no Keychain on this platform" from real Keychain failures.
var ErrKeychainNotSupported = errors.New(
	context.Background(),
	"keychain storage is supported on macOS only in v1",
)

// Keychain reads and writes TeamVault passwords from the OS credential store.
// On macOS it backs onto the login Keychain via the `security(1)` binary.
// On other platforms it is a no-op: ReadPassword returns ("", nil); WritePassword
// returns ErrKeychainNotSupported.
type Keychain interface {
	// ReadPassword returns the password stored for the given TeamVault URL,
	// or ("", nil) if no entry exists. A non-nil error indicates a real
	// failure (Keychain locked, security binary error, etc.) — callers
	// should surface this to the user, not fall through silently.
	ReadPassword(ctx context.Context, url Url) (Password, error)

	// WritePassword stores or overwrites the password for the given URL.
	// On non-darwin platforms it returns ErrKeychainNotSupported.
	WritePassword(ctx context.Context, url Url, password Password) error
}
