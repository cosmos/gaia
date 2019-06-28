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

echo ${PASSWORD} | gaiacli keys add ${ACCOUNT} --home ${HOMEC}
echo ${PASSWORD} | gaiacli keys add sender --home ${HOMEC}
echo ${PASSWORD} | gaiacli keys add receiver --home ${HOMEC}

SENDER=$(gaiacli keys show sender --address --home ${HOMEC})
RECEIVER=$(gaiacli keys show receiver --address --home ${HOMEC})
ADDR=$(gaiacli keys show ${ACCOUNT} --address --home ${HOMEC})

gaiad add-genesis-account ${ADDR} 1000000000000000000000000stake --home ${HOMED}
gaiad add-genesis-account ${SENDER} 1000000000000000000000stake --home ${HOMED}
gaiad add-genesis-account ${RECEIVER} 1000000000000000000000stake --home ${HOMED}
echo ${PASSWORD} | gaiad gentx --name ${ACCOUNT} --home ${HOMED} --home-client ${HOMEC}

gaiad collect-gentxs --home ${HOMED}

./lcd_test/testdata/config.sh

sed -i .bak -e 's/\"inflation\".*$/\"inflation\":\ \"0.000000001300000000\",/' -e 's/\"inflation_max\".*$/\"inflation_max\":\ \"0.000000002000000000\",/' -e 's/\"inflation_min\".*$/\"inflation_min\":\ \"0.000000000700000000\",/' -e 's/\"goal_bonded\".*$/\"goal_bonded\":\ \"0.000000006700000000\",/' "${HOMED}/config/genesis.json"

gaiad start --home ${HOMED} &

sleep 1s
echo "submit proposal"
echo ${PASSWORD} | gaiacli tx gov submit-proposal --from ${SENDER} --chain-id ${CHAIN} --type text --title test --description test_description --deposit 10000stake --home ${HOMEC} --yes
sleep 1s
echo "rewards"
echo ${PASSWORD} | gaiacli tx distr withdraw-all-rewards --chain-id ${CHAIN} --yes --from ${SENDER} --home ${HOMEC}

VALIDATOR=$(gaiad tendermint show-validator)

#sleep 1s
#echo "delegate"
#echo ${PASSWORD} | gaiacli tx staking delegate ${VALIDATOR} 1000stake --from ${SENDER} --yes --home ${HOMEC}
#sleep 1s
#echo "unbond"
#echo ${PASSWORD} | gaiacli tx staking unbond --home ${HOMEC} --from ${SENDER} ${VALIDATOR} 100stake --yes --chain-id ${CHAIN}
