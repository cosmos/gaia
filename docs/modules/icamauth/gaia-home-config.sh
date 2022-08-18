#!/bin/zsh
gaiad config chain-id test-1 --home test-1
gaiad config keyring-backend test --home test-1
gaiad config node http://localhost:16657 --home test-1

gaiad config chain-id test-2 --home test-2
gaiad config keyring-backend test  --home test-2
gaiad config node http://localhost:26657 --home test-2
