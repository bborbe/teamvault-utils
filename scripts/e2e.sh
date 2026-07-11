#!/usr/bin/env bash
# Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# Hermetic end-to-end test: start cmd/fakevault and drive the real teamvault-cli
# binary against it (no live TeamVault, no Keychain). Shares setup/assert helpers
# with the scenarios via scenarios/helper/lib.sh.
set -uo pipefail

source "$(cd "$(dirname "$0")/.." && pwd)/scenarios/helper/lib.sh"

echo "e2e: building binaries"
build_binaries
echo "e2e: starting fakevault"
start_fakevault
echo "e2e: fakevault at $FV_URL"

# Secret reads (config resolved via TEAMVAULT_CONFIG env — set by start_fakevault).
assert_eq "password" "demo-pass-123"              "$("$TV" password --teamvault-key demo)"
assert_eq "username" "demo-user"                  "$("$TV" username --teamvault-key demo)"
assert_eq "url"      "https://demo.example/login" "$("$TV" url --teamvault-key demo)"
assert_eq "file"     "demo-file-contents"         "$("$TV" file --teamvault-key demo)"

# config parse — Go-template funcs resolve secrets from stdin to stdout.
assert_eq "config parse" "user=demo-user pass=demo-pass-123" \
	"$(printf 'user={{ teamvaultUser "demo" }} pass={{ teamvaultPassword "demo" }}' | "$TV" config parse)"

# not-found — an unknown key errors (non-zero exit).
assert_exit_nonzero "not-found error" "$TV" password --teamvault-key nope-does-not-exist

# config via env var (TEAMVAULT_CONFIG) — already exported; resolve without a flag.
assert_eq "config via env" "demo-user" "$("$TV" username --teamvault-key demo)"

# config via flag (--teamvault-config) — with TEAMVAULT_CONFIG unset, the flag wins.
assert_eq "config via flag" "demo-user" \
	"$(env -u TEAMVAULT_CONFIG "$TV" username --teamvault-config "$WORK_DIR/config.json" --teamvault-key demo)"

# config generate — render a template directory tree to a target directory.
mkdir -p "$WORK_DIR/gen-src"
printf 'db_pass={{ teamvaultPassword "demo" }}' >"$WORK_DIR/gen-src/app.conf"
"$TV" config generate --source-dir "$WORK_DIR/gen-src" --target-dir "$WORK_DIR/gen-out"
assert_eq "config generate" "db_pass=demo-pass-123" "$(cat "$WORK_DIR/gen-out/app.conf")"

# auth failure — a config with the wrong password gets 401 → non-zero exit.
printf '{"url":"%s","user":"test","pass":"wrong"}\n' "$FV_URL" >"$WORK_DIR/badconfig.json"
assert_exit_nonzero "auth failure (401)" \
	"$TV" password --teamvault-config "$WORK_DIR/badconfig.json" --teamvault-key demo

# Basic-auth-safe: raw output has NO trailing newline ("demo-pass-123" = 13 bytes).
assert_eq "no trailing newline" "13" "$("$TV" password --teamvault-key demo | wc -c | tr -d ' ')"

scenario_done
