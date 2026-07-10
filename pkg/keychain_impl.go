//go:build darwin || linux || windows || freebsd || openbsd || dragonfly

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"context"
	"strings"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
	"github.com/zalando/go-keyring"
)

//counterfeiter:generate -o mocks/keyring_client.go --fake-name KeyringClient . KeyringClient

// KeyringClient is the package-private seam over zalando/go-keyring used by
// darwinKeychain. It exists so unit tests can drive WritePassword/ReadPassword
// without touching the real macOS Keychain. NewKeychain wires up the real
// implementation; tests construct darwinKeychain with a Counterfeiter fake.
type KeyringClient interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
}

type RealKeyringClient struct{}

func (RealKeyringClient) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (RealKeyringClient) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

// NewKeychain returns a Keychain backed by the OS credential store.
// On macOS uses Keychain, on Linux uses Secret Service, on Windows uses Credential Manager.
// On platforms without a supported backend, ReadPassword returns ("", nil) for missing entries
// and Read/WritePassword return ErrKeychainNotSupported for no-backend errors.
func NewKeychain() Keychain {
	return NewKeychainWithClient(RealKeyringClient{})
}

// NewKeychainWithClient returns a Keychain using the given KeyringClient.
// Useful for tests that need to inject a fake credential store.
func NewKeychainWithClient(client KeyringClient) Keychain {
	return &darwinKeychain{client: client}
}

type darwinKeychain struct {
	client KeyringClient
}

func (d *darwinKeychain) ReadPassword(ctx context.Context, url Url) (Password, error) {
	if url == "" {
		glog.V(3).Infof("keychain read skipped: empty URL")
		return "", nil
	}
	pwd, err := d.client.Get(KeychainServiceName, string(url))
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			glog.V(3).Infof("keychain miss for url %q", url)
			return "", nil
		}
		if isNoBackendError(err) {
			return "", ErrKeychainNotSupported
		}
		glog.V(2).Infof("keychain read error for url %q: %v", url, err)
		return "", errors.Wrapf(ctx, err, "keychain read failed for url %q", url)
	}
	glog.V(3).Infof("keychain hit for url %q", url)
	return Password(pwd), nil
}

func (d *darwinKeychain) WritePassword(ctx context.Context, url Url, password Password) error {
	if url == "" {
		glog.V(3).Infof("keychain write skipped: empty URL")
		return nil
	}
	if err := validatePasswordForKeychain(ctx, password); err != nil {
		return err
	}
	if err := d.client.Set(KeychainServiceName, string(url), string(password)); err != nil {
		if isNoBackendError(err) {
			return ErrKeychainNotSupported
		}
		glog.V(2).Infof("keychain write error for url %q: %v", url, err)
		return errors.Wrapf(ctx, err, "keychain write failed for url %q", url)
	}
	glog.V(2).Infof("keychain write succeeded for url %q", url)
	return nil
}

// isNoBackendError returns true when err indicates zalando has no usable
// credential backend on this platform (e.g. Linux without Secret Service,
// or an unsupported platform).
func isNoBackendError(err error) bool {
	if err == nil {
		return false
	}
	// ErrUnsupportedPlatform is returned by zalando's fallback provider
	// for unsupported operating systems.
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return true
	}
	// Linux without dbus/Secret Service: zalano returns an exec.Error
	// when dbus-launch is not found.
	errMsg := err.Error()
	return strings.Contains(errMsg, "dbus-launch") ||
		strings.Contains(errMsg, "unsupported platform")
}

func validatePasswordForKeychain(ctx context.Context, password Password) error {
	for i, ch := range password {
		if ch == 0 {
			return errors.Errorf(
				ctx,
				"password contains NUL byte at position %d which is not supported by Keychain",
				i,
			)
		}
		if ch == '\n' {
			return errors.Errorf(
				ctx,
				"password contains newline which is not supported by Keychain",
			)
		}
	}
	return nil
}
