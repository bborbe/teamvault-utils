// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
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
			},
		)
	})
})
