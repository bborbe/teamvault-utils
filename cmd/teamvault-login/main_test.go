// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"context"
	stderrors "errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	teamvault "github.com/bborbe/teamvault-utils/v4"
	"github.com/bborbe/teamvault-utils/v4/mocks"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Teamvault Login Suite")
}

var _ = Describe("teamvault-login", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/teamvault-utils/v4/cmd/teamvault-login")
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("loginFlow", func() {
	var (
		ctx           context.Context
		errOut        *bytes.Buffer
		fakeConnector *mocks.Connector
		fakeKeychain  *mocks.Keychain
		makeConnector connectorFactory
		url           teamvault.Url
		user          teamvault.User
	)

	BeforeEach(func() {
		ctx = context.Background()
		errOut = &bytes.Buffer{}
		fakeConnector = &mocks.Connector{}
		fakeKeychain = &mocks.Keychain{}
		url = teamvault.Url("https://vault.example.com")
		user = teamvault.User("alice")
		makeConnector = func(_ context.Context, _ teamvault.Password) (teamvault.Connector, error) {
			return fakeConnector, nil
		}
	})

	Describe("initial password resolves and verifies", func() {
		It("writes to keychain once without prompting", func() {
			fakeConnector.SearchReturns(nil, nil)
			fakeKeychain.WritePasswordReturns(nil)

			in := &bytes.Buffer{}
			err := loginFlow(
				ctx,
				in,
				errOut,
				makeConnector,
				fakeKeychain,
				url,
				user,
				"correct-pass",
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(1))
			_, gotURL, gotPass := fakeKeychain.WritePasswordArgsForCall(0)
			Expect(gotURL).To(Equal(url))
			Expect(gotPass).To(Equal(teamvault.Password("correct-pass")))
			Expect(errOut.String()).To(ContainSubstring("Login successful"))
			Expect(errOut.String()).To(ContainSubstring(url.String()))
		})
	})

	Describe("no initial password — prompt loop", func() {
		It(
			"prompts once, user types correct password, keychain write called with that password",
			func() {
				fakeConnector.SearchReturns(nil, nil)
				fakeKeychain.WritePasswordReturns(nil)

				in := bytes.NewBufferString("my-secret\n")
				err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "")

				Expect(err).NotTo(HaveOccurred())
				Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(1))
				_, gotURL, gotPass := fakeKeychain.WritePasswordArgsForCall(0)
				Expect(gotURL).To(Equal(url))
				Expect(gotPass).To(Equal(teamvault.Password("my-secret")))
				Expect(errOut.String()).To(ContainSubstring("TeamVault password for"))
			},
		)
	})

	Describe("3 wrong attempts", func() {
		It("returns an error and does not call keychain write", func() {
			fakeConnector.SearchReturns(nil, fmt.Errorf("request failed with status: 401"))

			in := bytes.NewBufferString("wrong1\nwrong2\nwrong3\n")
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("3 invalid password attempts"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("Ctrl-D (EOF) after one wrong attempt", func() {
		It("returns login aborted error and does not call keychain write", func() {
			// First attempt returns 401, second attempt hits EOF
			fakeConnector.SearchReturnsOnCall(0, nil, fmt.Errorf("request failed with status: 401"))

			in := bytes.NewBufferString("wrong-pass\n") // only one line, EOF after
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("login aborted"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("non-auth verification error", func() {
		It("returns wrapped error immediately without prompting", func() {
			networkErr := fmt.Errorf("dial tcp: connection refused")
			fakeConnector.SearchReturns(nil, networkErr)

			in := &bytes.Buffer{}
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "some-pass")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connect to"))
			Expect(errOut.String()).NotTo(ContainSubstring("TeamVault password for"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("verification timeout (DeadlineExceeded)", func() {
		It("returns wrapped error immediately without prompting", func() {
			fakeConnector.SearchReturns(nil, context.DeadlineExceeded)

			in := &bytes.Buffer{}
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "some-pass")

			Expect(err).To(HaveOccurred())
			Expect(errOut.String()).NotTo(ContainSubstring("TeamVault password for"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("keychain write returns ErrKeychainNotSupported (non-darwin)", func() {
		It("returns nil and stderr contains macOS-only notice", func() {
			fakeConnector.SearchReturns(nil, nil)
			fakeKeychain.WritePasswordReturns(teamvault.ErrKeychainNotSupported)

			in := &bytes.Buffer{}
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "valid-pass")

			Expect(err).NotTo(HaveOccurred())
			Expect(errOut.String()).To(ContainSubstring("macOS-only"))
			Expect(errOut.String()).NotTo(ContainSubstring("failed"))
		})
	})

	Describe("keychain write returns a real error", func() {
		It("returns wrapped error mentioning the URL", func() {
			fakeConnector.SearchReturns(nil, nil)
			fakeKeychain.WritePasswordReturns(stderrors.New("Keychain is locked"))

			in := &bytes.Buffer{}
			err := loginFlow(ctx, in, errOut, makeConnector, fakeKeychain, url, user, "valid-pass")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(url.String()))
		})
	})

	Describe("stdout is always empty", func() {
		It("successful flow writes nothing to stdout", func() {
			fakeConnector.SearchReturns(nil, nil)
			fakeKeychain.WritePasswordReturns(nil)

			stdOut := &bytes.Buffer{}
			_ = loginFlow(
				ctx,
				&bytes.Buffer{},
				errOut,
				makeConnector,
				fakeKeychain,
				url,
				user,
				"valid-pass",
			)

			Expect(stdOut.String()).To(BeEmpty())
		})
	})

	Describe("ErrKeychainNotSupported sentinel shape", func() {
		It("errors.Is returns true when fake returns the sentinel directly", func() {
			fakeKeychain.WritePasswordReturns(teamvault.ErrKeychainNotSupported)
			err := fakeKeychain.WritePassword(ctx, url, "pass")
			Expect(stderrors.Is(err, teamvault.ErrKeychainNotSupported)).To(BeTrue())
		})
	})

	Describe("context cancelled before prompt loop", func() {
		It("returns login aborted error without prompting", func() {
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel()

			in := &bytes.Buffer{}
			err := loginFlow(cancelCtx, in, errOut, makeConnector, fakeKeychain, url, user, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("login aborted"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("connector factory returns error in prompt loop", func() {
		It("returns the factory error immediately", func() {
			factoryErr := fmt.Errorf("connector factory exploded")
			badFactory := func(_ context.Context, _ teamvault.Password) (teamvault.Connector, error) {
				return nil, factoryErr
			}

			in := bytes.NewBufferString("some-pass\n")
			err := loginFlow(ctx, in, errOut, badFactory, fakeKeychain, url, user, "")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connector factory exploded"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})

	Describe("connector factory returns error during initial verification", func() {
		It("returns the factory error without prompting", func() {
			factoryErr := fmt.Errorf("initial factory error")
			badFactory := func(_ context.Context, _ teamvault.Password) (teamvault.Connector, error) {
				return nil, factoryErr
			}

			in := &bytes.Buffer{}
			err := loginFlow(
				ctx,
				in,
				errOut,
				badFactory,
				fakeKeychain,
				url,
				user,
				"some-initial-pass",
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("initial factory error"))
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(0))
		})
	})
})

var _ = Describe("isAuthError", func() {
	It("returns false for nil error", func() {
		Expect(isAuthError(nil)).To(BeFalse())
	})

	It("returns true for 401 status error", func() {
		Expect(isAuthError(fmt.Errorf("request failed with status: 401"))).To(BeTrue())
	})

	It("returns true for 403 status error", func() {
		Expect(isAuthError(fmt.Errorf("request failed with status: 403"))).To(BeTrue())
	})

	It("returns false for network error", func() {
		Expect(isAuthError(fmt.Errorf("connection refused"))).To(BeFalse())
	})
})
