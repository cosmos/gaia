#!/bin/sh

set -o errexit -o nounset

UPGRADE_HEIGHT=$1

if [ -z "$1" ]; then
echo "Need to add an upgrade height"
exit 1
fi

NODE_HOME=$(realpath ./build/.gaia)

echo "NODE_HOME = ${NODE_HOME}"

BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"

$BINARY version

USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
CHAINID=cosmoshub-4

if test -f "$BINARY"; then

echo "wait 10 seconds for blockchain to start"
sleep 10

$BINARY config chain-id $CHAINID --home $NODE_HOME
$BINARY config output json --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME
$BINARY config --home $NODE_HOME

key=$($BINARY keys show val --home $NODE_HOME)
if [ -z "$key" ]; then
echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test
fi


echo "\n"
echo "Submitting proposal... \n"
$BINARY tx gov submit-proposal software-upgrade v12 \
--title v12 \
--deposit 10000000uatom \
--upgrade-height $UPGRADE_HEIGHT \
--upgrade-info "upgrade to v12" \
--description "upgrade to v12" \
--fees 400uatom \
--from val \
--keyring-backend test \
--chain-id $CHAINID \
--home $NODE_HOME \
--node tcp://localhost:26657 \
--yes
echo "Done \n"

sleep 6
echo "Casting vote... \n"

$BINARY tx gov vote 1 yes \
--from val \
--keyring-backend test \
--chain-id $CHAINID \
--home $NODE_HOME \
--fees 400uatom \
--node tcp://localhost:26657 \
--yes

echo "Done \n"

else
echo "Please build gaia v11 and move to ./build/gaiad11"
fi
