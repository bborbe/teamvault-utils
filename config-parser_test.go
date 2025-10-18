// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teamvault_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/teamvault-utils/v4"
	"github.com/bborbe/teamvault-utils/v4/mocks"
)

var _ = Describe("Parser", func() {
	var ctx context.Context
	var err error
	var parser teamvault.ConfigParser
	var connector *mocks.Connector
	var content []byte
	var result []byte
	BeforeEach(func() {
		ctx = context.Background()
		connector = &mocks.Connector{}
		parser = teamvault.NewConfigParser(connector)
	})
	Context("Parse", func() {
		JustBeforeEach(func() {
			result, err = parser.Parse(ctx, content)
		})
		Context("content without placeholder", func() {
			BeforeEach(func() {
				content = []byte("hello world")
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal(content))
			})
		})
		Context("content teamvault user", func() {
			BeforeEach(func() {
				connector.UserReturns("myuser", nil)
				content = []byte(`{{ "key123" | teamvaultUser }}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("myuser")))
			})
		})
		Context("content teamvault password", func() {
			BeforeEach(func() {
				connector.PasswordReturns("mypass", nil)
				content = []byte(`{{ "key123" | teamvaultPassword }}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("mypass")))
			})
		})
		Context("content teamvault url", func() {
			BeforeEach(func() {
				connector.UrlReturns("http://example.com", nil)
				content = []byte(`{{ "key123" | teamvaultUrl }}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("http://example.com")))
			})
		})
		Context("content teamvault file", func() {
			BeforeEach(func() {
				connector.FileReturns(
					teamvault.File(base64.URLEncoding.EncodeToString([]byte("my-content"))),
					nil,
				)
				content = []byte(`{{ "key123" | teamvaultFile }}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("my-content")))
			})
		})
		Context("content teamvault fileBase64", func() {
			BeforeEach(func() {
				connector.FileReturns("YXNkZi1maWxl", nil)
				content = []byte(`{{ "key123" | teamvaultFileBase64 }}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("YXNkZi1maWxl")))
			})
		})
		Context("content teamvault base64", func() {
			BeforeEach(func() {
				content = []byte(`{{ "abc" | base64}}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("YWJj")))
			})
		})
		Context("content teamvault lower", func() {
			BeforeEach(func() {
				content = []byte(`{{ "aBc" | lower}}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("abc")))
			})
		})
		Context("content teamvault upper", func() {
			BeforeEach(func() {
				content = []byte(`{{ "aBc" | upper}}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("ABC")))
			})
		})
		Context("content teamvault env", func() {
			BeforeEach(func() {
				_ = os.Setenv("testEnv", "hello")
				content = []byte(`{{ "testEnv" | env}}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("hello")))
			})
		})
		Context("content teamvault htpasswd", func() {
			BeforeEach(func() {
				connector.UserReturns("myuser", nil)
				connector.PasswordReturns("mypass", nil)

				content = []byte(`{{ "abc" | teamvaultHtpasswd}}`)
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(HaveLen(68))
			})
		})
		Context("content teamvault file", func() {
			var f *os.File
			BeforeEach(func() {
				f, err = os.CreateTemp("", "")
				Expect(err).To(BeNil())
				_, _ = f.WriteString("hello world")
				_ = f.Close()
				content = []byte(fmt.Sprintf(`{{ "%s" | readfile }}`, f.Name()))
			})
			AfterEach(func() {
				if f != nil {
					_ = os.Remove(f.Name())
				}
			})
			It("returns no error", func() {
				Expect(err).To(BeNil())
			})
			It("correct result", func() {
				Expect(result).To(Equal([]byte("hello world")))
			})
		})
	})
})
