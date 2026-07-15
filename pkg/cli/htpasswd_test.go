// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli_test

import (
	"bytes"
	"context"
	stderrors "errors"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/cli"
	"github.com/Seibert-Data/teamvault-cli/v5/pkg/mocks"
)

var _ = Describe("htpasswd", func() {
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
		It("NewRootCommand includes the htpasswd subcommand", func() {
			rootCmd := cli.NewRootCommand(ctx)
			subNames := make([]string, len(rootCmd.Commands()))
			for i, c := range rootCmd.Commands() {
				subNames[i] = c.Name()
			}
			Expect(subNames).To(ContainElement("htpasswd"))
		})
	})

	Describe("valid key", func() {
		It("prints a user:bcrypt line that verifies against the password", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.UserReturns(teamvault.User("myuser"), nil)
			fakeConn.PasswordReturns(teamvault.Password("s3cret"), nil)

			var outBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"htpasswd", "ABC123"})
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

			line := strings.TrimRight(outBuf.String(), "\n")
			Expect(line).To(HavePrefix("myuser:"))
			user, hash, found := strings.Cut(line, ":")
			Expect(found).To(BeTrue())
			Expect(user).To(Equal("myuser"))
			Expect(hash).To(HavePrefix("$2"))
			// The emitted hash is a real bcrypt hash of the secret password.
			Expect(bcrypt.CompareHashAndPassword([]byte(hash), []byte("s3cret"))).To(Succeed())
			// Key is resolved and passed through to the connector.
			Expect(fakeConn.PasswordCallCount()).To(Equal(1))
			_, pwKey := fakeConn.PasswordArgsForCall(0)
			Expect(pwKey).To(Equal(teamvault.Key("ABC123")))
		})
	})

	Describe("connector error", func() {
		It("returns an error when the password lookup fails", func() {
			fakeConn := &mocks.Connector{}
			fakeConn.PasswordReturns(teamvault.Password(""), stderrors.New("boom"))

			var errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"htpasswd", "ABC123"})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&errBuf)

			cli.SetNewConnectorForTest(
				func(sf *cli.SharedFlags) func(context.Context) (teamvault.Connector, error) {
					return func(ctx context.Context) (teamvault.Connector, error) {
						return fakeConn, nil
					}
				},
			)
			defer cli.ResetNewConnectorForTest()

			err := cmd.Execute()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("generate htpasswd failed"))
		})
	})

	Describe("missing key", func() {
		It("returns an error when no key is provided", func() {
			var errBuf bytes.Buffer
			cmd := cli.NewRootCommand(ctx)
			cmd.SetArgs([]string{"htpasswd"})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&errBuf)

			err := cmd.Execute()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("teamvault key required"))
		})
	})
})
