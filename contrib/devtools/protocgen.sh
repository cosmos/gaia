#!/usr/bin/env bash

set -eo pipefail

# If these fail run the first run commands in the module README.md
# make sure $GO/bin is in your PATH
which protoc-gen-gocosmos
which protoc-gen-go-grpc
which buf

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  buf protoc \
  -I "proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

  # # command to generate gRPC gateway (*.pb.gw.go in respective modules) files
  buf protoc \
  -I "proto" \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

done

# move proto files to the right places
cp -r github.com/althea-net/althea-chain/* ./
rm -rf github.com
