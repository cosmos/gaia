#!/usr/bin/env bash

set -eo pipefail
echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./ -type f -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "$dir" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" "$file" &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml "$file"
    fi
  done
done
cd ..
# move proto files to the right places
if [ -d "github.com/cosmos/gaia" ]; then
    cp -r "github.com/cosmos/gaia/"* ./
else
    echo "Directory github.com/cosmos/gaia does not exist"
fi
if [ -d "github.com" ]; then
    rm -rf github.com
fi
