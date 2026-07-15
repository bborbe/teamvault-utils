// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/spf13/cobra"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

// createHtpasswdCommand builds the `htpasswd` subcommand, which prints an
// htpasswd-format credential (user:$2y$... bcrypt) built from a TeamVault
// entry's username + password. This enables secret-free htpasswd generation at
// deploy time, e.g. `--set-string secrets.htpasswd=$(teamvault-cli htpasswd <key>)`,
// so no pre-computed hash needs to live in git.
func createHtpasswdCommand(ctx context.Context, sf *SharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "htpasswd [key]",
		Short: "Print an htpasswd-format credential (user:bcrypt) for a TeamVault secret",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := resolveKey(cmd, args)
			if err != nil {
				return err
			}
			conn, err := newConnector(sf)(ctx)
			if err != nil {
				return errors.Wrap(ctx, err, "create connector failed")
			}
			// Reuse the shared HtpasswdGenerator (same bcrypt logic the
			// teamvaultHtpasswd config template func uses) — no duplicated hashing.
			content, err := teamvault.NewHtpasswdGenerator(conn).Generate(ctx, key)
			if err != nil {
				return errors.Wrap(ctx, err, "generate htpasswd failed")
			}
			// Write the generator's bytes verbatim: they are htpasswd-file
			// format (`user:$2y$...\n`), correct to append to an htpasswd file,
			// and the trailing newline is stripped by `$(...)` command
			// substitution for the --set-string use.
			if _, err := cmd.OutOrStdout().Write(content); err != nil {
				return errors.Wrapf(ctx, err, "write htpasswd failed")
			}
			return nil
		},
	}

	var key string
	cmd.Flags().
		StringVar(&key, "teamvault-key", "", "teamvault key (alternative to positional argument)")

	return cmd
}
