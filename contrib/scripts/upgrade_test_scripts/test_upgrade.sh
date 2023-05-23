#!/bin/bash

CNT=0
ITER=$1
SLEEP=$2
NUMBLOCKS=$3
NODEADDR=$4

if [ -z "$1" ]; then
  echo "Invalid argument: missing number of iterations"
  echo "sh test_upgrade.sh <iterations> <sleep> <num-blocks> <node-address>"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Invalid argument: missing sleep duration"
  echo "sh test_upgrade.sh <iterations> <sleep> <num-blocks> <node-address>"
  exit 1
fi

if [ -z "$3" ]; then
  echo "Invalid argument: missing number of blocks"
  echo "sh test_upgrade.sh <iterations> <sleep> <num-blocks> <node-address>"
  exit 1
fi

if [ -z "$4" ]; then
  echo "Invalid argument: missing node address"
  echo "sh test_upgrade.sh <iterations> <sleep> <num-blocks> <node-address>"
  exit 1
fi

echo "running 'sh test_upgrade.sh iterations=$ITER sleep=$SLEEP num-blocks=$NUMBLOCKS node-address=$NODEADDR'"

started=false
first_version=""

while [ ${CNT} -lt $ITER ]; do
  curr_block=$(curl -s $NODEADDR:26657/status | jq -r '.result.sync_info.latest_block_height')
  curr_version=$(curl -s $NODEADDR:1317/cosmos/base/tendermint/v1beta1/node_info | jq -r '.application_version.version')
  

  # tail v7.out

  if [[ $started = "false" && $curr_version != "" && $curr_version != "null" ]]; then
    started=true
    first_version=$curr_version
    echo "First version: ${first_version}"
  fi

  echo "count is ${CNT}, iteration ${ITER}, version is " $curr_version

  if [[ "$started" = "true" &&  ${curr_block} -gt ${NUMBLOCKS}  && "$curr_version" != $first_version && "$curr_version" != "" && "$curr_version" != "null" ]]; then
    echo "new version running"
    exit 0
  fi

  if [[ ${curr_block} -gt ${NUMBLOCKS} ]]; then
    echo "Failed: produced ${curr_block} without upgrading"
    exit 1
  fi


  CNT=$(($CNT+1))

  sleep $SLEEP
done

echo "Failed: timeout reached"
exit 1
