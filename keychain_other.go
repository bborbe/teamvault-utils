//go:build !darwin

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import "context"

// NewKeychain returns a no-op Keychain for non-macOS platforms.
func NewKeychain() Keychain {
	return &stubKeychain{}
}

type stubKeychain struct{}

func (s *stubKeychain) ReadPassword(_ context.Context, _ Url) (Password, error) {
	return "", nil
}

func (s *stubKeychain) WritePassword(_ context.Context, _ Url, _ Password) error {
	return ErrKeychainNotSupported
}
