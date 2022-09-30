#!/usr/bin/env bash

cd proto
buf generate --template buf.gen.doc.yaml
cd ..
