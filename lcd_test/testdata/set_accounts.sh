#!/usr/bin/env bash

# do not run this script, it's just the way those static addresses were created
PASSWORD="1234567890"
HOMEC="/tmp/contract_tests/.gaiacli"
ACCOUNT="contract-tests"

echo ${PASSWORD} | gaiacli keys add ${ACCOUNT} --home ${HOMEC}
echo ${PASSWORD} | gaiacli keys add sender --home ${HOMEC}
echo ${PASSWORD} | gaiacli keys add receiver --home ${HOMEC}
