#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

buf generate --template buf.gen.gogo.yaml $file

cd ..



# move proto files to the right places
cp -r ./github.com/strangelove-ventures/tokenfactory/x/* x/
rm -rf ./github.com

# replace incorrect namespace
find ./x -type f -name '*.pb.go' -exec sed -i -e 's|cosmossdk.io/x/bank/types|github.com/cosmos/cosmos-sdk/x/bank/types|g' {} \;

go mod tidy