// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"
	"io"

	"github.com/bborbe/errors"
	"github.com/spf13/cobra"

	teamvault "github.com/bborbe/teamvault-utils/v5/pkg/teamvault"
)

// createConfigCommand creates the config parent command.
func createConfigCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration templating commands",
	}
	cmd.AddCommand(createConfigParseCommand(ctx, sf))
	cmd.AddCommand(createConfigGenerateCommand(ctx, sf))
	return cmd
}

// createConfigParseCommand creates the config parse subcommand.
func createConfigParseCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse a configuration template from stdin and render it to stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			content, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return errors.Wrapf(ctx, err, "read stdin failed")
			}
			output, err := teamvault.NewConfigParser(conn).Parse(ctx, content)
			if err != nil {
				return errors.Wrapf(ctx, err, "parse config failed")
			}
			if _, err := cmd.OutOrStdout().Write(output); err != nil {
				return errors.Wrapf(ctx, err, "write output failed")
			}
			return nil
		},
	}
	return cmd
}

// createConfigGenerateCommand creates the config generate subcommand.
func createConfigGenerateCommand(ctx context.Context, sf *sharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate configuration files from a source directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			src, _ := cmd.Flags().GetString("source-dir")
			dst, _ := cmd.Flags().GetString("target-dir")

			conn, err := sf.buildConnector(ctx)
			if err != nil {
				return err
			}
			gen := teamvault.NewConfigGenerator(teamvault.NewConfigParser(conn))
			if err := gen.Generate(ctx, teamvault.SourceDirectory(src), teamvault.TargetDirectory(dst)); err != nil {
				return errors.Wrapf(ctx, err, "generate failed")
			}
			return nil
		},
	}

	cmd.Flags().String("source-dir", "", "source directory")
	cmd.Flags().String("target-dir", "", "target directory")
	_ = cmd.MarkFlagRequired("source-dir")
	_ = cmd.MarkFlagRequired("target-dir")

	return cmd
}
