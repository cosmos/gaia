#!/usr/bin/env bash

set -eo pipefail

protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the gaia folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
}

protoc_gen_gocosmos

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  protoc \
  -I "proto" \
  -I "third_party/proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

done

# move proto files to the right places
cp -r github.com/cosmos/gaia/x/* x/
rm -rf github.com
