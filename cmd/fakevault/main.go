// Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command fakevault is a minimal fake TeamVault HTTP server for hermetic,
// CI-runnable end-to-end tests of teamvault-cli. It implements only the four
// read endpoints the remote connector calls (see pkg/remote-connector.go) and
// serves a fixed set of in-memory secrets. It is a test helper and is never
// shipped (goreleaser builds the root `main: .` only).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// secret is one fixture entry served by the fake.
type secret struct {
	Username string
	URL      string
	Password string
	File     string
}

// fixtures is the seeded secret set. Scenarios/e2e tests reference these keys
// and assert the values below.
var fixtures = map[string]secret{
	"demo": {
		Username: "demo-user",
		URL:      "https://demo.example/login",
		Password: "demo-pass-123",
		File:     "demo-file-contents",
	},
	"AbC123": {
		Username: "alice",
		URL:      "https://api.internal",
		Password: "s3cr3t-value",
		File:     "certificate-bytes",
	},
}

// wantUser / wantPass are the Basic-auth credentials the server accepts (set
// via flags). A request with any other credential gets 401 — so tests can
// exercise the CLI's auth-failure path.
var (
	wantUser = "test"
	wantPass = "test"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:0", "listen address (use :0 for an OS-assigned port)")
	flag.StringVar(&wantUser, "user", "test", "required Basic-auth username")
	flag.StringVar(&wantPass, "pass", "test", "required Basic-auth password")
	flag.Parse()

	mux := http.NewServeMux()

	// GET /api/secrets/{key}/ — secret metadata (username, url, current_revision).
	mux.HandleFunc("GET /api/secrets/{key}/", func(w http.ResponseWriter, r *http.Request) {
		if !authOK(w, r) {
			return
		}
		s, ok := fixtures[r.PathValue("key")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]any{
			"username": s.Username,
			"url":      s.URL,
			"current_revision": fmt.Sprintf(
				"http://%s/api/secret-revisions/%s/",
				r.Host,
				r.PathValue("key"),
			),
		})
	})

	// GET /api/secret-revisions/{key}/data — revision data (password, file).
	mux.HandleFunc(
		"GET /api/secret-revisions/{key}/data",
		func(w http.ResponseWriter, r *http.Request) {
			if !authOK(w, r) {
				return
			}
			s, ok := fixtures[r.PathValue("key")]
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, map[string]any{"password": s.Password, "file": s.File})
		},
	)

	// GET /api/secrets/?search=q — search (login probe). Matches on key or username.
	mux.HandleFunc("GET /api/secrets/{$}", func(w http.ResponseWriter, r *http.Request) {
		if !authOK(w, r) {
			return
		}
		q := r.URL.Query().Get("search")
		results := make([]map[string]string, 0, len(fixtures))
		for key, s := range fixtures {
			if q == "" || strings.Contains(key, q) || strings.Contains(s.Username, q) {
				results = append(results, map[string]string{
					"api_url": fmt.Sprintf("http://%s/api/secrets/%s/", r.Host, key),
				})
			}
		}
		writeJSON(w, map[string]any{"results": results})
	})

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("fakevault: listen %s: %v", *addr, err)
	}
	fmt.Printf("fakevault listening on http://%s\n", ln.Addr())
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	log.Fatal(srv.Serve(ln))
}

// authOK requires the configured Basic-auth credential so tests can exercise
// both the happy path (correct creds) and the CLI's auth-failure path (wrong
// creds → 401). It writes 401 and returns false on any mismatch.
func authOK(w http.ResponseWriter, r *http.Request) bool {
	if user, pass, ok := r.BasicAuth(); ok && user == wantUser && pass == wantPass {
		return true
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="fakevault"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("fakevault: encode: %v", err)
	}
}
