#!/usr/bin/env bash

CONFIG=/tmp/contract_tests/.gaiad/config/config.toml

# Tune config.toml in order to speed up tests, by reducing timeouts
sed -i.bak -e "s/^timeout_propose = .*/timeout_propose = \"200ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_propose_delta = .*/timeout_propose_delta = \"200ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_prevote = .*/timeout_prevote = \"500ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_prevote_delta = .*/timeout_prevote_delta = \"100ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_precommit = .*/timeout_precommit = \"200ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_precommit_delta = .*/timeout_precommit_delta = \"200ms\"/g" ${CONFIG}
sed -i.bak -e "s/^timeout_commit = .*/timeout_commit = \"500ms\"/g" ${CONFIG}
