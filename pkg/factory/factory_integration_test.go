// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package factory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v5"
	"github.com/bborbe/teamvault-utils/v5/mocks"
	"github.com/bborbe/teamvault-utils/v5/pkg/factory"
)

var _ = Describe("Factory Integration", func() {
	var (
		ctx             context.Context
		fakeKeychain    *mocks.Keychain
		httpClient      *http.Client
		currentDateTime libtime.CurrentDateTime
		stubServer      *httptest.Server
		stubHandler     *slowHandler
		originalHome    string
		tempHome        string
	)

	BeforeEach(func() {
		ctx = context.Background()
		fakeKeychain = &mocks.Keychain{}
		httpClient = &http.Client{}
		currentDateTime = libtime.NewCurrentDateTime()
		stubHandler = &slowHandler{sleepDuration: 2 * time.Second}
		stubServer = httptest.NewServer(stubHandler)
		originalHome = os.Getenv("HOME")
		tempHome = os.TempDir()
	})

	AfterEach(func() {
		stubServer.Close()
		_ = os.Setenv("HOME", originalHome)
	})

	Describe("timeout with disk fallback", func() {
		var cacheKey teamvault.Key

		BeforeEach(func() {
			cacheKey = teamvault.Key("test-key")
		})

		Context("with cache enabled and pre-populated disk cache", func() {
			BeforeEach(func() {
				_ = os.Setenv("HOME", tempHome)
				cacheDir := filepath.Join(tempHome, ".teamvault-cache", cacheKey.String())
				err := os.MkdirAll(cacheDir, 0700)
				Expect(err).NotTo(HaveOccurred())
				cacheFile := filepath.Join(cacheDir, "password")
				err = os.WriteFile(cacheFile, []byte("cached-value"), 0600)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns cached value when server sleeps longer than timeout", func() {
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(""),
					teamvault.Url(stubServer.URL),
					teamvault.User("admin"),
					teamvault.Password("pwd"),
					teamvault.Staging(false),
					true, // cache enabled
					currentDateTime,
					fakeKeychain,
					libtime.Duration(200*time.Millisecond),
				)
				Expect(err).NotTo(HaveOccurred())

				start := time.Now()
				password, err := connector.Password(ctx, cacheKey)
				elapsed := time.Since(start)

				Expect(err).NotTo(HaveOccurred())
				Expect(string(password)).To(Equal("cached-value"))
				// Should complete well under 2s — proves fallback was used
				Expect(elapsed).To(BeNumerically("<", 1*time.Second))
			})
		})

		Context("without cache and server sleeps longer than timeout", func() {
			BeforeEach(func() {
				_ = os.Setenv("HOME", tempHome)
				// Ensure no cache file exists
				cacheDir := filepath.Join(tempHome, ".teamvault-cache", cacheKey.String())
				_ = os.RemoveAll(cacheDir)
			})

			It("returns timeout error when no cache is available", func() {
				connector, err := factory.CreateConnectorWithConfigAndTimeout(
					ctx,
					httpClient,
					teamvault.TeamvaultConfigPath(""),
					teamvault.Url(stubServer.URL),
					teamvault.User("admin"),
					teamvault.Password("pwd"),
					teamvault.Staging(false),
					false, // cache disabled
					currentDateTime,
					fakeKeychain,
					libtime.Duration(200*time.Millisecond),
				)
				Expect(err).NotTo(HaveOccurred())

				start := time.Now()
				_, err = connector.Password(ctx, cacheKey)
				elapsed := time.Since(start)

				Expect(err).To(HaveOccurred())
				// Should fail within ~500ms (timeout fires, no fallback)
				Expect(elapsed).To(BeNumerically("<", 500*time.Millisecond))
			})
		})
	})
})

type slowHandler struct {
	sleepDuration time.Duration
}

func (h *slowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	time.Sleep(h.sleepDuration)
	w.Header().Set("Content-Type", "application/json")
	// Return a valid TeamVault response structure
	_, _ = w.Write([]byte(`{"current_revision":"rev123","password":"live-value"}`))
}
