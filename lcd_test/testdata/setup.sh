#!/usr/bin/env bash

PASSWORD="1234567890"
AMOUNT="1000000stake"
CHAIN="lcd"
PROPOSALID="2"
HOMEC="/tmp/contract_tests/.gaiacli"
GENESIS_PATH="/tmp/contract_tests/.gaiad/config/genesis.json"
SWAGGER='/tmp/contract_tests/swagger.yaml'
DUMMY_HASH='BCBE20E8D46758B96AE5883B792858296AC06E51435490FBDCAE25A72B3CC76B'
DUMMY_VALIDATOR='cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l'
VALIDATOR=$(awk '$1 ~ /"validator_address"/ {print $2}' ${GENESIS_PATH} | tr -d \", )
SENDER=$(./build/gaiacli keys show sender --address --home ${HOMEC})
RECEIVER=$(./build/gaiacli keys show receiver --address --home ${HOMEC})
# sleeping a whole second between each step is a conservative precaution
# check lcd_test/testdata/state.tar.gz -> .gaiad/config/config.toml precommit_timeout = 500ms

sleep 1s
echo "submit proposal"
echo ${PASSWORD} | ./build/gaiacli tx gov submit-proposal --from ${SENDER} --type text --title test --description test_description --deposit 10000stake --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo "delegate"
echo ${PASSWORD} | ./build/gaiacli tx staking delegate ${VALIDATOR} 1000stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo "unbond"
echo ${PASSWORD} | ./build/gaiacli tx staking unbond ${VALIDATOR} 100stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes

# Create, deposit and vote for a proposal
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov submit-proposal --from ${SENDER} --type text --title test --description test_description --deposit 10000stake --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov deposit ${PROPOSALID} 1000000000stake --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov vote ${PROPOSALID} Yes --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC}  --yes
sleep 1s

# make a transaction with known sender and receiver, replace this hash witith the existing one in the swagger file
HASH=$(echo ${PASSWORD} | ./build/gaiacli tx send ${SENDER} ${RECEIVER} ${AMOUNT} --chain-id ${CHAIN} --home ${HOMEC} --yes | awk '/txhash.*/{print $2}')
sleep 1s

# unbound from a validator
echo ${PASSWORD} | ./build/gaiacli tx staking unbond --from ${SENDER} ${VALIDATOR} 100stake --chain-id ${CHAIN} --home ${HOMEC} --yes

sleep 1s
echo "withdraw rewards"
echo ${PASSWORD} | ./build/gaiacli tx distribution withdraw-all-rewards --from ${SENDER} --chain-id ${CHAIN} --home ${HOMEC} --yes

# Replace dummy values in swagger with new hashes and addresses
sed -i.bak -e "s/${DUMMY_HASH}/${HASH}/g" "${SWAGGER}"
echo "Replaced ${DUMMY_HASH} with actual transaction hash ${HASH}"

sed -i.bak -e "s/${DUMMY_VALIDATOR}/${VALIDATOR}/g" "${SWAGGER}"
echo "Replaced ${DUMMY_VALIDATOR} with actual validator address ${VALIDATOR}"

