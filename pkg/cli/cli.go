// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
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
	"syscall"

	"github.com/bborbe/errors"
	libtime "github.com/bborbe/time"
	"github.com/spf13/cobra"

	"github.com/bborbe/teamvault-utils/v5"
	"github.com/bborbe/teamvault-utils/v5/factory"
)

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
		Use:          "teamvault",
		Short:        "TeamVault CLI for retrieving secrets",
		SilenceUsage: true,
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

	rootCmd.AddCommand(createPasswordCommand(ctx, sf))
	rootCmd.AddCommand(createUsernameCommand(ctx, sf))
	rootCmd.AddCommand(createUrlCommand(ctx, sf))
	rootCmd.AddCommand(createFileCommand(ctx, sf))

	return rootCmd
}

// envBool returns true if the environment variable is set to "true", false otherwise.
func envBool(name string) bool {
	return os.Getenv(name) == "true"
}

// createPasswordCommand creates the password subcommand.
func createPasswordCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Retrieve a password from TeamVault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := cmd.Flags().GetString("teamvault-key")
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			result, err := conn.Password(ctx, teamvault.Key(key))
			if err != nil {
				return errors.Wrapf(ctx, err, "get password failed")
			}
			return writeSecret(cmd.OutOrStdout(), result)
		},
	}

	var key string
	cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
	_ = cmd.MarkFlagRequired("teamvault-key")

	return cmd
}

// createUsernameCommand creates the username subcommand.
func createUsernameCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "username",
		Short: "Retrieve a username from TeamVault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := cmd.Flags().GetString("teamvault-key")
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			result, err := conn.User(ctx, teamvault.Key(key))
			if err != nil {
				return errors.Wrapf(ctx, err, "get user failed")
			}
			return writeSecret(cmd.OutOrStdout(), result)
		},
	}

	var key string
	cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
	_ = cmd.MarkFlagRequired("teamvault-key")

	return cmd
}

// createUrlCommand creates the url subcommand.
func createUrlCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url",
		Short: "Retrieve a URL from TeamVault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := cmd.Flags().GetString("teamvault-key")
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			result, err := conn.Url(ctx, teamvault.Key(key))
			if err != nil {
				return errors.Wrapf(ctx, err, "get url failed")
			}
			return writeSecret(cmd.OutOrStdout(), result)
		},
	}

	var key string
	cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
	_ = cmd.MarkFlagRequired("teamvault-key")

	return cmd
}

// createFileCommand creates the file subcommand.
func createFileCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Retrieve a file from TeamVault",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			key, _ := cmd.Flags().GetString("teamvault-key")
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			result, err := conn.File(ctx, teamvault.Key(key))
			if err != nil {
				return errors.Wrapf(ctx, err, "get file failed")
			}
			return writeSecret(cmd.OutOrStdout(), result)
		},
	}

	var key string
	cmd.Flags().StringVar(&key, "teamvault-key", "", "teamvault key")
	_ = cmd.MarkFlagRequired("teamvault-key")

	return cmd
}

// writeSecret writes the secret value to the given writer with no trailing newline.
// This ensures curl -u style basic-auth usage works correctly.
func writeSecret(out io.Writer, value fmt.Stringer) error {
	_, err := fmt.Fprintf(out, "%v", value)
	return err
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
