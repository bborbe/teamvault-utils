#!/usr/bin/env bash
# Copyright (c) 2016-2026 Benjamin Borbe All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
#
# Shared helpers for teamvault-cli scenarios and `make e2e`. Mirrors the
# convention in vault-cli / dark-factory (build_binary + assert_* + scenario_done).
# Fakes the external TeamVault dependency with cmd/fakevault (cf. dark-factory's
# local-bare-remote in setup_sandbox_copy).
#
# Usage:
#   source scenarios/helper/lib.sh
#   build_binaries; start_fakevault
#   assert_eq "desc" "expected" "$actual"
#   assert_exit_nonzero "desc" some-command --with args
#   scenario_done

# Repo root, derived from this file's location (works in worktrees too).
TV_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
_FAIL=0
FV_PID=""

_cleanup() {
	[ -n "$FV_PID" ] && kill "$FV_PID" 2>/dev/null || true
	[ -n "${WORK_DIR:-}" ] && rm -rf "$WORK_DIR"
}
trap _cleanup EXIT

# build_binaries builds teamvault-cli + fakevault into a temp dir and sets $TV.
build_binaries() {
	WORK_DIR="$(mktemp -d)"
	go build -C "$TV_ROOT" -o "$WORK_DIR/teamvault-cli" . || exit 1
	go build -C "$TV_ROOT" -o "$WORK_DIR/fakevault" ./cmd/fakevault || exit 1
	TV="$WORK_DIR/teamvault-cli"
}

# start_fakevault launches the fake server on an OS-assigned port, writes a temp
# config (url+user+pass) pointing at it, and exports TEAMVAULT_CONFIG.
start_fakevault() {
	"$WORK_DIR/fakevault" --addr 127.0.0.1:0 >"$WORK_DIR/fv.log" 2>&1 &
	FV_PID=$!
	FV_URL=""
	for _ in $(seq 1 100); do
		FV_URL="$(sed -n 's#^fakevault listening on ##p' "$WORK_DIR/fv.log")"
		[ -n "$FV_URL" ] && break
		sleep 0.1
	done
	if [ -z "$FV_URL" ]; then
		echo "fakevault did not start"; cat "$WORK_DIR/fv.log"; exit 1
	fi
	# Password lives in the config file so the Keychain is never consulted — the
	# CI runner has no macOS Keychain / freedesktop secret service.
	printf '{"url":"%s","user":"test","pass":"test"}\n' "$FV_URL" >"$WORK_DIR/config.json"
	export TEAMVAULT_CONFIG="$WORK_DIR/config.json"
}

# assert_eq <desc> <expected> <actual>
assert_eq() {
	if [ "$2" = "$3" ]; then
		echo "  ok: $1"
	else
		echo "  FAIL: $1 — expected '$2' got '$3'"; _FAIL=1
	fi
}

# assert_exit_nonzero <desc> <command...>
assert_exit_nonzero() {
	local desc="$1"; shift
	if "$@" >/dev/null 2>&1; then
		echo "  FAIL: $desc — expected non-zero exit"; _FAIL=1
	else
		echo "  ok: $desc"
	fi
}

# scenario_done reports and exits non-zero if any assertion failed.
scenario_done() {
	if [ "$_FAIL" -eq 0 ]; then
		echo "e2e: PASS"
	else
		echo "e2e: FAIL"; exit 1
	fi
}
