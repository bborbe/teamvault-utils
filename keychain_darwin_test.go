//go:build darwin

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v4"
	"github.com/bborbe/teamvault-utils/v4/mocks"
)

var _ = Describe("DarwinKeychain", func() {
	var (
		ctx          context.Context
		fakeExecutor *mocks.Executor
		keychain     teamvault.Keychain
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeExecutor = &mocks.Executor{}
		keychain = teamvault.NewKeychainWithExecutor(fakeExecutor)
	})

	Describe("ReadPassword", func() {
		Context("when the entry exists (exit 0)", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("mysecret\n", "", 0, nil)
			})

			It("returns the password with trailing newline trimmed", func() {
				pwd, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("mysecret")))
			})

			It("calls security find-generic-password with correct args", func() {
				_, _ = keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(fakeExecutor.RunCallCount()).To(Equal(1))
				_, name, args, stdin := fakeExecutor.RunArgsForCall(0)
				Expect(name).To(Equal("security"))
				Expect(args).To(Equal([]string{
					"find-generic-password", "-s", "teamvault-utils", "-a", "https://vault.example.com", "-w",
				}))
				Expect(stdin).To(BeEmpty())
			})
		})

		Context("when the entry is not found (exit 44)", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns(
					"",
					"security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.",
					44,
					nil,
				)
			})

			It("returns empty password with no error", func() {
				pwd, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("")))
			})
		})

		Context("when stderr matches 'could not be found' at exit 1", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "the item could not be found in the keychain", 1, nil)
			})

			It("returns empty password with no error (miss path)", func() {
				pwd, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("")))
			})
		})

		Context("when the Keychain is locked (exit 36)", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "Keychain could not be unlocked", 36, nil)
			})

			It("returns an error mentioning unlock", func() {
				_, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unlock"))
			})
		})

		Context("when stderr mentions 'user interaction is not allowed'", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "user interaction is not allowed", 36, nil)
			})

			It("returns an error mentioning unlock", func() {
				_, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unlock"))
			})
		})

		Context("when the URL is empty", func() {
			It("returns empty password without calling executor", func() {
				pwd, err := keychain.ReadPassword(ctx, teamvault.Url(""))
				Expect(err).NotTo(HaveOccurred())
				Expect(pwd).To(Equal(teamvault.Password("")))
				Expect(fakeExecutor.RunCallCount()).To(Equal(0))
			})
		})

		Context("when the executor returns an unknown non-zero exit code", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "some unexpected error", 1, nil)
			})

			It("returns a wrapped error", func() {
				_, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the executor returns an error (e.g. binary not found)", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "", 0, stderrors.New("binary not found"))
			})

			It("returns a wrapped error", func() {
				_, err := keychain.ReadPassword(ctx, teamvault.Url("https://vault.example.com"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("binary not found"))
			})
		})

		Context("when context is cancelled", func() {
			It("propagates the cancellation error via executor", func() {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				fakeExecutor.RunReturns("", "", 0, cancelCtx.Err())

				_, err := keychain.ReadPassword(
					cancelCtx,
					teamvault.Url("https://vault.example.com"),
				)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("WritePassword", func() {
		Context("when write succeeds (exit 0)", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "", 0, nil)
			})

			It("returns nil error", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls security add-generic-password with -U flag and password via stdin", func() {
				_ = keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(fakeExecutor.RunCallCount()).To(Equal(1))
				_, name, args, stdin := fakeExecutor.RunArgsForCall(0)
				Expect(name).To(Equal("security"))
				Expect(args).To(Equal([]string{
					"add-generic-password", "-U", "-s", "teamvault-utils", "-a", "https://vault.example.com", "-w",
				}))
				Expect(stdin).To(Equal("mysecret"))
			})
		})

		Context("when write fails with non-zero exit", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "permission denied", 1, nil)
			})

			It("returns a wrapped error", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when executor returns an error", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "", 0, stderrors.New("exec error"))
			})

			It("returns a wrapped error", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("exec error"))
			})
		})

		Context("regression: password never appears in argv", func() {
			It("uses security -i REPL mode with no -w flag in args", func() {
				_ = keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("secret123"),
				)
				_, name, args, stdin := fakeExecutor.RunArgsForCall(0)
				Expect(name).To(Equal("security"))
				Expect(args).To(Equal([]string{"-i"}))
				Expect(
					stdin,
				).To(ContainSubstring("add-generic-password -U -s teamvault-utils -a https://vault.example.com"))
				Expect(stdin).To(ContainSubstring("-w secret123"))
			})

			It("does not pass password as a positional -w argument", func() {
				_ = keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("secret123"),
				)
				_, _, args, _ := fakeExecutor.RunArgsForCall(0)
				Expect(args).NotTo(ContainElement(ContainSubstring("-w")))
			})
		})

		Context("metacharacter passthrough", func() {
			It(
				"passes dollar and backtick characters literally without shell interpretation",
				func() {
					_ = keychain.WritePassword(
						ctx,
						teamvault.Url("https://vault.example.com"),
						teamvault.Password("pass$with`backticks"),
					)
					_, _, _, stdin := fakeExecutor.RunArgsForCall(0)
					Expect(stdin).To(ContainSubstring("pass$with`backticks"))
					Expect(stdin).NotTo(ContainSubstring("$("))
				},
			)
		})

		Context("space and quote bearing password", func() {
			It("correctly quotes password containing spaces and double quotes", func() {
				_ = keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password(`hello "world" foo`),
				)
				_, _, _, stdin := fakeExecutor.RunArgsForCall(0)
				Expect(stdin).To(ContainSubstring(`"hello \"world\" foo"`))
			})
		})

		Context("NUL byte rejection", func() {
			It("rejects password containing NUL byte without calling executor", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("foo\x00bar"),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("NUL"))
				Expect(fakeExecutor.RunCallCount()).To(Equal(0))
			})
		})

		Context("newline rejection", func() {
			It("rejects password containing newline without calling executor", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("foo\nbar"),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("newline"))
				Expect(fakeExecutor.RunCallCount()).To(Equal(0))
			})
		})

		Context("locked Keychain error path", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "Keychain could not be unlocked", 36, nil)
			})

			It("returns an error mentioning Keychain and unlock", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Keychain"))
				Expect(err.Error()).To(ContainSubstring("unlock"))
			})
		})

		Context("successful write", func() {
			BeforeEach(func() {
				fakeExecutor.RunReturns("", "", 0, nil)
			})

			It("returns nil error on success", func() {
				err := keychain.WritePassword(
					ctx,
					teamvault.Url("https://vault.example.com"),
					teamvault.Password("mysecret"),
				)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
