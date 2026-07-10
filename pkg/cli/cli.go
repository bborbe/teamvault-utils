// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cli provides the command-line interface for the teamvault utility.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/spf13/cobra"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/factory"
)

// version is injected at build time via -ldflags (see the Makefile). For a
// plain `go install …@vX.Y.Z` build (no ldflags) it falls back to the module
// version recorded in the binary's build info, so `--version` still reports the
// real release.
var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

// Execute runs the CLI application. It sets up signal handling for SIGINT and
// SIGTERM, configures structured logging, and executes the root command.
// The sole context.Background() is created here and passed to Run.
func Execute() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
	)

	if err := Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run builds the root command and executes it with the given arguments.
// It returns any error from command execution.
func Run(ctx context.Context, args []string) error {
	rootCmd := NewRootCommand(ctx)
	rootCmd.SetArgs(args)
	return rootCmd.ExecuteContext(ctx)
}

// sharedFlags holds the seven shared CLI flags that apply to all subcommands.
// Each flag falls back to its corresponding environment variable when not set.
type sharedFlags struct {
	url        string
	user       string
	pass       string
	configPath string
	staging    bool
	cache      bool
	timeout    string
}

// NewRootCommand creates the root cobra command with all persistent flags
// and subcommands registered.
func NewRootCommand(ctx context.Context) *cobra.Command {
	sf := &sharedFlags{}
	rootCmd := &cobra.Command{
		Use:           "teamvault-cli",
		Short:         "TeamVault CLI for retrieving secrets",
		Version:       resolveVersion(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&sf.url, "teamvault-url", os.Getenv("TEAMVAULT_URL"), "teamvault url")
	pf.StringVar(&sf.user, "teamvault-user", os.Getenv("TEAMVAULT_USER"), "teamvault user")
	pf.StringVar(&sf.pass, "teamvault-pass", os.Getenv("TEAMVAULT_PASS"), "teamvault password")
	pf.StringVar(
		&sf.configPath,
		"teamvault-config",
		os.Getenv("TEAMVAULT_CONFIG"),
		"teamvault config file path",
	)
	pf.BoolVar(&sf.staging, "staging", envBool("STAGING"), "staging status")
	pf.BoolVar(&sf.cache, "cache", envBool("CACHE"), "enable teamvault secret cache")
	pf.StringVar(
		&sf.timeout,
		"teamvault-timeout",
		os.Getenv("TEAMVAULT_TIMEOUT"),
		"HTTP request timeout for TeamVault API calls (e.g. 5s, 30s); 0 = default 5s",
	)

	rootCmd.AddCommand(createLoginCommand(ctx, sf))
	rootCmd.AddCommand(createSecretCommand(
		ctx,
		sf,
		"password",
		"Retrieve a password from TeamVault",
		"get password failed",
		func(ctx context.Context, conn teamvault.Connector, key teamvault.Key) (fmt.Stringer, error) {
			return conn.Password(ctx, key)
		},
	))
	rootCmd.AddCommand(createSecretCommand(
		ctx,
		sf,
		"username",
		"Retrieve a username from TeamVault",
		"get user failed",
		func(ctx context.Context, conn teamvault.Connector, key teamvault.Key) (fmt.Stringer, error) {
			return conn.User(ctx, key)
		},
	))
	rootCmd.AddCommand(createSecretCommand(
		ctx,
		sf,
		"url",
		"Retrieve a URL from TeamVault",
		"get url failed",
		func(ctx context.Context, conn teamvault.Connector, key teamvault.Key) (fmt.Stringer, error) {
			return conn.Url(ctx, key)
		},
	))
	rootCmd.AddCommand(createSecretCommand(
		ctx,
		sf,
		"file",
		"Retrieve a file from TeamVault",
		"get file failed",
		func(ctx context.Context, conn teamvault.Connector, key teamvault.Key) (fmt.Stringer, error) {
			return conn.File(ctx, key)
		},
	))
	rootCmd.AddCommand(createConfigCommand(ctx, sf))

	return rootCmd
}

// envBool returns true if the environment variable is set to "true", false otherwise.
func envBool(name string) bool {
	return os.Getenv(name) == "true"
}

// createSecretCommand builds a secret-reader subcommand. The four secret
// readers (password/username/url/file) differ only in their Use/Short strings,
// the connector method invoked, and the error message; this helper captures the
// shared wiring (required --teamvault-key, buildConnector, writeSecret).
func createSecretCommand(
	ctx context.Context,
	sf *sharedFlags,
	use, short, errMsg string,
	fetch func(context.Context, teamvault.Connector, teamvault.Key) (fmt.Stringer, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := cmd.Flags().GetString("teamvault-key")
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			result, err := fetch(ctx, conn, teamvault.Key(key))
			if err != nil {
				return errors.Wrap(ctx, err, errMsg)
			}
			return writeSecret(ctx, cmd.OutOrStdout(), result)
		},
	}

	var key string
	cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
	_ = cmd.MarkFlagRequired("teamvault-key")

	return cmd
}

// writeSecret writes the secret value to the given writer with no trailing newline.
// This ensures curl -u style basic-auth usage works correctly.
func writeSecret(ctx context.Context, out io.Writer, value fmt.Stringer) error {
	if _, err := fmt.Fprintf(out, "%v", value); err != nil {
		return errors.Wrapf(ctx, err, "write secret failed")
	}
	return nil
}

// buildConnector creates a TeamVault connector using the shared flags.
func (sf *sharedFlags) buildConnector(ctx context.Context) (teamvault.Connector, error) {
	httpClient, err := factory.CreateHttpClient(ctx)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create httpClient failed")
	}

	var timeout libtime.Duration
	if sf.timeout != "" {
		d, err := libtime.ParseDuration(ctx, sf.timeout)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse teamvault-timeout %q failed", sf.timeout)
		}
		timeout = *d
	}

	conn, err := factory.CreateConnectorWithConfigAndTimeout(
		ctx,
		httpClient,
		teamvault.TeamvaultConfigPath(sf.configPath),
		teamvault.Url(sf.url),
		teamvault.User(sf.user),
		teamvault.Password(sf.pass),
		teamvault.Staging(sf.staging),
		sf.cache,
		libtime.NewCurrentDateTime(),
		teamvault.NewKeychain(),
		timeout,
	)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "create connector failed")
	}
	return conn, nil
}
