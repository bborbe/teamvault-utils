package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Teamvault Config Dir Generator", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/teamvault-utils/v4/cmd/teamvault-config-dir-generator")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Teamvault Config Dir Generator Suite")
}
