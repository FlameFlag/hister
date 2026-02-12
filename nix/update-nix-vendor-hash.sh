#!/usr/bin/env bash

set -euo pipefail

echo "::notice::Running nix build to get vendorHash..."
echo "::group::Build output"
OUTPUT=$(nix build .#hister 2>&1 || true)
echo "::endgroup::"

if echo "$OUTPUT" | grep -q "hash mismatch in fixed-output derivation"; then
  NEW_HASH=$(echo "$OUTPUT" | grep "got:" | sed 's/.*got: *//')
  echo "::notice::Found new hash: $NEW_HASH"
  
  sed -i.bak "s|vendorHash = \".*\";|vendorHash = \"$NEW_HASH\";|" nix/package.nix
  rm -f nix/package.nix.bak
  
  echo "::notice file=nix/package.nix::Updated vendorHash"
  
  echo "::group::Verifying build with new hash"
  nix build .#hister 2>&1 | tail -5
  echo "::endgroup::"
  echo "::notice::Build successful!"
else
  echo "::notice::No hash mismatch found or build succeeded already"
  if echo "$OUTPUT" | grep -q "these 2 derivations will be built"; then
    echo "::notice::Build already passes, no update needed"
  fi
  exit 0
fi
