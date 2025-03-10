#!/usr/bin/env bash

set -eo pipefail
echo "Generating gogo proto code"

cd proto

proto_dirs=$(find . -name '*.proto' -print0 | xargs -0 -n1 dirname | sort -u)

for dir in $proto_dirs; do
  for file in "$dir"/*.proto; do
    if grep -q "option go_package" "$file"; then
      buf generate --template buf.gen.gogo.yaml "$file"
    fi
  done
done

cd ..

# move proto files to the right places
cp -r github.com/cosmos/gaia/* ./
rm -rf github.com
