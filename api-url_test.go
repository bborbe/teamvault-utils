package teamvault_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bborbe/teamvault-utils"
)

var _ = Describe("ApiUrl", func() {
	var apiUrl teamvault.ApiUrl
	BeforeEach(func() {
		apiUrl = "foo"
	})
	Context("String", func() {
		var result string
		JustBeforeEach(func() {
			result = apiUrl.String()
		})
		It("returns correct string", func() {
			Expect(result).To(Equal("foo"))
		})
	})
	DescribeTable("parse key", func(
		apiUrl teamvault.ApiUrl,
		expectedError bool,
		expectedKey teamvault.Key,
	) {
		key, err := apiUrl.Key()
		if expectedError {
			Expect(err).NotTo(BeNil())
		} else {
			Expect(err).To(BeNil())
			Expect(key).To(Equal(expectedKey))
		}
	},
		Entry("empty", teamvault.ApiUrl(""), true, teamvault.Key("")),
		Entry("slash", teamvault.ApiUrl("/"), true, teamvault.Key("")),
		Entry("two slashes", teamvault.ApiUrl("hello/my/world"), false, teamvault.Key("my")),
		Entry("valid url", teamvault.ApiUrl("https://teamvault.example.com/api/secrets/key123/"), false, teamvault.Key("key123")),
	)
})
