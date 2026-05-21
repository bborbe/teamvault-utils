// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bborbe/errors"
	libservice "github.com/bborbe/service"
	libtime "github.com/bborbe/time"
	"golang.org/x/term"

	teamvault "github.com/bborbe/teamvault-utils/v4"
	"github.com/bborbe/teamvault-utils/v4/factory"
)

func main() {
	app := &application{}
	os.Exit(libservice.MainCmd(context.Background(), app))
}

type application struct {
	TeamvaultUrl        string           `required:"false" arg:"teamvault-url"     env:"TEAMVAULT_URL"     usage:"teamvault url"`
	TeamvaultUser       string           `required:"false" arg:"teamvault-user"    env:"TEAMVAULT_USER"    usage:"teamvault user"`
	TeamvaultPass       string           `required:"false" arg:"teamvault-pass"    env:"TEAMVAULT_PASS"    usage:"teamvault password"                                                          display:"length"`
	TeamvaultConfigPath string           `required:"false" arg:"teamvault-config"  env:"TEAMVAULT_CONFIG"  usage:"teamvault config"`
	Staging             bool             `required:"false" arg:"staging"           env:"STAGING"           usage:"staging status"                                                                               default:"false"`
	TeamvaultTimeout    libtime.Duration `required:"false" arg:"teamvault-timeout" env:"TEAMVAULT_TIMEOUT" usage:"HTTP request timeout for TeamVault API calls (e.g. 5s, 30s); 0 = default 5s"`
}

func (a *application) Run(ctx context.Context) error {
	kc := teamvault.NewKeychain()

	resolvedURL := teamvault.Url(a.TeamvaultUrl)
	resolvedUser := teamvault.User(a.TeamvaultUser)
	initialPass := teamvault.Password(a.TeamvaultPass)

	configPath := teamvault.TeamvaultConfigPath(a.TeamvaultConfigPath)
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
			return errors.Wrapf(ctx, err, "read keychain password for %s failed", resolvedURL)
		}
		initialPass = pass
	}

	httpClient, err := factory.CreateHttpClient(ctx)
	if err != nil {
		return errors.Wrapf(ctx, err, "create httpClient failed")
	}
	timeout := a.TeamvaultTimeout.Duration()
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	httpClient.Timeout = timeout
	currentDateTime := libtime.NewCurrentDateTime()
	staging := teamvault.Staging(a.Staging)

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
		os.Stderr,
		makeConnector,
		kc,
		resolvedURL,
		resolvedUser,
		initialPass,
	)
}

type connectorFactory func(context.Context, teamvault.Password) (teamvault.Connector, error)

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
