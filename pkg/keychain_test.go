// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/seibert-media/teamvault-cli/v5/pkg"
	"github.com/seibert-media/teamvault-cli/v5/pkg/mocks"
)

var _ = Describe("Keychain", func() {
	It("KeychainServiceName has expected value", func() {
		Expect(teamvault.KeychainServiceName).To(Equal("teamvault-utils"))
	})

	It("ErrKeychainNotSupported matches itself via errors.Is", func() {
		Expect(
			stderrors.Is(teamvault.ErrKeychainNotSupported, teamvault.ErrKeychainNotSupported),
		).To(BeTrue())
	})

	Describe("Keychain fake", func() {
		var (
			fakeKeychain *mocks.Keychain
			ctx          context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
			fakeKeychain = &mocks.Keychain{}
		})

		It("ReadPassword can return a password", func() {
			fakeKeychain.ReadPasswordReturns(teamvault.Password("secret"), nil)
			pwd, err := fakeKeychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
			Expect(err).NotTo(HaveOccurred())
			Expect(pwd).To(Equal(teamvault.Password("secret")))
			Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(1))
		})

		It("WritePassword can return nil error", func() {
			fakeKeychain.WritePasswordReturns(nil)
			err := fakeKeychain.WritePassword(
				ctx,
				teamvault.Url("https://vault.example.com"),
				teamvault.Password("secret"),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeKeychain.WritePasswordCallCount()).To(Equal(1))
		})
	})
})
