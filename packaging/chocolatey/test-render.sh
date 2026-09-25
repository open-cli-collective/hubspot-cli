#!/usr/bin/env bash
set -euo pipefail

script="$(dirname "$0")/tools/chocolateyInstall.ps1"
amd64_url='https://github.com/open-cli-collective/hubspot-cli/releases/download/v9.9.9/hspt_9.9.9_windows_amd64.zip'
arm64_url='https://github.com/open-cli-collective/hubspot-cli/releases/download/v9.9.9/hspt_9.9.9_windows_arm64.zip'
rendered="$(sed \
  -e "s|URL_AMD64_PLACEHOLDER|$amd64_url|g" \
  -e "s|URL_ARM64_PLACEHOLDER|$arm64_url|g" \
  -e 's|CHECKSUM_AMD64_PLACEHOLDER|aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa|g' \
  -e 's|CHECKSUM_ARM64_PLACEHOLDER|bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb|g' \
  "$script")"

if grep -Eq 'URL_(AMD64|ARM64)_PLACEHOLDER|CHECKSUM_(AMD64|ARM64)_PLACEHOLDER|ChocolateyPackageVersion|\$\{[[:space:]]*version[[:space:]]*\}' <<<"$rendered"; then
  echo "Chocolatey template still contains an unrendered placeholder" >&2
  exit 1
fi
grep -Fq "$amd64_url" <<<"$rendered"
grep -Fq "$arm64_url" <<<"$rendered"
