// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli_test

import (
	"bytes"
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/cli"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/mocks"
)

var _ = Describe("search", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		os.Setenv("STAGING", "true")
		os.Unsetenv("TEAMVAULT_URL")
		os.Unsetenv("TEAMVAULT_USER")
		os.Unsetenv("TEAMVAULT_PASS")
		os.Unsetenv("TEAMVAULT_CONFIG")
		os.Unsetenv("TEAMVAULT_TIMEOUT")
	})

	AfterEach(func() {
		os.Unsetenv("STAGING")
	})

	Describe("command registration", func() {
		It("NewRootCommand includes the search subcommand", func() {
			rootCmd := cli.NewRootCommand(ctx)
			subNames := make([]string, len(rootCmd.Commands()))
			for i, c := range rootCmd.Commands() {
				subNames[i] = c.Name()
			}
			Expect(subNames).To(ContainElement("search"))
		})
	})

	Describe("positional query", func() {
		It("passes the query to Search", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.SearchResult{{Key: "ABC123"}}, nil)

			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "my-query"})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			Expect(fakeConn.SearchCallCount()).To(Equal(1))
			_, query := fakeConn.SearchArgsForCall(0)
			Expect(query).To(Equal("my-query"))
		})
	})

	Describe("output", func() {
		It("prints KEY / NAME table by default", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.SearchResult{
				{Key: "ABC123", Name: "alpha"},
				{Key: "DEF456", Name: "beta"},
			}, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			out := outBuf.String()
			Expect(out).To(ContainSubstring("KEY"))
			Expect(out).To(ContainSubstring("NAME"))
			Expect(out).To(ContainSubstring("ABC123"))
			Expect(out).To(ContainSubstring("alpha"))
			Expect(out).To(ContainSubstring("DEF456"))
			Expect(out).To(ContainSubstring("beta"))
		})

		It("prints bare keys with --keys-only", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.SearchResult{
				{Key: "ABC123", Name: "alpha"},
				{Key: "DEF456", Name: "beta"},
			}, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo", "--keys-only"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			Expect(outBuf.String()).To(Equal("ABC123\nDEF456\n"))
		})

		It("prints JSON array of objects with --json", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.SearchResult{
				{
					Key:      "ABC123",
					Name:     "alpha",
					Username: "user1",
					Url:      teamvault.Url("https://a.example"),
				},
				{
					Key:      "DEF456",
					Name:     "beta",
					Username: "user2",
					Url:      teamvault.Url("https://b.example"),
				},
			}, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo", "--json"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			Expect(strings.TrimSpace(outBuf.String())).To(Equal(
				`[{"key":"ABC123","name":"alpha","username":"user1","url":"https://a.example"},{"key":"DEF456","name":"beta","username":"user2","url":"https://b.example"}]`,
			))
		})

		It("respects --limit", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.SearchResult{
				{Key: "ABC123", Name: "alpha"},
				{Key: "DEF456", Name: "beta"},
			}, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo", "--limit", "1"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			out := outBuf.String()
			Expect(out).To(ContainSubstring("ABC123"))
			Expect(out).To(ContainSubstring("alpha"))
			Expect(out).NotTo(ContainSubstring("DEF456"))
			Expect(out).NotTo(ContainSubstring("beta"))
		})
	})

	Describe("empty results", func() {
		It("exits 0 printing the table header only (no rows) for zero matches", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns(nil, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			err := cmd.Execute()
			Expect(err).To(BeNil())
			// Table mode prints the header even when there are no results.
			Expect(outBuf.String()).To(Equal("KEY  NAME\n"))
		})

		It("prints empty JSON array with --json for zero matches", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns(nil, nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search", "foo", "--json"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			_ = cmd.Execute()
			Expect(strings.TrimSpace(outBuf.String())).To(Equal(`[]`))
		})
	})

	Describe("missing query", func() {
		It("returns an error when no query is provided", func() {
			var errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"search"})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&errBuf)

			err := cmd.Execute()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("accepts 1 arg(s), received 0"))
		})
	})
})
