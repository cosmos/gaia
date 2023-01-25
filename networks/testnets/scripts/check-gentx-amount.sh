#!/bin/bash

set -eo pipefail

path="$1"

declare -i maxbond=1000000000

extraquery='[.value.msg[]| select(.type != "cosmos-sdk/MsgCreateValidator")]|length'

gentxquery='.value.msg[]| select(.type == "cosmos-sdk/MsgCreateValidator")|.value.value'

denomquery="[$gentxquery | select(.denom != \"star\")] | length"

amountquery="$gentxquery | .amount"

# only allow MsgCreateValidator transactions.
if [ "$(jq "$extraquery" "$path")" != "0" ]; then
  echo "spurious transactions"
  exit 1
fi

# only allow "star" tokens to be bonded
if [ "$(jq "$denomquery" "$path")" != "0" ]; then
  echo "invalid denomination"
  exit 1
fi

# limit the amount that can be bonded
for amount in "$(jq -rM "$amountquery" "$path")"; do
  declare -i amt="$amount"
  if [ $amt -gt $maxbond ]; then
    echo "bonded too much: $amt > $maxbond"
    exit 1
  fi
done

exit 0