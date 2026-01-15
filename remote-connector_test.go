// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
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

	"github.com/bborbe/teamvault-utils/v4"
)

var _ = Describe("RemoteConnector", func() {
	var ctx context.Context
	var err error
	var remoteConnector teamvault.Connector
	var key teamvault.Key
	var username string
	var password string
	var server *ghttp.Server
	BeforeEach(func() {
		ctx = context.Background()
		server = ghttp.NewServer()
		key = "key123"
		username = "user"
		password = "pass"
		remoteConnector = teamvault.NewRemoteConnector(
			libhttp.CreateDefaultHttpClient(),
			teamvault.Url(server.URL()),
			teamvault.User(username),
			teamvault.Password(password),
			libtime.NewCurrentDateTime(),
		)
	})
	AfterEach(func() {
		server.Close()
	})
	Context("Username", func() {
		var result teamvault.User
		JustBeforeEach(func() {
			result, err = remoteConnector.User(ctx, key)
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/key123/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `{"username":"myuser"}`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns user", func() {
			Expect(result).NotTo(BeNil())
			Expect(result.String()).To(Equal("myuser"))
		})
	})
	Context("Username as number", func() {
		var result teamvault.User
		JustBeforeEach(func() {
			result, err = remoteConnector.User(ctx, key)
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/key123/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `{"username":9876}`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns username as string", func() {
			Expect(result).NotTo(BeNil())
			Expect(result.String()).To(Equal("9876"))
		})
	})
	Context("Password", func() {
		var result teamvault.Password
		JustBeforeEach(func() {
			result, err = remoteConnector.Password(ctx, key)
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/key123/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(
						resp,
						`{"current_revision":"%s/api/secret-revisions/ref123/"}`,
						server.URL(),
					)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/secret-revisions/ref123/data",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `{"password":"S3CR3T"}`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns password", func() {
			Expect(result).NotTo(BeNil())
			Expect(result.String()).To(Equal("S3CR3T"))
		})
	})
	Context("Password as number", func() {
		var result teamvault.Password
		JustBeforeEach(func() {
			result, err = remoteConnector.Password(ctx, key)
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/key123/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(
						resp,
						`{"current_revision":"%s/api/secret-revisions/ref123/"}`,
						server.URL(),
					)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/secret-revisions/ref123/data",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `{"password":5784}`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns password as string", func() {
			Expect(result).NotTo(BeNil())
			Expect(result.String()).To(Equal("5784"))
		})
	})
	Context("Url", func() {
		var result teamvault.Url
		JustBeforeEach(func() {
			result, err = remoteConnector.Url(ctx, key)
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/key123/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `{"url":"http://my.example.com"}`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns url", func() {
			Expect(result).NotTo(BeNil())
			Expect(result.String()).To(Equal("http://my.example.com"))
		})
	})
	Context("Search", func() {
		var result []teamvault.Key
		JustBeforeEach(func() {
			result, err = remoteConnector.Search(ctx, "searchString")
		})
		BeforeEach(func() {
			server.RouteToHandler(
				http.MethodGet,
				"/api/secrets/",
				func(resp http.ResponseWriter, req *http.Request) {
					argUsername, argPassword, ok := req.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(argUsername).To(Equal(username))
					Expect(argPassword).To(Equal(password))
					Expect(req.FormValue("search")).To(Equal("searchString"))
					resp.WriteHeader(http.StatusOK)
					fmt.Fprintf(resp, `
					{
						"count": 1,
						"next": null,
						"previous": null,
						"results": [
							{
								"access_policy": "request",
								"allowed_groups": [],
								"allowed_users": [],
								"api_url": "https://teamvault.example.com/api/secrets/key123/",
								"content_type": "password",
								"created": "2017-08-21T12:29:53.252282Z",
								"created_by": "skegel",
								"current_revision": "https://teamvault.example.com/api/secret-revisions/rKp1x5/",
								"data_readable": [],
								"description": "",
								"filename": null,
								"last_read": "2017-08-30T08:37:02.189161Z",
								"name": "SearchString",
								"needs_changing_on_leave": true,
								"status": "ok",
								"url": "https://example.com",
								"username": "foo",
								"web_url": "https://teamvault.example.com/secrets/key123"
							}
						]
					}
				`)
				},
			)
			server.RouteToHandler(
				http.MethodGet,
				"/api/method/login",
				func(resp http.ResponseWriter, req *http.Request) {
					resp.WriteHeader(http.StatusOK)
				},
			)
		})
		It("returns no error", func() {
			Expect(err).To(BeNil())
		})
		It("returns url", func() {
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(teamvault.Key("key123")))
		})
	})
})
