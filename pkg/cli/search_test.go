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
			fakeConn.SearchReturns([]teamvault.Key{"ABC123"}, nil)

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
		It("prints one key per line by default", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.Key{"ABC123", "DEF456"}, nil)

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
			Expect(outBuf.String()).To(Equal("ABC123\nDEF456\n"))
		})

		It("prints JSON array with --json", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.SearchReturns([]teamvault.Key{"ABC123", "DEF456"}, nil)

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
			Expect(strings.TrimSpace(outBuf.String())).To(Equal(`["ABC123","DEF456"]`))
		})
	})

	Describe("empty results", func() {
		It("exits 0 with no output for zero matches", func() {
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
			Expect(outBuf.String()).To(Equal(""))
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
