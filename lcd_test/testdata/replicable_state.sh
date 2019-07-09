#!/usr/bin/env bash

PASSWORD="1234567890"
AMOUNT="1000000stake"
CHAIN="lcd"
PROPOSALID="2"
HOMEC="/tmp/contract_tests/.gaiacli"
HOMED="/tmp/contract_tests/.gaiad"
ACCOUNT="contract-tests"

./build/gaiad init ${ACCOUNT} --chain-id ${CHAIN} --home ${HOMED}

tar -xzf lcd_test/testdata/gaiad.tar.gz -C /tmp/contract_tests/
tar -xzf lcd_test/testdata/gaiacli.tar.gz -C /tmp/contract_tests/

SENDER=$(./build/gaiacli keys show sender --address --home ${HOMEC})
RECEIVER=$(./build/gaiacli keys show receiver --address --home ${HOMEC})
ADDR=$(./build/gaiacli keys show ${ACCOUNT} --address --home ${HOMEC})

./build/gaiad add-genesis-account ${ADDR} 1000000000000000000000000stake --home ${HOMED}
./build/gaiad add-genesis-account ${SENDER} 1000000000000000000000stake --home ${HOMED}
./build/gaiad add-genesis-account ${RECEIVER} 1000000000000000000000stake --home ${HOMED}
echo ${PASSWORD} | gaiad gentx --name ${ACCOUNT} --home ${HOMED} --home-client ${HOMEC}

./build/gaiad collect-gentxs --home ${HOMED}

# adjust inflation to see rewards
sed -i .bak -e 's/\"inflation\".*$/\"inflation\":\ \"0.000000001300000000\",/' -e 's/\"inflation_max\".*$/\"inflation_max\":\ \"0.000000002000000000\",/' -e 's/\"inflation_min\".*$/\"inflation_min\":\ \"0.000000000700000000\",/' -e 's/\"goal_bonded\".*$/\"goal_bonded\":\ \"0.000000006700000000\",/' "${HOMED}/config/genesis.json"
