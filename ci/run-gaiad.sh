#!/bin/sh

gaiad start --home=/chain/gaia --grpc.address=$CHAIN_ID:9090 --pruning=nothing --log_level error 2>&1 | tee /chain/gaia/gaiad.log