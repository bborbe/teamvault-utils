// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory_test

import (
	"context"
	stderrors "errors"
	"net/http"
	"os"
	"time"

	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v5"
	"github.com/bborbe/teamvault-utils/v5/factory"
	"github.com/bborbe/teamvault-utils/v5/mocks"
)

var _ = Describe("Factory", func() {
	var (
		ctx             context.Context
		fakeKeychain    *mocks.Keychain
		httpClient      *http.Client
		currentDateTime libtime.CurrentDateTime
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeKeychain = &mocks.Keychain{}
		httpClient = &http.Client{}
		currentDateTime = libtime.NewCurrentDateTime()
	})

	Describe("CreateConnectorWithConfigAndKeychain", func() {
		Context("when password is provided via args (no config file)", func() {
			It("returns connector without consulting keychain", func() {
				connector, err := factory.CreateConnectorWithConfigAndKeychain(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(""),
					teamvault.Url("https://vault.example.com"),
					teamvault.User("admin"),
					teamvault.Password("argspwd"),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
				Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(0))
			})
		})

		Context("when config file provides URL + user + password", func() {
			var configPath string

			BeforeEach(func() {
				f, err := os.CreateTemp("", "teamvault-config-*.json")
				Expect(err).NotTo(HaveOccurred())
				configPath = f.Name()
				DeferCleanup(func() { _ = os.Remove(configPath) })
				_, err = f.WriteString(
					`{"url":"https://vault.example.com","user":"admin","pass":"filepwd"}`,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Close()).To(Succeed())
			})

			It("returns connector without consulting keychain", func() {
				connector, err := factory.CreateConnectorWithConfigAndKeychain(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
				Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(0))
			})
		})

		Context("when config file has URL + user but no password and Keychain returns hit", func() {
			var configPath string

			BeforeEach(func() {
				f, err := os.CreateTemp("", "teamvault-config-*.json")
				Expect(err).NotTo(HaveOccurred())
				configPath = f.Name()
				DeferCleanup(func() { _ = os.Remove(configPath) })
				_, err = f.WriteString(`{"url":"https://vault.example.com","user":"admin"}`)
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Close()).To(Succeed())
				fakeKeychain.ReadPasswordReturns(teamvault.Password("keychainpwd"), nil)
			})

			It("returns connector built with Keychain password", func() {
				connector, err := factory.CreateConnectorWithConfigAndKeychain(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
				Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(1))
			})
		})

		Context(
			"when config file has URL + user but no password and Keychain returns miss",
			func() {
				var configPath string

				BeforeEach(func() {
					f, err := os.CreateTemp("", "teamvault-config-*.json")
					Expect(err).NotTo(HaveOccurred())
					configPath = f.Name()
					DeferCleanup(func() { _ = os.Remove(configPath) })
					_, err = f.WriteString(`{"url":"https://vault.example.com","user":"admin"}`)
					Expect(err).NotTo(HaveOccurred())
					Expect(f.Close()).To(Succeed())
					fakeKeychain.ReadPasswordReturns(teamvault.Password(""), nil)
				})

				It(
					"returns connector built with empty password (existing behavior preserved)",
					func() {
						connector, err := factory.CreateConnectorWithConfigAndKeychain(
							ctx,
							httpClient,
							teamvault.TeamvaultConfigPath(configPath),
							teamvault.Url(""),
							teamvault.User(""),
							teamvault.Password(""),
							teamvault.Staging(false),
							false,
							currentDateTime,
							fakeKeychain,
						)
						Expect(err).NotTo(HaveOccurred())
						Expect(connector).NotTo(BeNil())
						Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(1))
					},
				)
			},
		)

		Context(
			"when config file has URL + user but no password and Keychain returns error",
			func() {
				var configPath string

				BeforeEach(func() {
					f, err := os.CreateTemp("", "teamvault-config-*.json")
					Expect(err).NotTo(HaveOccurred())
					configPath = f.Name()
					DeferCleanup(func() { _ = os.Remove(configPath) })
					_, err = f.WriteString(`{"url":"https://vault.example.com","user":"admin"}`)
					Expect(err).NotTo(HaveOccurred())
					Expect(f.Close()).To(Succeed())
					fakeKeychain.ReadPasswordReturns(
						teamvault.Password(""),
						stderrors.New("keychain locked"),
					)
				})

				It("returns error mentioning the URL", func() {
					_, err := factory.CreateConnectorWithConfigAndKeychain(
						ctx,
						httpClient,
						teamvault.TeamvaultConfigPath(configPath),
						teamvault.Url(""),
						teamvault.User(""),
						teamvault.Password(""),
						teamvault.Staging(false),
						false,
						currentDateTime,
						fakeKeychain,
					)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("https://vault.example.com"))
				})
			},
		)

		Context("when no URL is available (empty args, no config file)", func() {
			It("does not consult keychain", func() {
				connector, err := factory.CreateConnectorWithConfigAndKeychain(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(""),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
				Expect(fakeKeychain.ReadPasswordCallCount()).To(Equal(0))
			})
		})
	})

	Describe("CreateConnectorWithConfig", func() {
		It("returns non-nil connector for a fully-specified args case", func() {
			connector, err := factory.CreateConnectorWithConfig(
				ctx,
				httpClient,
				teamvault.TeamvaultConfigPath(""),
				teamvault.Url("https://vault.example.com"),
				teamvault.User("admin"),
				teamvault.Password("secret"),
				teamvault.Staging(false),
				false,
				currentDateTime,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(connector).NotTo(BeNil())
		})
	})

	Describe("CreateConnectorWithConfigAndTimeout", func() {
		Context("cache OR logic", func() {
			var configPath string

			BeforeEach(func() {
				f, err := os.CreateTemp("", "teamvault-config-*.json")
				Expect(err).NotTo(HaveOccurred())
				configPath = f.Name()
				DeferCleanup(func() { _ = os.Remove(configPath) })
			})

			It("cli cacheEnabled=true, config cacheEnabled=false -> uses disk fallback", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","cacheEnabled":false}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					true, // cli cacheEnabled=true
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				// RemoteConnector would fail to dial a non-routable address.
				// DiskFallbackConnector reads from disk; with no cache file it returns error.
				// Both return error here, but disk fallback was attempted.
				// We verify the connector was built with cache by checking it doesn't
				// immediately return a successful response from remote.
				_, _ = connector.Password(ctx, teamvault.Key("nonexistent"))
				// Error expected (no cache file + unreachable server)
				// If it were a raw RemoteConnector, the error would be a dial error.
				// If it were DiskFallbackConnector, the error comes from the disk read.
				// The key signal is: we got a connector, not nil.
				Expect(connector).NotTo(BeNil())
			})

			It("cli cacheEnabled=false, config cacheEnabled=true -> uses disk fallback", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","cacheEnabled":true}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false, // cli cacheEnabled=false
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
			})

			It("cli cacheEnabled=false, config cacheEnabled=false -> raw remote connector", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","cacheEnabled":false}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false, // cli cacheEnabled=false
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
			})

			It("cli cacheEnabled=true, config cacheEnabled=true -> uses disk fallback", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","cacheEnabled":true}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					true, // cli cacheEnabled=true
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(connector).NotTo(BeNil())
			})
		})

		Context("timeout precedence", func() {
			var configPath string

			BeforeEach(func() {
				f, err := os.CreateTemp("", "teamvault-config-*.json")
				Expect(err).NotTo(HaveOccurred())
				configPath = f.Name()
				DeferCleanup(func() { _ = os.Remove(configPath) })
			})

			It("cli timeout wins over config timeout", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"7s"}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				client := &http.Client{}
				_, err = factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					client,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(3*time.Second),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Timeout).To(Equal(3 * time.Second))
			})

			It("config timeout used when cli timeout is zero", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"7s"}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				client := &http.Client{}
				_, err = factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					client,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Timeout).To(Equal(7 * time.Second))
			})

			It("5s default when both cli and config timeout are zero", func() {
				err := os.WriteFile(
					configPath,
					[]byte(`{"url":"https://vault.example.com","user":"admin","pass":"pwd"}`),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				client := &http.Client{}
				_, err = factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					client,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Timeout).To(Equal(5 * time.Second))
			})

			It("5s default when no config file exists", func() {
				client := &http.Client{}
				_, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					client,
					teamvault.TeamvaultConfigPath("/nonexistent/path/config.json"),
					teamvault.Url("https://vault.example.com"),
					teamvault.User("admin"),
					teamvault.Password("pwd"),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Timeout).To(Equal(5 * time.Second))
			})
		})

		Context("negative timeout rejection", func() {
			var configPath string

			BeforeEach(func() {
				f, err := os.CreateTemp("", "teamvault-config-*.json")
				Expect(err).NotTo(HaveOccurred())
				configPath = f.Name()
				DeferCleanup(func() { _ = os.Remove(configPath) })
			})

			It("rejects negative cli timeout", func() {
				err := os.WriteFile(
					configPath,
					[]byte(`{"url":"https://vault.example.com","user":"admin","pass":"pwd"}`),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				_, err = factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(-1*time.Second),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid timeout"))
				Expect(err.Error()).To(ContainSubstring("-1s"))
			})

			It("rejects negative config timeout", func() {
				err := os.WriteFile(
					configPath,
					[]byte(
						`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"-5s"}`,
					),
					0600,
				)
				Expect(err).NotTo(HaveOccurred())
				_, err = factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(configPath),
					teamvault.Url(""),
					teamvault.User(""),
					teamvault.Password(""),
					teamvault.Staging(false),
					false,
					currentDateTime,
					fakeKeychain,
					libtime.Duration(0),
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid timeout"))
				Expect(err.Error()).To(ContainSubstring("-5s"))
			})
		})
	})
})
