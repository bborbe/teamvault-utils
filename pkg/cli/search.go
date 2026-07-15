// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/bborbe/errors"
	"github.com/spf13/cobra"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

// createSearchCommand creates the `search` subcommand.
func createSearchCommand(ctx context.Context, sf *SharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for secrets by name and print matching keys and names",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			conn, err := newConnector(sf)(ctx)
			if err != nil {
				return errors.Wrap(ctx, err, "create connector failed")
			}
			results, err := conn.Search(ctx, query)
			if err != nil {
				return errors.Wrap(ctx, err, "search failed")
			}
			asJSON, _ := cmd.Flags().GetBool("json")
			keysOnly, _ := cmd.Flags().GetBool("keys-only")
			// --limit truncates the results client-side after the (paginated,
			// cap-bounded) fetch; it caps output, not server round-trips.
			limit, _ := cmd.Flags().GetInt("limit")
			if limit > 0 && len(results) > limit {
				results = results[:limit]
			}
			return writeSearch(ctx, cmd.OutOrStdout(), results, asJSON, keysOnly)
		},
	}

	cmd.Flags().
		Bool("json", false, "print results as a JSON array of objects {key,name,username,url}")
	cmd.Flags().Bool("keys-only", false, "print only bare keys, one per line (for scripting)")
	cmd.Flags().Int("limit", 0, "maximum number of results to print (0 = no limit)")
	return cmd
}

// writeSearch writes the search results to the given writer.
// When keysOnly is true, scripting mode takes precedence and one bare key per
// line is printed. When asJSON is true, a JSON array of {key,name,username,url}
// objects is written. Otherwise, an aligned KEY / NAME table is printed.
func writeSearch(
	ctx context.Context,
	out io.Writer,
	results []teamvault.SearchResult,
	asJSON bool,
	keysOnly bool,
) error {
	// keysOnly takes precedence over asJSON for scripting.
	if keysOnly {
		for _, r := range results {
			if _, err := fmt.Fprintf(out, "%s\n", r.Key.String()); err != nil {
				return errors.Wrapf(ctx, err, "write search result failed")
			}
		}
		return nil
	}
	if asJSON {
		type searchJSON struct {
			Key      string `json:"key"`
			Name     string `json:"name"`
			Username string `json:"username"`
			Url      string `json:"url"`
		}
		items := make([]searchJSON, 0, len(results))
		for _, r := range results {
			items = append(items, searchJSON{
				Key:      r.Key.String(),
				Name:     r.Name,
				Username: r.Username,
				Url:      r.Url.String(),
			})
		}
		encoded, err := json.Marshal(items)
		if err != nil {
			return errors.Wrapf(ctx, err, "marshal json failed")
		}
		if _, err := fmt.Fprintf(out, "%s\n", encoded); err != nil {
			return errors.Wrapf(ctx, err, "write search result failed")
		}
		return nil
	}
	// Default: aligned KEY / NAME table.
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "KEY\tNAME")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\n", r.Key.String(), r.Name)
	}
	if err := tw.Flush(); err != nil {
		return errors.Wrapf(ctx, err, "flush search table failed")
	}
	return nil
}
