#!/bin/bash
set -eu

althea init --chain-id=testing local
althea add-genesis-account validator 1000000000ualtg
althea gentx --name validator  --amount 1000000000ualtg
althea collect-gentxs
