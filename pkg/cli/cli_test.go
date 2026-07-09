// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
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

	"github.com/bborbe/teamvault-utils/v5/pkg/cli"
)

var _ = Describe("CLI", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		// Clean up any env vars that might be set
		os.Unsetenv("TEAMVAULT_URL")
		os.Unsetenv("TEAMVAULT_USER")
		os.Unsetenv("TEAMVAULT_PASS")
		os.Unsetenv("TEAMVAULT_CONFIG")
		os.Unsetenv("STAGING")
		os.Unsetenv("TEAMVAULT_TIMEOUT")
		os.Unsetenv("CACHE")
	})

	Describe("env-var seeding", func() {
		DescribeTable(
			"each shared flag falls back to its env var",
			func(envName, flagName, envValue string) {
				os.Setenv(envName, envValue)
				DeferCleanup(func() {
					os.Unsetenv(envName)
				})
				rootCmd := cli.NewRootCommand(ctx)
				flagValue := rootCmd.PersistentFlags().Lookup(flagName).Value.String()
				Expect(flagValue).To(Equal(envValue))
			},
			Entry(
				"TEAMVAULT_URL -> teamvault-url",
				"TEAMVAULT_URL",
				"teamvault-url",
				"https://vault.example.com",
			),
			Entry("TEAMVAULT_USER -> teamvault-user", "TEAMVAULT_USER", "teamvault-user", "admin"),
			Entry(
				"TEAMVAULT_PASS -> teamvault-pass",
				"TEAMVAULT_PASS",
				"teamvault-pass",
				"secretpass",
			),
			Entry(
				"TEAMVAULT_CONFIG -> teamvault-config",
				"TEAMVAULT_CONFIG",
				"teamvault-config",
				"/path/to/config.yaml",
			),
			Entry("STAGING -> staging (true)", "STAGING", "staging", "true"),
			Entry(
				"TEAMVAULT_TIMEOUT -> teamvault-timeout",
				"TEAMVAULT_TIMEOUT",
				"teamvault-timeout",
				"30s",
			),
			Entry("CACHE -> cache (true)", "CACHE", "cache", "true"),
		)
	})

	Describe("flag precedence over env", func() {
		It("explicit flag overrides env default", func() {
			os.Setenv("TEAMVAULT_URL", "from-env")
			DeferCleanup(func() {
				os.Unsetenv("TEAMVAULT_URL")
			})
			rootCmd := cli.NewRootCommand(ctx)
			err := rootCmd.PersistentFlags().Parse([]string{"--teamvault-url=from-flag"})
			Expect(err).To(BeNil())
			Expect(
				rootCmd.PersistentFlags().Lookup("teamvault-url").Value.String(),
			).To(Equal("from-flag"))
		})
	})

	Describe("clean --help", func() {
		It("root --help contains no ginkgo leakage", func() {
			var outBuf, errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"--help"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&errBuf)
			_, _ = cmd.ExecuteC()
			output := outBuf.String() + errBuf.String()
			Expect(strings.ToLower(output)).NotTo(ContainSubstring("ginkgo"))
		})

		It("password --help contains no ginkgo leakage", func() {
			var outBuf, errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"password", "--help"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&errBuf)
			_, _ = cmd.ExecuteC()
			output := outBuf.String() + errBuf.String()
			Expect(strings.ToLower(output)).NotTo(ContainSubstring("ginkgo"))
		})

		It("password --help shows shared persistent flags and teamvault-key", func() {
			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"password", "--help"})
			cmd.SetOut(&outBuf)
			cmd.SetErr(&bytes.Buffer{})
			_, _ = cmd.ExecuteC()
			output := outBuf.String()
			Expect(output).To(ContainSubstring("--teamvault-url"))
			Expect(output).To(ContainSubstring("--teamvault-user"))
			Expect(output).To(ContainSubstring("--teamvault-pass"))
			Expect(output).To(ContainSubstring("--teamvault-config"))
			Expect(output).To(ContainSubstring("--staging"))
			Expect(output).To(ContainSubstring("--teamvault-timeout"))
			Expect(output).To(ContainSubstring("--cache"))
			Expect(output).To(ContainSubstring("--teamvault-key"))
		})
	})

	Describe("missing --teamvault-key", func() {
		It("returns required flag error without calling connector", func() {
			var errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"password"})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&errBuf)
			err := cmd.Execute()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring(`required flag(s) "teamvault-key" not set`))
		})
	})

	Describe("no trailing newline in output", func() {
		It("password subcommand outputs without trailing newline when staging", func() {
			os.Setenv("STAGING", "true")
			DeferCleanup(func() {
				os.Unsetenv("STAGING")
			})
			var buf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"password", "--teamvault-key", "testkey"})
			cmd.SetOut(&buf)
			cmd.SetErr(&bytes.Buffer{})
			_ = cmd.Execute()
			// Staging connector returns a dummy password; verify no trailing newline
			Expect(buf.String()).ToNot(HaveSuffix("\n"))
		})

		It("username subcommand outputs without trailing newline when staging", func() {
			os.Setenv("STAGING", "true")
			DeferCleanup(func() {
				os.Unsetenv("STAGING")
			})
			var buf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"username", "--teamvault-key", "testkey"})
			cmd.SetOut(&buf)
			cmd.SetErr(&bytes.Buffer{})
			_ = cmd.Execute()
			Expect(buf.String()).ToNot(HaveSuffix("\n"))
		})

		It("url subcommand outputs without trailing newline when staging", func() {
			os.Setenv("STAGING", "true")
			DeferCleanup(func() {
				os.Unsetenv("STAGING")
			})
			var buf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"url", "--teamvault-key", "testkey"})
			cmd.SetOut(&buf)
			cmd.SetErr(&bytes.Buffer{})
			_ = cmd.Execute()
			Expect(buf.String()).ToNot(HaveSuffix("\n"))
		})

		It("file subcommand outputs without trailing newline when staging", func() {
			os.Setenv("STAGING", "true")
			DeferCleanup(func() {
				os.Unsetenv("STAGING")
			})
			var buf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"file", "--teamvault-key", "testkey"})
			cmd.SetOut(&buf)
			cmd.SetErr(&bytes.Buffer{})
			_ = cmd.Execute()
			Expect(buf.String()).ToNot(HaveSuffix("\n"))
		})
	})
})
