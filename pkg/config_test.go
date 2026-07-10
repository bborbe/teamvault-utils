// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	teamvault "github.com/bborbe/teamvault-utils/v5/pkg"
)

var _ = Describe("Config", func() {
	Describe("ParseTeamvaultConfig", func() {
		It("parses timeout from JSON as duration string", func() {
			cfg, err := teamvault.ParseTeamvaultConfig([]byte(
				`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"5s"}`,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Timeout.Duration()).To(Equal(5 * time.Second))
		})

		It("parses absent timeout as zero", func() {
			cfg, err := teamvault.ParseTeamvaultConfig([]byte(
				`{"url":"https://vault.example.com","user":"admin","pass":"pwd"}`,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Timeout.Duration()).To(Equal(time.Duration(0)))
		})

		It("returns error for unparseable timeout", func() {
			_, err := teamvault.ParseTeamvaultConfig([]byte(
				`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"banana"}`,
			))
			Expect(err).To(HaveOccurred())
		})

		It("parses negative timeout without error (factory validates negativity)", func() {
			cfg, err := teamvault.ParseTeamvaultConfig([]byte(
				`{"url":"https://vault.example.com","user":"admin","pass":"pwd","timeout":"-5s"}`,
			))
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Timeout.Duration()).To(Equal(-5 * time.Second))
		})
	})
})
