#!/usr/bin/env bash

PASSWORD="1234567890"
SENDER="cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv"
RECEIVER="cosmos17gx5vwpm0y2k59tw0x00ccug234n56cgltx2w2"
AMOUNT="1000000stake"
CHAIN="lcd"
PROPOSALID="2"
HOMEC="/tmp/contract_tests/.gaiacli"
GENESIS_PATH="/tmp/contract_tests/.gaiad/config/genesis.json"
SWAGGER='/tmp/contract_tests/swagger.yaml'
VALIDATOR=$(awk '$1 ~ /"validator_address"/ {print $2}' ${GENESIS_PATH} | tr -d \", )
SENDER=$(gaiacli keys show sender --address --home ${HOMEC})
RECEIVER=$(gaiacli keys show receiver --address --home ${HOMEC})
# sleeping a whole second between each step is a conservative precaution
# check lcd_test/testdata/state.tar.gz -> .gaiad/config/config.toml precommit_timeout = 500ms

sleep 1s
echo "submit proposal"
echo ${PASSWORD} | gaiacli tx gov submit-proposal --from ${SENDER} --chain-id ${CHAIN} --type text --title test --description test_description --deposit 10000stake --home ${HOMEC} --yes
sleep 1s
echo "delegate"
echo ${PASSWORD} | gaiacli tx staking delegate ${VALIDATOR} 1000stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo "unbond"
echo ${PASSWORD} | gaiacli tx staking unbond ${VALIDATOR} 100stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes

# Create, deposit and vote for a proposal
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov submit-proposal --from ${SENDER} --type text --title test --description test_description --deposit 10000stake --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov deposit ${PROPOSALID} 1000000000stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov vote ${PROPOSALID} Yes --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC}  --yes
sleep 1s

# make a transaction with known sender and receiver, replace this hash witith the existing one in the swagger file
HASH=$(echo ${PASSWORD} | ./build/gaiacli tx send --home ${HOMEC} ${SENDER} ${RECEIVER} ${AMOUNT} --yes --chain-id ${CHAIN} | awk '/txhash.*/{print $2}')
sed -i.bak -e "s/BCBE20E8D46758B96AE5883B792858296AC06E51435490FBDCAE25A72B3CC76B/${HASH}/g" "${SWAGGER}"
echo "Replaced dummy with actual transaction hash ${HASH}"
sleep 1s

# unbound from a validator
echo ${PASSWORD} | ./build/gaiacli tx staking unbond --home ${HOMEC} --from ${SENDER} ${VALIDATOR} 100stake --yes --chain-id ${CHAIN}

sleep 1s
echo "withdraw rewards"
echo ${PASSWORD} | gaiacli tx distribution withdraw-all-rewards --chain-id ${CHAIN} --from ${SENDER} --home ${HOMEC}
