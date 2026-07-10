// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/mocks"
)

var _ = Describe("config parse", func() {
	var (
		ctx    context.Context
		cmd    *cobra.Command
		outBuf *bytes.Buffer
	)

	BeforeEach(func() {
		ctx = context.Background()
		outBuf = &bytes.Buffer{}
	})

	Describe("round-trip with no placeholders", func() {
		It("writes input unchanged to stdout", func() {
			// A template without [[ placeholders is returned unchanged
			// without the connector being called.
			template := "plain text with no placeholders\n"

			sf := &sharedFlags{
				url:  "https://vault.example.com",
				user: "alice",
				pass: "secret",
			}
			cmd = createConfigParseCommand(ctx, sf)
			cmd.SetArgs([]string{})
			cmd.SetIn(bytes.NewBufferString(template))
			cmd.SetOut(outBuf)
			cmd.SetErr(&bytes.Buffer{})

			err := cmd.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(outBuf.String()).To(Equal(template))
		})
	})

	Describe("config generate required flags", func() {
		It(
			"errors when --source-dir and --target-dir are missing, before any connector call",
			func() {
				sf := &sharedFlags{url: "https://vault.example.com", user: "alice"}
				cmd = createConfigGenerateCommand(ctx, sf)
				cmd.SetArgs([]string{})
				cmd.SetOut(outBuf)
				cmd.SetErr(&bytes.Buffer{})

				err := cmd.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(`required flag(s)`))
				Expect(err.Error()).To(ContainSubstring("source-dir"))
				Expect(err.Error()).To(ContainSubstring("target-dir"))
			},
		)
	})

	Describe("config generate happy path", func() {
		It("renders a template file from source to target using the connector", func() {
			srcDir, err := os.MkdirTemp("", "tvsrc")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = os.RemoveAll(srcDir) })
			dstDir, err := os.MkdirTemp("", "tvdst")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = os.RemoveAll(dstDir) })

			template := `{{ "my-key" | teamvaultPassword }}`
			Expect(
				os.WriteFile(filepath.Join(srcDir, "secret.txt"), []byte(template), 0600),
			).To(Succeed())

			fakeConn := &mocks.Connector{}
			fakeConn.PasswordReturns(teamvault.Password("s3cr3t"), nil)

			gen := teamvault.NewConfigGenerator(teamvault.NewConfigParser(fakeConn))
			err = gen.Generate(
				ctx,
				teamvault.SourceDirectory(srcDir),
				teamvault.TargetDirectory(dstDir),
			)
			Expect(err).NotTo(HaveOccurred())

			got, err := os.ReadFile(filepath.Join(dstDir, "secret.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(got)).To(Equal("s3cr3t"))
			Expect(fakeConn.PasswordCallCount()).To(Equal(1))
		})

		It("generates through the CLI command wiring with a staging connector", func() {
			srcDir, err := os.MkdirTemp("", "tvsrc")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = os.RemoveAll(srcDir) })
			dstDir, err := os.MkdirTemp("", "tvdst")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { _ = os.RemoveAll(dstDir) })

			content := "plain text, no placeholders"
			Expect(
				os.WriteFile(filepath.Join(srcDir, "config.txt"), []byte(content), 0600),
			).To(Succeed())

			sf := &sharedFlags{staging: true}
			cmd = createConfigGenerateCommand(ctx, sf)
			cmd.SetArgs([]string{"--source-dir", srcDir, "--target-dir", dstDir})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			err = cmd.Execute()
			Expect(err).NotTo(HaveOccurred())

			got, err := os.ReadFile(filepath.Join(dstDir, "config.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(got)).To(Equal(content))
		})
	})
})
