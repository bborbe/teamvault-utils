#!/usr/bin/env bash
# Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# Hermetic end-to-end test: start the fake TeamVault server (cmd/fakevault),
# point the real teamvault-cli binary at it via a temp config + TEAMVAULT_PASS
# (no Keychain, no live TeamVault), and assert the seeded secret values.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
FV_PID=""
cleanup() {
	[ -n "$FV_PID" ] && kill "$FV_PID" 2>/dev/null || true
	rm -rf "$TMP"
}
trap cleanup EXIT

echo "e2e: building binaries"
go build -C "$ROOT" -o "$TMP/teamvault-cli" .
go build -C "$ROOT" -o "$TMP/fakevault" ./cmd/fakevault

echo "e2e: starting fakevault"
"$TMP/fakevault" --addr 127.0.0.1:0 >"$TMP/fv.log" 2>&1 &
FV_PID=$!

URL=""
for _ in $(seq 1 50); do
	URL="$(sed -n 's#^fakevault listening on ##p' "$TMP/fv.log")"
	[ -n "$URL" ] && break
	sleep 0.1
done
if [ -z "$URL" ]; then
	echo "e2e: fakevault did not start"; cat "$TMP/fv.log"; exit 1
fi
echo "e2e: fakevault at $URL"

printf '{"url":"%s","user":"test"}\n' "$URL" >"$TMP/config.json"
export TEAMVAULT_CONFIG="$TMP/config.json" TEAMVAULT_PASS="test"

fail=0
check() { # check <desc> <expected> <actual>
	if [ "$2" = "$3" ]; then
		echo "  ok: $1"
	else
		echo "  FAIL: $1 — expected '$2' got '$3'"; fail=1
	fi
}

check "password demo"   "demo-pass-123"               "$("$TMP/teamvault-cli" password --teamvault-key demo)"
check "username demo"   "demo-user"                   "$("$TMP/teamvault-cli" username --teamvault-key demo)"
check "url demo"        "https://demo.example/login"  "$("$TMP/teamvault-cli" url --teamvault-key demo)"
check "file demo"       "demo-file-contents"          "$("$TMP/teamvault-cli" file --teamvault-key demo)"
check "password AbC123" "s3cr3t-value"                "$("$TMP/teamvault-cli" password --teamvault-key AbC123)"

# Basic-auth-safe: raw output must have NO trailing newline ("demo-pass-123" = 13 bytes).
BYTES="$("$TMP/teamvault-cli" password --teamvault-key demo | wc -c | tr -d ' ')"
check "no trailing newline (byte count)" "13" "$BYTES"

if [ "$fail" -eq 0 ]; then
	echo "e2e: PASS"
else
	echo "e2e: FAIL"; exit 1
fi
