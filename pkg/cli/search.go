// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bborbe/errors"
	"github.com/spf13/cobra"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

// createSearchCommand creates the `search` subcommand.
func createSearchCommand(ctx context.Context, sf *SharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for secrets by name and print matching keys",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			conn, err := newConnector(sf)(ctx)
			if err != nil {
				return err
			}
			keys, err := conn.Search(ctx, query)
			if err != nil {
				return errors.Wrap(ctx, err, "search failed")
			}
			asJSON, _ := cmd.Flags().GetBool("json")
			return writeSearch(ctx, cmd.OutOrStdout(), keys, asJSON)
		},
	}

	cmd.Flags().Bool("json", false, "print output as a JSON array of keys")
	return cmd
}

// writeSearch writes the search result keys to the given writer. In the default
// mode it writes one key per line. In --json mode it writes a JSON array of keys
// with a trailing newline.
func writeSearch(
	ctx context.Context,
	out io.Writer,
	keys []teamvault.Key,
	asJSON bool,
) error {
	if !asJSON {
		for _, key := range keys {
			if _, err := fmt.Fprintf(out, "%s\n", key.String()); err != nil {
				return errors.Wrapf(ctx, err, "write search result failed")
			}
		}
		return nil
	}
	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		ids = append(ids, k.String())
	}
	encoded, err := json.Marshal(ids)
	if err != nil {
		return errors.Wrapf(ctx, err, "marshal json failed")
	}
	if _, err := fmt.Fprintf(out, "%s\n", encoded); err != nil {
		return errors.Wrapf(ctx, err, "write search result failed")
	}
	return nil
}
