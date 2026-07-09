// Copyright (c) 2016-2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/bborbe/argument/v2" //nolint:depguard // test needs to verify libargument × libtime.Duration contract
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var origCommandLine *flag.FlagSet

var _ = BeforeSuite(func() {
	origCommandLine = flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
})

var _ = AfterSuite(func() {
	flag.CommandLine = origCommandLine
})

var _ = Describe("TeamvaultTimeout CLI parse contract", func() {
	It("parses --teamvault-timeout=5s into a 5-second libtime.Duration", func() {
		app := &application{}
		err := argument.ParseArgs(context.Background(), app, []string{"--teamvault-timeout=5s"})
		Expect(err).NotTo(HaveOccurred())
		Expect(app.TeamvaultTimeout.Duration()).To(Equal(5 * time.Second))
	})
})

var _ = Describe("Teamvault Password", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/teamvault-utils/v5/cmd/teamvault-password")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Teamvault Password Suite")
}
