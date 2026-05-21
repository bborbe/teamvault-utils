//go:build darwin && integration

// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v4"
)

var _ = Describe("DarwinKeychain integration", func() {
	var (
		ctx         context.Context
		keychain    teamvault.Keychain
		serviceName string
		testURL     teamvault.Url
	)

	BeforeEach(func() {
		if _, err := os.Stat("/usr/bin/security"); os.IsNotExist(err) {
			Skip("/usr/bin/security not present — skipping integration test")
		}
		ctx = context.Background()
		keychain = teamvault.NewKeychain()
		// Use a unique service name to avoid clobbering real entries
		serviceName = fmt.Sprintf("teamvault-utils-integration-test-%d", time.Now().UnixNano())
		testURL = teamvault.Url(serviceName)
	})

	AfterEach(func() {
		// Clean up the test entry regardless of test outcome
		_ = keychain.WritePassword(ctx, testURL, teamvault.Password(""))
	})

	It("round-trips a password through write and read", func() {
		const testPwd = teamvault.Password("integration-test-password-12345")

		By("writing the password")
		Expect(keychain.WritePassword(ctx, testURL, testPwd)).To(Succeed())

		By("reading it back")
		got, err := keychain.ReadPassword(ctx, testURL)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(testPwd))
	})

	It("round-trips a password written via the new code path", func() {
		// Uses a unique service-name per test run to avoid clobbering real entries.
		svc := fmt.Sprintf("teamvault-utils-it-%d", time.Now().UnixNano())
		url := teamvault.Url(svc)
		pwd := teamvault.Password("integration,test,password,with,commas,and \"quotes\"")
		DeferCleanup(func() {
			_ = exec.Command("security", "delete-generic-password", "-s", svc, "-a", svc).Run()
		})
		Expect(keychain.WritePassword(ctx, url, pwd)).To(Succeed())
		got, err := keychain.ReadPassword(ctx, url)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(pwd))
	})

	It("returns empty password for a missing entry", func() {
		pwd, err := keychain.ReadPassword(
			ctx,
			teamvault.Url("nonexistent-teamvault-utils-integration-test"),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(pwd).To(BeEmpty())
	})
})
