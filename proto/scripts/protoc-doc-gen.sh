#!/usr/bin/env bash

# command to generate docs using protoc-gen-doc
protoc \
-I "proto" \
-I "third_party/proto" \
--doc_out=./docs/proto \
--doc_opt=./docs/proto/protodoc-markdown.tmpl,proto-docs.md \
$(find "proto" -maxdepth 5 -name '*.proto')
