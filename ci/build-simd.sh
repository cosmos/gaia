#!/usr/bin/env bash

## Programmatic list for creating a simd chain for testing IBC.
## Instead of blindly running this code, read it line by line and understand the dependecies and tasks.
## Prerequisites: Log into Docker Hub
set -eou pipefail
GAIA_BRANCH="master"

echo "*** Requirements"
which git && which go && which make && which docker

echo "*** Fetch gaiad source code"
git clone https://github.com/cosmos/cosmos-sdk || echo "Already cloned."
cd cosmos-sdk
git checkout "${GAIA_BRANCH}" -q

echo "*** Build binary"
GOOS=linux make build-simd

echo "*** Create Docker image and upload to Docker Hub"
cd ..
docker build -t informaldev/simd -f simd.Dockerfile .
read -p "Press ENTER to push image to Docker Hub or CTRL-C to cancel. " dontcare
docker push informaldev/simd
