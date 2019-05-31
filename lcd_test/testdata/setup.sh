#!/usr/bin/env bash

PASSWORD="1234567890"
ADDR="cosmos16xyempempp92x9hyzz9wrgf94r6j9h5f06pxxv"
VALIDATOR="cosmosvaloper16xyempempp92x9hyzz9wrgf94r6j9h5f2w4n2l"
CHAIN="lcd"
GOVID="2"
HOME="/tmp/.gaiacli"

# luckily governance are down in the swagger sequence of calls, this 15s are massive sleep time
# TODO: find out why the signature verification still fails without sleeps
# 3 seconds works sometims, 4 seconds often, 5 always but is huge!
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov submit-proposal --home ${HOME} --from ${ADDR} --chain-id ${CHAIN} --type text --title test --description test_description --deposit 10000stake --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov deposit --home ${HOME} --from ${ADDR} --chain-id ${CHAIN} ${GOVID} 1000000000stake --yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx gov vote --home ${HOME} --from ${ADDR} --yes --chain-id ${CHAIN} ${GOVID} Yes
sleep 1s
echo ${PASSWORD} | ./build/gaiacli tx staking unbond --home ${HOME} --from ${ADDR} ${VALIDATOR} 100stake --yes --chain-id ${CHAIN}
