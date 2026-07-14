// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	libhttp "github.com/bborbe/http"
	libtime "github.com/bborbe/time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
)

var _ = Describe("RemoteWriter", func() {
	var ctx context.Context
	var server *ghttp.Server
	var username string
	var password string
	var writer teamvault.Writer

	BeforeEach(func() {
		ctx = context.Background()
		server = ghttp.NewServer()
		username = "user"
		password = "pass"
		writer = teamvault.NewRemoteWriter(
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

	Describe("Create", func() {
		Context("password secret", func() {
			It("posts to /api/secrets/ with correct auth and body", func() {
				var receivedBody map[string]any

				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/secrets/"),
						ghttp.VerifyBasicAuth(username, password),
						ghttp.VerifyContentType("application/json"),
						func(resp http.ResponseWriter, req *http.Request) {
							receivedBody = decodeJSONBody(req)
							resp.WriteHeader(http.StatusCreated)
							//nolint:errcheck
							fmt.Fprintf(resp, `{"api_url": "%s/api/secrets/AbC123/"}`, server.URL())
						},
					),
				)

				key, apiUrl, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "my-secret",
					Username:    "user@example.com",
					Url:         "http://example.com",
					Description: "my description",
					Password:    teamvault.Password("secret123"),
				})

				Expect(err).To(BeNil())
				Expect(key).To(Equal(teamvault.Key("AbC123")))
				Expect(apiUrl).To(Equal(teamvault.ApiUrl(server.URL() + "/api/secrets/AbC123/")))

				Expect(receivedBody["content_type"]).To(Equal("password"))
				Expect(receivedBody["name"]).To(Equal("my-secret"))
				Expect(receivedBody["username"]).To(Equal("user@example.com"))
				Expect(receivedBody["url"]).To(Equal("http://example.com"))
				Expect(receivedBody["description"]).To(Equal("my description"))
				secretData := getMap(receivedBody, "secret_data")
				Expect(secretData["password"]).To(Equal("secret123"))
				_, hasFile := secretData["file_content"]
				Expect(hasFile).To(BeFalse())
			})
		})

		Context("file secret", func() {
			It("base64-encodes file content", func() {
				var receivedBody map[string]any

				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/secrets/"),
						ghttp.VerifyBasicAuth(username, password),
						ghttp.VerifyContentType("application/json"),
						func(resp http.ResponseWriter, req *http.Request) {
							receivedBody = decodeJSONBody(req)
							resp.WriteHeader(http.StatusCreated)
							//nolint:errcheck
							fmt.Fprintf(resp, `{"api_url": "%s/api/secrets/AbC456/"}`, server.URL())
						},
					),
				)

				fileContent := []byte("hello world")
				key, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypeFile,
					Name:        "my-file",
					FileContent: fileContent,
				})

				Expect(err).To(BeNil())
				Expect(key).To(Equal(teamvault.Key("AbC456")))

				Expect(receivedBody["content_type"]).To(Equal("file"))
				Expect(receivedBody["name"]).To(Equal("my-file"))
				secretData := getMap(receivedBody, "secret_data")
				Expect(
					secretData["file_content"],
				).To(Equal(base64.StdEncoding.EncodeToString(fileContent)))
				_, hasPassword := secretData["password"]
				Expect(hasPassword).To(BeFalse())
			})
		})
	})

	Describe("Update", func() {
		Context("metadata-only", func() {
			It("PATCHes without secret_data", func() {
				var receivedBody map[string]any

				server.RouteToHandler(
					http.MethodPatch,
					"/api/secrets/AbC123/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPatch, "/api/secrets/AbC123/"),
						ghttp.VerifyBasicAuth(username, password),
						ghttp.VerifyContentType("application/json"),
						func(resp http.ResponseWriter, req *http.Request) {
							receivedBody = decodeJSONBody(req)
							resp.WriteHeader(http.StatusOK)
							//nolint:errcheck
							fmt.Fprintf(resp, `{"api_url": "%s/api/secrets/AbC123/"}`, server.URL())
						},
					),
				)

				desc := "new description"
				key, apiUrl, err := writer.Update(
					ctx,
					teamvault.Key("AbC123"),
					teamvault.UpdateSecret{
						Description: &desc,
					},
				)

				Expect(err).To(BeNil())
				Expect(key).To(Equal(teamvault.Key("AbC123")))
				Expect(apiUrl).To(Equal(teamvault.ApiUrl(server.URL() + "/api/secrets/AbC123/")))
				Expect(receivedBody["description"]).To(Equal("new description"))
				_, hasSecretData := receivedBody["secret_data"]
				Expect(hasSecretData).To(BeFalse())
			})
		})

		Context("with new password", func() {
			It("PATCHes with secret_data.password", func() {
				var receivedBody map[string]any

				server.RouteToHandler(
					http.MethodPatch,
					"/api/secrets/AbC123/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPatch, "/api/secrets/AbC123/"),
						ghttp.VerifyBasicAuth(username, password),
						ghttp.VerifyContentType("application/json"),
						func(resp http.ResponseWriter, req *http.Request) {
							receivedBody = decodeJSONBody(req)
							resp.WriteHeader(http.StatusOK)
							//nolint:errcheck
							fmt.Fprintf(resp, `{"api_url": "%s/api/secrets/AbC123/"}`, server.URL())
						},
					),
				)

				newPw := teamvault.Password("new-secret-pw")
				key, _, err := writer.Update(ctx, teamvault.Key("AbC123"), teamvault.UpdateSecret{
					Password: &newPw,
				})

				Expect(err).To(BeNil())
				Expect(key).To(Equal(teamvault.Key("AbC123")))
				secretData := getMap(receivedBody, "secret_data")
				Expect(secretData["password"]).To(Equal("new-secret-pw"))
			})
		})

		Context("with new file content", func() {
			It("PATCHes with secret_data.file_content base64-encoded", func() {
				var receivedBody map[string]any

				server.RouteToHandler(
					http.MethodPatch,
					"/api/secrets/AbC123/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPatch, "/api/secrets/AbC123/"),
						ghttp.VerifyBasicAuth(username, password),
						ghttp.VerifyContentType("application/json"),
						func(resp http.ResponseWriter, req *http.Request) {
							receivedBody = decodeJSONBody(req)
							resp.WriteHeader(http.StatusOK)
							//nolint:errcheck
							fmt.Fprintf(resp, `{"api_url": "%s/api/secrets/AbC123/"}`, server.URL())
						},
					),
				)

				fileContent := []byte("updated file content")
				key, _, err := writer.Update(ctx, teamvault.Key("AbC123"), teamvault.UpdateSecret{
					FileContent: fileContent,
				})

				Expect(err).To(BeNil())
				Expect(key).To(Equal(teamvault.Key("AbC123")))
				secretData := getMap(receivedBody, "secret_data")
				Expect(
					secretData["file_content"],
				).To(Equal(base64.StdEncoding.EncodeToString(fileContent)))
			})
		})
	})

	Describe("GeneratePassword", func() {
		It("POSTs to /api/generate_password/ and returns the password", func() {
			server.RouteToHandler(
				http.MethodPost,
				"/api/generate_password/",
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/api/generate_password/"),
					ghttp.VerifyBasicAuth(username, password),
					ghttp.VerifyContentType("application/json"),
					func(resp http.ResponseWriter, req *http.Request) {
						resp.WriteHeader(http.StatusOK)
						//nolint:errcheck
						fmt.Fprint(resp, `{"password": "gen3rat3d"}`)
					},
				),
			)

			pwd, err := writer.GeneratePassword(ctx)

			Expect(err).To(BeNil())
			Expect(pwd).To(Equal(teamvault.Password("gen3rat3d")))
		})

		It("returns auth error on 401", func() {
			server.RouteToHandler(
				http.MethodPost,
				"/api/generate_password/",
				ghttp.RespondWith(http.StatusUnauthorized, nil),
			)

			_, err := writer.GeneratePassword(ctx)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("status: 401"))
			Expect(err.Error()).To(ContainSubstring("teamvault-cli login"))
		})
	})

	Describe("error handling", func() {
		Context("401 authentication failure", func() {
			It("returns error with login hint", func() {
				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.RespondWith(http.StatusUnauthorized, nil),
				)

				_, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "n",
					Password:    teamvault.Password("p"),
				})

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("status: 401"))
				Expect(err.Error()).To(ContainSubstring("teamvault-cli login"))
			})
		})

		Context("403 forbidden", func() {
			It("returns error with login hint", func() {
				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.RespondWith(http.StatusForbidden, nil),
				)

				_, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "n",
					Password:    teamvault.Password("p"),
				})

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("status: 403"))
				Expect(err.Error()).To(ContainSubstring("teamvault-cli login"))
			})
		})

		Context("400 bad request", func() {
			It("returns error with status code", func() {
				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.RespondWith(http.StatusBadRequest, nil),
				)

				_, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "n",
					Password:    teamvault.Password("p"),
				})

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("status: 400"))
			})
		})

		Context("transport failure", func() {
			It("returns error when server is closed", func() {
				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.RespondWith(http.StatusOK, nil),
				)
				server.Close()

				_, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "n",
					Password:    teamvault.Password("p"),
				})

				Expect(err).NotTo(BeNil())
				var urlErr *url.Error
				Expect(errors.As(err, &urlErr)).To(BeTrue())
			})
		})

		Context("invalid JSON response", func() {
			It("returns decode error", func() {
				server.RouteToHandler(
					http.MethodPost,
					"/api/secrets/",
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(http.MethodPost, "/api/secrets/"),
						ghttp.VerifyBasicAuth(username, password),
						func(resp http.ResponseWriter, req *http.Request) {
							resp.WriteHeader(http.StatusOK)
							//nolint:errcheck
							fmt.Fprint(resp, `not valid json`)
						},
					),
				)

				_, _, err := writer.Create(ctx, teamvault.CreateSecret{
					ContentType: teamvault.ContentTypePassword,
					Name:        "n",
					Password:    teamvault.Password("p"),
				})

				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("decode response failed"))
			})
		})
	})
})

func decodeJSONBody(req *http.Request) map[string]any {
	var result map[string]any
	err := json.NewDecoder(req.Body).Decode(&result)
	if err != nil {
		return nil
	}
	return result
}

// getMap safely extracts a map from a map with string key.
func getMap(m map[string]any, key string) map[string]any {
	v, ok := m[key].(map[string]any)
	if !ok {
		return nil
	}
	return v
}
