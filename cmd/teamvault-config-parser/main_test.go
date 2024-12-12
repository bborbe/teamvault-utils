package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Teamvault Config Parser", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/teamvault-utils/v4/cmd/teamvault-config-parser")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Teamvault Config Parser Suite")
}
