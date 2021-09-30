#!/bin/bash

make install GAIA_BUILD_OPTIONS="cleveldb"

gaiad init "t6" --home ./t6 --chain-id t6

gaiad unsafe-reset-all --home ./t6

mkdir -p ./t6/data/snapshots/metadata.db

gaiad keys add validator --keyring-backend test --home ./t6

gaiad add-genesis-account $(gaiad keys show validator -a --keyring-backend test --home ./t6) 100000000stake --keyring-backend test --home ./t6

gaiad gentx validator 100000000stake --keyring-backend test --home ./t6 --chain-id t6

gaiad collect-gentxs --home ./t6

gaiad start --db_backend cleveldb --home ./t6
