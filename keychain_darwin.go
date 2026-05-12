//go:build darwin

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

// NewKeychain returns a macOS login Keychain-backed Keychain implementation.
func NewKeychain() Keychain {
	return NewKeychainWithExecutor(&osExecutor{})
}

// NewKeychainWithExecutor returns a darwin Keychain using the given Executor.
// This is the dependency-injection constructor used in tests.
func NewKeychainWithExecutor(executor Executor) Keychain {
	return &darwinKeychain{executor: executor}
}

type darwinKeychain struct {
	executor Executor
}

func (d *darwinKeychain) ReadPassword(ctx context.Context, url Url) (Password, error) {
	if url == "" {
		glog.V(3).Infof("keychain read skipped: empty URL")
		return "", nil
	}
	stdout, stderr, exitCode, err := d.executor.Run(
		ctx,
		"security",
		[]string{"find-generic-password", "-s", KeychainServiceName, "-a", string(url), "-w"},
		"",
	)
	if err != nil {
		return "", errors.Wrapf(ctx, err, "execute security command failed")
	}
	if exitCode != 0 {
		if exitCode == 44 || strings.Contains(stderr, "could not be found") {
			glog.V(3).Infof("keychain miss for url %q", url)
			return "", nil
		}
		if exitCode == 36 || strings.Contains(stderr, "could not be unlocked") ||
			strings.Contains(stderr, "user interaction is not allowed") {
			glog.V(2).Infof("keychain locked for url %q", url)
			return "", errors.Errorf(
				ctx,
				"TeamVault password requires Keychain unlock; unlock your Keychain and retry",
			)
		}
		glog.V(2).Infof("keychain error for url %q: exit %d", url, exitCode)
		return "", errors.Errorf(
			ctx,
			"security command failed with exit code %d: %s",
			exitCode,
			stderr,
		)
	}
	glog.V(3).Infof("keychain hit for url %q", url)
	return Password(strings.TrimSuffix(stdout, "\n")), nil
}

func (d *darwinKeychain) WritePassword(ctx context.Context, url Url, password Password) error {
	_, stderr, exitCode, err := d.executor.Run(
		ctx,
		"security",
		[]string{"add-generic-password", "-U", "-s", KeychainServiceName, "-a", string(url), "-w"},
		string(password),
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "execute security command failed")
	}
	if exitCode != 0 {
		return errors.Errorf(
			ctx,
			"security add-generic-password failed with exit code %d: %s",
			exitCode,
			stderr,
		)
	}
	glog.V(2).Infof("keychain write succeeded for url %q", url)
	return nil
}

type osExecutor struct{}

func (o *osExecutor) Run(
	ctx context.Context,
	name string,
	args []string,
	stdin string,
) (string, string, int, error) {
	cmd := exec.CommandContext(
		ctx,
		name,
		args...) // #nosec G204 -- name is always "security" from internal callers
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(stdin)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout.String(), stderr.String(), exitErr.ExitCode(), nil
		}
		return "", "", 0, err
	}
	return stdout.String(), stderr.String(), 0, nil
}
