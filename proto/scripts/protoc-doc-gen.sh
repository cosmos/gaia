#!/usr/bin/env bash

echo "Generating proto doc"

cd proto
buf generate --template buf.gen.doc.yaml
cd ..
