#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto

buf generate --template buf.gen.gogo.yaml $file

cd ..

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
cp -r github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v*/packetforward/* packetforward/
rm -rf github.com

go mod tidy