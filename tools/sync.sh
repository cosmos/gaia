#!/usr/bin/env bash

set -x
set -euo pipefail

read -r HEIGHT HASH <<<$(curl -sSf 'localhost:26657/commit?height=1' | jq -r '"\(.result.signed_header.header.height) \(.result.signed_header.commit.block_id.hash)"')

if docker inspect gaiadnode3 >/dev/null; then
  docker stop gaiadnode3
fi
gsed -ire 's/^enabled = .*/enabled = true/g' build/node3/gaiad/config/config.toml
gsed -ire 's|^rpc_servers = .*|rpc_servers = "http://192.168.10.2:26657,http://192.168.10.3:26657"|g' build/node3/gaiad/config/config.toml
gsed -ire 's/^trusted_height = .*/trusted_height = '"$HEIGHT"'/g' build/node3/gaiad/config/config.toml
gsed -ire 's/^trusted_hash = .*/trusted_hash = "'"$HASH"'"/g' build/node3/gaiad/config/config.toml
gsed -ire 's/^trusted_period = .*/trusted_period = "1h"/g' build/node3/gaiad/config/config.toml
docker-compose up gaiadnode3
