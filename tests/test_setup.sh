#!/usr/bin/env bash
set -euo pipefail

# File: tests/test_setup.sh
# Purpose: Self-contained, no-network regression checks for setup.sh using stubbed tools.
# Problem: Validate first-run/update flows, Starship config patching, and version marker without touching real system.
# Role: Developer test harness; safe to run locally. Uses temp HOME and stubbed brew/curl/git/starship.
# Usage: bash tests/test_setup.sh (from repo root). Leaves no state after success.
# Assumptions: bash available; no external deps; script resides at repo root with setup.sh present.

fail() {
  echo "FAILED: $*" >&2
  exit 1
}

make_stub_tools() {
  STUB_BIN="$(mktemp -d)"

  cat >"${STUB_BIN}/brew" <<'EOF'
#!/usr/bin/env bash
cmd="$1"; shift || true
case "$cmd" in
  tap)
    if [ "$#" -eq 0 ]; then
      echo "homebrew/core"
      exit 0
    else
      echo "$1"
      exit 0
    fi
    ;;
  info)
    # Simulate missing formulas/casks so installs skip optional AI CLIs.
    exit 1
    ;;
  ls)
    # Simulate not-installed formulas.
    exit 1
    ;;
  list)
    # Simulate not-installed casks.
    exit 1
    ;;
  install)
    exit 0
    ;;
  *)
    exit 0
    ;;
esac
EOF

  cat >"${STUB_BIN}/curl" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF

  cat >"${STUB_BIN}/git" <<'EOF'
#!/usr/bin/env bash
echo "[git stub] $*" >&2
exit 0
EOF

  cat >"${STUB_BIN}/starship" <<'EOF'
#!/usr/bin/env bash
# Minimal preset handler to create config when -o is provided.
if [ "$1" = "preset" ] && [ "$2" = "nerd-font-symbols" ]; then
  while getopts ":o:" opt; do
    [ "$opt" = "o" ] && printf "# stub preset\n[package]\nsymbol = \"★\"\n" >"$OPTARG"
  done
  exit 0
fi
echo "[starship stub] $*" >&2
exit 0
EOF

  # Pretend uv exists so install_uv skips.
  cat >"${STUB_BIN}/uv" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF

  chmod +x "${STUB_BIN}/"*
}

cleanup() {
  rm -rf "${STUB_BIN:-}" "${TMP_HOME:-}"
}

assert_single_package_block() {
  local cfg="$1"
  local count
  count="$(grep -c '^\[package\]' "$cfg")"
  [ "$count" -eq 1 ] || fail "expected 1 [package] block, found $count"
  grep -q 'display_private = true' "$cfg" || fail "display_private=true missing"
}

run_first_install() {
  TMP_HOME="$(mktemp -d)"
  make_stub_tools
  PATH="${STUB_BIN}:/usr/bin:/bin"
  HOME="$TMP_HOME" PATH="$PATH" ./setup.sh >/tmp/dev-setup-test.log 2>&1

  [ -f "${TMP_HOME}/.local/share/dev-setup/version" ] || fail "version file missing"
  VERSION_VALUE="$(cat "${TMP_HOME}/.local/share/dev-setup/version")"
  [ -n "$VERSION_VALUE" ] || fail "version value empty"

  STAR_CFG="${TMP_HOME}/.config/starship.toml"
  [ -f "$STAR_CFG" ] || fail "starship config not created"
  assert_single_package_block "$STAR_CFG"
}

run_update_install() {
  echo "0.0.1" >"${TMP_HOME}/.local/share/dev-setup/version"
  PATH="${STUB_BIN}:/usr/bin:/bin"
  HOME="$TMP_HOME" PATH="$PATH" ./setup.sh >/tmp/dev-setup-test-2.log 2>&1

  NEW_VERSION="$(cat "${TMP_HOME}/.local/share/dev-setup/version")"
  [ "$NEW_VERSION" = "$VERSION_VALUE" ] || fail "version not updated to $VERSION_VALUE"
  assert_single_package_block "${TMP_HOME}/.config/starship.toml"
}

main() {
  trap cleanup EXIT
  run_first_install
  run_update_install
  echo "All tests passed."
}

main "$@"
