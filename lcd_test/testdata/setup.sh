#!/usr/bin/env bash

PASSWORD="1234567890"
SENDER="cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv"
RECEIVER="cosmos17gx5vwpm0y2k59tw0x00ccug234n56cgltx2w2"
VALIDATOR="cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l"
AMOUNT="1000000stake"
CHAIN="lcd"
PROPOSALID="2"
HOME="/tmp/contract_tests/.gaiacli"
SWAGGER='/tmp/contract_tests/swagger.yaml'

# sleeping a whole second between each step is a conservative precaution
# check lcd_test/testdata/state.tar.gz -> .gaiad/config/config.toml precommit_timeout = 500ms

# Create, deposit and vote for a proposal
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov submit-proposal --home ${HOME} --from ${SENDER} --chain-id ${CHAIN} --type text --title test --description test_description --deposit 10000stake --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov deposit --home ${HOME} --from ${SENDER} --chain-id ${CHAIN} ${PROPOSALID} 1000000000stake --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov vote --home ${HOME} --from ${SENDER} --yes --chain-id ${CHAIN} ${PROPOSALID} Yes
sleep 1s

# make a transaction with known sender and receiver, replace this hash witith the existing one in the swagger file
HASH=$(echo ${PASSWORD} | ./build/gaiacli tx send --home ${HOME} ${SENDER} ${RECEIVER} ${AMOUNT} --yes --chain-id ${CHAIN} | awk '/txhash.*/{print $2}')
sed -i.bak -e "s/BCBE20E8D46758B96AE5883B792858296AC06E51435490FBDCAE25A72B3CC76B/${HASH}/g" "${SWAGGER}"
echo "Replaced dummy with actual transaction hash ${HASH}"
sleep 1s

# unbound from a validator
echo ${PASSWORD} | ./build/gaiacli tx staking unbond --home ${HOME} --from ${SENDER} ${VALIDATOR} 100stake --yes --chain-id ${CHAIN}

