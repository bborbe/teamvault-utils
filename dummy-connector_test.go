// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/teamvault-utils/v4"
)

var _ = Describe("", func() {
	var ctx context.Context
	var err error
	var dummyConnector teamvault.Connector
	BeforeEach(func() {
		ctx = context.Background()
		dummyConnector = teamvault.NewDummyConnector()
	})
	Context("User", func() {
		var user teamvault.User
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			user, err = dummyConnector.User(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct user", func() {
			Expect(user).To(Equal(teamvault.User("key123")))
		})
	})
	Context("Password", func() {
		var password teamvault.Password
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			password, err = dummyConnector.Password(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct password", func() {
			Expect(
				password,
			).To(Equal(teamvault.Password("LgIWz7BC2r68P9WTtVJdfFOYrpT2tv_yw95BzhzECiU=")))
		})
	})
	Context("Url", func() {
		var url teamvault.Url
		JustBeforeEach(func() {
			key := teamvault.Key("key123")
			url, err = dummyConnector.Url(ctx, key)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns correct password", func() {
			Expect(url).To(Equal(teamvault.Url("dk9kTUjDqGcvPlvF0ZOovq3sBE-0_-Y62i8mlTX_g1M=")))
		})
	})
})
