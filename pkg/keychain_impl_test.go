//go:build darwin || linux || windows || freebsd || openbsd || dragonfly

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/zalando/go-keyring"

	teamvault "github.com/seibert-media/teamvault-cli/v5/pkg"
	"github.com/seibert-media/teamvault-cli/v5/pkg/mocks"
)

var _ = Describe("darwinKeychain", func() {
	var (
		ctx         context.Context
		fakeKeyring *mocks.KeyringClient
		kc          teamvault.Keychain
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeKeyring = &mocks.KeyringClient{}
		kc = teamvault.NewKeychainWithClient(fakeKeyring)
	})

	Describe("ReadPassword", func() {
		Context("when the entry exists", func() {
			BeforeEach(func() {
				fakeKeyring.GetReturns("mysecret", nil)
			})

			It("returns the password", func() {
				pwd, err := kc.ReadPassword(ctx, "https://vault.example.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("mysecret")))
			})

			It("calls client.Get with correct args", func() {
				_, _ = kc.ReadPassword(ctx, "https://vault.example.com")
				Expect(fakeKeyring.GetCallCount()).To(Equal(1))
				svc, user := fakeKeyring.GetArgsForCall(0)
				Expect(svc).To(Equal("teamvault-utils"))
				Expect(user).To(Equal("https://vault.example.com"))
			})
		})

		Context("when keyring returns ErrNotFound", func() {
			BeforeEach(func() {
				fakeKeyring.GetReturns("", keyring.ErrNotFound)
			})

			It("returns empty password with no error", func() {
				pwd, err := kc.ReadPassword(ctx, "https://vault.example.com")
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("")))
			})
		})

		Context("when keyring returns other error", func() {
			BeforeEach(func() {
				fakeKeyring.GetReturns("", stderrors.New("locked"))
			})

			It("returns a wrapped error", func() {
				_, err := kc.ReadPassword(ctx, "https://vault.example.com")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("locked"))
			})
		})

		Context("when keyring returns no-backend error", func() {
			BeforeEach(func() {
				fakeKeyring.GetReturns("", keyring.ErrUnsupportedPlatform)
			})

			It("returns teamvault.ErrKeychainNotSupported", func() {
				_, err := kc.ReadPassword(ctx, "https://vault.example.com")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(teamvault.ErrKeychainNotSupported))
			})
		})

		Context("when URL is empty", func() {
			It("returns empty password without calling client", func() {
				pwd, err := kc.ReadPassword(ctx, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("")))
				Expect(fakeKeyring.GetCallCount()).To(Equal(0))
			})
		})
	})

	Describe("WritePassword", func() {
		Context("when write succeeds", func() {
			BeforeEach(func() {
				fakeKeyring.SetReturns(nil)
			})

			It("returns nil error", func() {
				err := kc.WritePassword(ctx, "https://vault.example.com", "mysecret")
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls client.Set with correct args", func() {
				_ = kc.WritePassword(ctx, "https://vault.example.com", "mysecret")
				Expect(fakeKeyring.SetCallCount()).To(Equal(1))
				svc, user, pwd := fakeKeyring.SetArgsForCall(0)
				Expect(svc).To(Equal("teamvault-utils"))
				Expect(user).To(Equal("https://vault.example.com"))
				Expect(pwd).To(Equal("mysecret"))
			})
		})

		Context("when client.Set returns error", func() {
			BeforeEach(func() {
				fakeKeyring.SetReturns(stderrors.New("locked"))
			})

			It("returns a wrapped error", func() {
				err := kc.WritePassword(ctx, "https://vault.example.com", "mysecret")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("locked"))
			})
		})

		Context("when client.Set returns no-backend error", func() {
			BeforeEach(func() {
				fakeKeyring.SetReturns(keyring.ErrUnsupportedPlatform)
			})

			It("returns teamvault.ErrKeychainNotSupported", func() {
				err := kc.WritePassword(ctx, "https://vault.example.com", "mysecret")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(teamvault.ErrKeychainNotSupported))
			})
		})

		Context("when URL is empty", func() {
			It("returns nil without calling client", func() {
				err := kc.WritePassword(ctx, "", "mysecret")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeKeyring.SetCallCount()).To(Equal(0))
			})
		})

		Context("when password contains NUL byte", func() {
			It("returns error without calling client", func() {
				err := kc.WritePassword(ctx, "https://vault.example.com", "foo\x00bar")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("NUL"))
				Expect(fakeKeyring.SetCallCount()).To(Equal(0))
			})
		})

		Context("when password contains newline", func() {
			It("returns error without calling client", func() {
				err := kc.WritePassword(ctx, "https://vault.example.com", "foo\nbar")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("newline"))
				Expect(fakeKeyring.SetCallCount()).To(Equal(0))
			})
		})
	})
})
