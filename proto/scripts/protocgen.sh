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

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./gaia -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 2 -name '*.proto'); do
      buf generate --template buf.gen.gogo.yaml $file
  done
done

cd ..

# move proto files to the right places
cp -r github.com/cosmos/gaia/* ./
rm -rf github.com

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
