// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/bborbe/teamvault-utils/v5/pkg/factory"
	teamvault "github.com/bborbe/teamvault-utils/v5/pkg/teamvault"
)

// connectorFactory creates a TeamVault connector given a context and password.
type connectorFactory func(context.Context, teamvault.Password) (teamvault.Connector, error)

// createLoginCommand creates the login subcommand.
func createLoginCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to TeamVault and store credentials in the macOS Keychain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			kc := teamvault.NewKeychain()

			resolvedURL := teamvault.Url(sf.url)
			resolvedUser := teamvault.User(sf.user)
			initialPass := teamvault.Password(sf.pass)

			configPath := teamvault.TeamvaultConfigPath(sf.configPath)
			if configPath.Exists() {
				config, err := configPath.Parse()
				if err != nil {
					return errors.Wrapf(ctx, err, "parse teamvault config failed")
				}
				resolvedURL = config.Url
				resolvedUser = config.User
				if initialPass == "" {
					initialPass = config.Password
				}
			}

			if resolvedURL == "" {
				return errors.New(
					ctx,
					"teamvault URL is required; use --teamvault-url, TEAMVAULT_URL, or configure in --teamvault-config",
				)
			}

			if resolvedUser == "" {
				return errors.New(
					ctx,
					"teamvault user is required; use --teamvault-user, TEAMVAULT_USER, or configure in --teamvault-config",
				)
			}

			if initialPass == "" {
				pass, err := kc.ReadPassword(ctx, resolvedURL)
				if err != nil {
					return errors.Wrapf(
						ctx,
						err,
						"read keychain password for %s failed",
						resolvedURL,
					)
				}
				initialPass = pass
			}

			httpClient, err := factory.CreateHttpClient(ctx)
			if err != nil {
				return errors.Wrapf(ctx, err, "create httpClient failed")
			}
			timeout := time.Duration(0)
			if sf.timeout != "" {
				d, err := libtime.ParseDuration(ctx, sf.timeout)
				if err != nil {
					return errors.Wrapf(ctx, err, "parse teamvault-timeout %q failed", sf.timeout)
				}
				timeout = d.Duration()
			}
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			httpClient.Timeout = timeout
			currentDateTime := libtime.NewCurrentDateTime()
			staging := teamvault.Staging(sf.staging)

			makeConnector := func(connCtx context.Context, pass teamvault.Password) (teamvault.Connector, error) {
				return factory.CreateConnector(
					httpClient,
					resolvedURL,
					resolvedUser,
					pass,
					staging,
					false,
					currentDateTime,
				), nil
			}

			return loginFlow(
				ctx,
				&termReader{},
				cmd.ErrOrStderr(),
				makeConnector,
				kc,
				resolvedURL,
				resolvedUser,
				initialPass,
			)
		},
	}

	return cmd
}

// loginFlow handles the interactive login flow for TeamVault.
func loginFlow(
	ctx context.Context,
	in io.Reader,
	errOut io.Writer,
	makeConnector connectorFactory,
	kc teamvault.Keychain,
	url teamvault.Url,
	user teamvault.User,
	initialPass teamvault.Password,
) error {
	if initialPass != "" {
		conn, err := makeConnector(ctx, initialPass)
		if err != nil {
			return errors.Wrapf(ctx, err, "create connector for %s failed", url)
		}
		verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err = conn.Search(verifyCtx, "_login_probe_")
		cancel()
		if err == nil {
			return writeAndReport(ctx, errOut, kc, url, initialPass)
		}
		if !isAuthError(err) {
			return errors.Wrapf(ctx, err, "connect to %s failed", url)
		}
	}

	reader := bufio.NewReader(in)
	for attempt := 1; attempt <= 3; attempt++ {
		if ctx.Err() != nil {
			return errors.Wrapf(ctx, ctx.Err(), "login aborted")
		}
		fmt.Fprintf(errOut, "TeamVault password for %s@%s: ", user, url)
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errors.New(ctx, "login aborted")
			}
			return errors.Wrapf(ctx, err, "read password failed")
		}
		typedPass := teamvault.Password(strings.TrimRight(line, "\n\r"))

		conn, err := makeConnector(ctx, typedPass)
		if err != nil {
			return errors.Wrapf(ctx, err, "create connector for %s failed", url)
		}
		verifyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err = conn.Search(verifyCtx, "_login_probe_")
		cancel()
		if err == nil {
			return writeAndReport(ctx, errOut, kc, url, typedPass)
		}
		if !isAuthError(err) {
			return errors.Wrapf(ctx, err, "connect to %s failed", url)
		}
		if attempt < 3 {
			fmt.Fprintln(errOut, "Invalid password, try again.")
		}
	}

	return errors.New(ctx, "login failed: 3 invalid password attempts")
}

// writeAndReport writes the validated password to the keychain and reports status.
func writeAndReport(
	ctx context.Context,
	errOut io.Writer,
	kc teamvault.Keychain,
	url teamvault.Url,
	pass teamvault.Password,
) error {
	if err := kc.WritePassword(ctx, url, pass); err != nil {
		if errors.Is(err, teamvault.ErrKeychainNotSupported) {
			fmt.Fprintln(
				errOut,
				"Login successful. (Keychain storage is macOS-only in v1; password not persisted.)",
			)
			return nil
		}
		return errors.Wrapf(
			ctx,
			err,
			"store password in keychain for %s failed; try unlocking your Keychain",
			url,
		)
	}
	fmt.Fprintf(errOut, "Login successful. Password stored in macOS Keychain for %s.\n", url)
	return nil
}

// isAuthError returns true if the error indicates an authentication failure (401 or 403).
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status: 401") || strings.Contains(msg, "status: 403")
}

// termReader adapts term.ReadPassword to the io.Reader interface so that
// loginFlow can be tested with a plain bytes.Buffer while production code
// reads from the terminal with echo suppressed.
type termReader struct {
	buf []byte
}

func (t *termReader) Read(p []byte) (int, error) {
	if len(t.buf) == 0 {
		pw, err := term.ReadPassword(
			int(os.Stdin.Fd()),
		) // #nosec G115 -- stdin fd is always a small non-negative integer; the conversion is safe
		if err != nil {
			return 0, err
		}
		t.buf = append(pw, '\n')
	}
	n := copy(p, t.buf)
	t.buf = t.buf[n:]
	return n, nil
}
