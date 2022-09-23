#!/usr/bin/env bash

#set -eo pipefail
#
#protoc_gen_gocosmos() {
#  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
#    echo -e "\tPlease run this command from somewhere inside the gaia folder."
#    return 1
#  fi
#
#  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
#}
#
#protoc_gen_gocosmos
#
#proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
#for dir in $proto_dirs; do
#  protoc \
#  -I "proto" \
#  -I "third_party/proto" \
#  --gocosmos_out=plugins=interfacetype+grpc,\
#Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
#  --grpc-gateway_out=logtostderr=true:. \
#  $(find "${dir}" -maxdepth 1 -name '*.proto')
#
#done
#
## command to generate docs using protoc-gen-doc
#protoc \
#-I "proto" \
#-I "third_party/proto" \
#--doc_out=./docs/proto \
#--doc_opt=./docs/proto/protodoc-markdown.tmpl,proto-docs.md \
#$(find "proto" -maxdepth 5 -name '*.proto')
#
## move proto files to the right places
#cp -r github.com/cosmos/gaia/x/* x/
#rm -rf github.com

set -e

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./cosmos -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # this regex checks if a proto file has its go_package set to cosmossdk.io/api/...
    # gogo proto files SHOULD ONLY be generated if this is false
    # we don't want gogo proto to run for proto files which are natively built for google.golang.org/protobuf
    if grep -q "option go_package" "$file" && grep -H -o -c 'option go_package.*cosmossdk.io/api' "$file" | grep -q ':0$'; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# generate codec/testdata proto code
(cd testutil/testdata; buf generate)

# generate baseapp test messages
(cd baseapp/testutil; buf generate)

# move proto files to the right places
cp -r github.com/cosmos/cosmos-sdk/* ./
rm -rf github.com
