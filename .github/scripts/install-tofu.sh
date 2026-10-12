#!/bin/bash
# Installs the pinned OpenTofu release into .bin/tofu, after checking it against the
# release's SHA256SUMS, itself checked against the hash pinned here, and prints its path. The tests, the acceptance tests and the
# docs build all run OpenTofu from here.
set -euo pipefail

TOFU_VERSION="1.13.1"
# sha256 of that release's SHA256SUMS: a version bump updates both
TOFU_SUMS_SHA256="f4a03dd8613320c665848cb6e1962ee112d777c9c6a8a725586dcc341b1c7252"

root=$(git rev-parse --show-toplevel)
bin="$root/.bin"
tofu="$bin/tofu"

if [ -x "$tofu" ] && "$tofu" version | head -1 | grep -qx "OpenTofu v$TOFU_VERSION"; then
  echo "$tofu"
  exit 0
fi

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

archive="tofu_${TOFU_VERSION}_${os}_${arch}.tar.gz"
url="https://github.com/opentofu/opentofu/releases/download/v$TOFU_VERSION"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -sSfL -o "$tmp/$archive" "$url/$archive"
curl -sSfL -o "$tmp/SHA256SUMS" "$url/tofu_${TOFU_VERSION}_SHA256SUMS"
echo "$TOFU_SUMS_SHA256  $tmp/SHA256SUMS" | sha256sum -c - >&2
(cd "$tmp" && grep " $archive\$" SHA256SUMS | sha256sum -c - >&2)

mkdir -p "$bin"
tar -xzf "$tmp/$archive" -C "$tmp" tofu
mv "$tmp/tofu" "$tofu"
echo "$tofu"
