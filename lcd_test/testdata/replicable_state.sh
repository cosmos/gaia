#!/usr/bin/env bash

PASSWORD="1234567890"
AMOUNT="1000000stake"
CHAIN="lcd"
PROPOSALID="2"
HOMEC="/tmp/contract_tests/.gaiacli"
HOMED="/tmp/contract_tests/.gaiad"
ACCOUNT="contract-tests"

rm -rf /tmp/contract_tests/
mkdir /tmp/contract_tests

gaiad init ${ACCOUNT} --chain-id ${CHAIN} --home ${HOMED}

tar -xzf lcd_test/testdata/gaiad.tar.gz -C /tmp/contract_tests/
tar -xzf lcd_test/testdata/gaiacli.tar.gz -C /tmp/contract_tests/

SENDER=$(gaiacli keys show sender --address --home ${HOMEC})
RECEIVER=$(gaiacli keys show receiver --address --home ${HOMEC})
ADDR=$(gaiacli keys show ${ACCOUNT} --address --home ${HOMEC})

gaiad add-genesis-account ${ADDR} 1000000000000000000000000stake --home ${HOMED}
gaiad add-genesis-account ${SENDER} 1000000000000000000000stake --home ${HOMED}
gaiad add-genesis-account ${RECEIVER} 1000000000000000000000stake --home ${HOMED}
echo ${PASSWORD} | gaiad gentx --name ${ACCOUNT} --home ${HOMED} --home-client ${HOMEC}

gaiad collect-gentxs --home ${HOMED}

# shorten timeouts in config.toml
./lcd_test/testdata/config.sh

# adjust inflation to see rewards
sed -i .bak -e 's/\"inflation\".*$/\"inflation\":\ \"0.000000001300000000\",/' -e 's/\"inflation_max\".*$/\"inflation_max\":\ \"0.000000002000000000\",/' -e 's/\"inflation_min\".*$/\"inflation_min\":\ \"0.000000000700000000\",/' -e 's/\"goal_bonded\".*$/\"goal_bonded\":\ \"0.000000006700000000\",/' "${HOMED}/config/genesis.json"
