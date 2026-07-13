// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"fmt"
	"net/http"

	libhttp "github.com/bborbe/http"
	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

var _ = Describe("Url.Normalize", func() {
	DescribeTable(
		"trims trailing slashes and surrounding whitespace",
		func(input teamvault.Url, expected teamvault.Url) {
			Expect(input.Normalize()).To(Equal(expected))
		},
		Entry(
			"already clean",
			teamvault.Url("https://teamvault.seibert.tools"),
			teamvault.Url("https://teamvault.seibert.tools"),
		),
		Entry(
			"single trailing slash",
			teamvault.Url("https://teamvault.seibert.tools/"),
			teamvault.Url("https://teamvault.seibert.tools"),
		),
		Entry(
			"multiple trailing slashes",
			teamvault.Url("https://teamvault.seibert.tools//"),
			teamvault.Url("https://teamvault.seibert.tools"),
		),
		Entry(
			"surrounding whitespace",
			teamvault.Url("  https://teamvault.seibert.tools/  "),
			teamvault.Url("https://teamvault.seibert.tools"),
		),
		Entry("empty stays empty", teamvault.Url(""), teamvault.Url("")),
	)
})

var _ = Describe("RemoteConnector with trailing-slash URL", func() {
	var ctx context.Context
	var server *ghttp.Server
	var remoteConnector teamvault.Connector

	BeforeEach(func() {
		ctx = context.Background()
		server = ghttp.NewServer()
		// Build the connector with a URL that ends in a slash — the value a
		// Seibert user copies from the browser (https://teamvault.seibert.tools/).
		// Without normalization this produces a double-slash path that 404s.
		remoteConnector = teamvault.NewRemoteConnector(
			libhttp.CreateDefaultHttpClient(),
			teamvault.Url(server.URL()+"/"),
			teamvault.User("user"),
			teamvault.Password("pass"),
			libtime.NewCurrentDateTime(),
		)
		server.RouteToHandler(
			http.MethodGet,
			"/api/secrets/key123/",
			func(resp http.ResponseWriter, req *http.Request) {
				resp.WriteHeader(http.StatusOK)
				fmt.Fprintf(resp, `{"username":"myuser"}`)
			},
		)
	})
	AfterEach(func() {
		server.Close()
	})

	It("builds a single-slash request path (no // after the host)", func() {
		result, err := remoteConnector.User(ctx, "key123")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(teamvault.User("myuser")))
	})
})
