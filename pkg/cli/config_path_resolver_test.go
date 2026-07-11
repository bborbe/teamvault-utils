// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	stderrors "errors"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// stubNoHome makes userHomeDir report failure for the duration of the spec, so
// the "no home directory" branch is exercised deterministically regardless of
// the platform's os.UserHomeDir getpwuid_r fallback.
func stubNoHome() {
	orig := userHomeDir
	userHomeDir = func() (string, error) { return "", stderrors.New("no home") }
	DeferCleanup(func() { userHomeDir = orig })
}

// saveEnv snapshots the given env vars and restores them after the spec.
func saveEnv(keys ...string) {
	saved := make(map[string]string, len(keys))
	for _, k := range keys {
		saved[k] = os.Getenv(k)
	}
	DeferCleanup(func() {
		for k, v := range saved {
			if v == "" {
				Expect(os.Unsetenv(k)).To(Succeed())
			} else {
				Expect(os.Setenv(k, v)).To(Succeed())
			}
		}
	})
}

var _ = Describe("resolveDefaultConfigPath", func() {
	var tmpDir string

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
		saveEnv("TEAMVAULT_CONFIG", "XDG_CONFIG_HOME", "HOME")
		Expect(os.Setenv("XDG_CONFIG_HOME", tmpDir)).To(Succeed())
	})

	It("returns TEAMVAULT_CONFIG when set, ignoring the candidates", func() {
		Expect(os.Setenv("TEAMVAULT_CONFIG", "/explicit/config.json")).To(Succeed())
		Expect(resolveDefaultConfigPath()).To(Equal("/explicit/config.json"))
	})

	It("returns the XDG path when its config exists and env is empty", func() {
		Expect(os.Unsetenv("TEAMVAULT_CONFIG")).To(Succeed())
		xdg := filepath.Join(tmpDir, "teamvault-cli", "config.json")
		Expect(os.MkdirAll(filepath.Dir(xdg), 0o700)).To(Succeed())
		Expect(os.WriteFile(xdg, []byte(`{"url":"x","user":"y"}`), 0o600)).To(Succeed())
		Expect(resolveDefaultConfigPath()).To(Equal(xdg))
	})

	It("falls back to the legacy path when the XDG config is absent", func() {
		Expect(os.Unsetenv("TEAMVAULT_CONFIG")).To(Succeed())
		// XDG_CONFIG_HOME points at tmpDir but no teamvault-cli/config.json exists.
		Expect(resolveDefaultConfigPath()).To(Equal(legacyConfigPath))
	})

	It("falls back to the legacy path when the home directory cannot be determined", func() {
		Expect(os.Unsetenv("TEAMVAULT_CONFIG")).To(Succeed())
		Expect(os.Unsetenv("XDG_CONFIG_HOME")).To(Succeed())
		stubNoHome()
		Expect(resolveDefaultConfigPath()).To(Equal(legacyConfigPath))
	})
})

var _ = Describe("xdgConfigPath", func() {
	BeforeEach(func() {
		saveEnv("XDG_CONFIG_HOME", "HOME")
	})

	It("uses XDG_CONFIG_HOME when set", func() {
		Expect(os.Setenv("XDG_CONFIG_HOME", "/custom/xdg")).To(Succeed())
		Expect(
			xdgConfigPath(),
		).To(Equal(filepath.Join("/custom/xdg", "teamvault-cli", "config.json")))
	})

	It("defaults to $HOME/.config when XDG_CONFIG_HOME is empty", func() {
		Expect(os.Unsetenv("XDG_CONFIG_HOME")).To(Succeed())
		Expect(os.Setenv("HOME", "/home/tester")).To(Succeed())
		Expect(
			xdgConfigPath(),
		).To(Equal(filepath.Join("/home/tester", ".config", "teamvault-cli", "config.json")))
	})

	It(
		"returns empty when XDG_CONFIG_HOME is unset and the home directory cannot be determined",
		func() {
			Expect(os.Unsetenv("XDG_CONFIG_HOME")).To(Succeed())
			stubNoHome()
			Expect(xdgConfigPath()).To(Equal(""))
		},
	)
})
