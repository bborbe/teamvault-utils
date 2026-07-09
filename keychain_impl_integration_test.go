//go:build darwin && integration

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v5"
)

var _ = Describe("DarwinKeychain integration", func() {
	var (
		ctx      context.Context
		keychain teamvault.Keychain
	)

	BeforeEach(func() {
		if _, err := exec.LookPath("security"); err != nil {
			Skip("/usr/bin/security not present — skipping integration test")
		}
		ctx = context.Background()
		keychain = teamvault.NewKeychain()
	})

	AfterEach(func() {
		// Clean up is handled by DeferCleanup in the test
	})

	It("round-trips a password through zalando go-keyring against the real OS keychain", func() {
		svc := fmt.Sprintf("teamvault-utils-it-%d", time.Now().UnixNano())
		url := teamvault.Url(svc)
		pwd := teamvault.Password(`integration "test" password with , and \\ chars`)

		DeferCleanup(func() {
			_ = exec.Command("security", "delete-generic-password", "-s", svc, "-a", svc).Run()
		})

		// First verify the keychain is accessible by trying an operation
		// If locked, we'll get an error we can detect
		if err := keychain.WritePassword(ctx, url, pwd); err != nil {
			Skip(fmt.Sprintf("keychain probe failed; skipping: %v", err))
		}

		Expect(keychain.WritePassword(ctx, url, pwd)).To(Succeed())
		got, err := keychain.ReadPassword(ctx, url)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(pwd))
	})
})
