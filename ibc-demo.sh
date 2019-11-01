#!/bin/zsh

GAIA_BRANCH=cwgoes/ibc-demo-fixes
GAIA_DIR=$(mktemp -d)
CONF_DIR=$(mktemp -d)

echo "GAIA_DIR: ${GAIA_DIR}"
echo "CONF_DIR: ${CONF_DIR}"

sleep 1

set -x

echo "Killing existing gaiad instances..."

killall gaiad

set -e

echo "Building Gaia..."

cd $GAIA_DIR
git clone git@github.com:cosmos/gaia
cd gaia
git checkout $GAIA_BRANCH
make install
gaiad version
gaiacli version

echo "Generating configurations..."

cd $CONF_DIR && mkdir ibc-testnets && cd ibc-testnets
echo -e "\n" | gaiad testnet -o ibc0 --v 1 --chain-id ibc0 --node-dir-prefix n
echo -e "\n" | gaiad testnet -o ibc1 --v 1 --chain-id ibc1 --node-dir-prefix n

if [ "$(uname)" = "Linux" ]; then
  sed -i 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
  sed -i 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml
  sed -i 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:26556"#g' ibc1/n0/gaiad/config/config.toml
  sed -i 's#"tcp://0.0.0.0:26657"#"tcp://0.0.0.0:26557"#g' ibc1/n0/gaiad/config/config.toml
  sed -i 's#"localhost:6060"#"localhost:6061"#g' ibc1/n0/gaiad/config/config.toml
  sed -i 's#"tcp://127.0.0.1:26658"#"tcp://127.0.0.1:26558"#g' ibc1/n0/gaiad/config/config.toml
else
  sed -i '' 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
  sed -i '' 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml
  sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:26556"#g' ibc1/n0/gaiad/config/config.toml
  sed -i '' 's#"tcp://0.0.0.0:26657"#"tcp://0.0.0.0:26557"#g' ibc1/n0/gaiad/config/config.toml
  sed -i '' 's#"localhost:6060"#"localhost:6061"#g' ibc1/n0/gaiad/config/config.toml
  sed -i '' 's#"tcp://127.0.0.1:26658"#"tcp://127.0.0.1:26558"#g' ibc1/n0/gaiad/config/config.toml
fi;

gaiacli config --home ibc0/n0/gaiacli/ chain-id ibc0
gaiacli config --home ibc1/n0/gaiacli/ chain-id ibc1
gaiacli config --home ibc0/n0/gaiacli/ output json
gaiacli config --home ibc1/n0/gaiacli/ output json
gaiacli config --home ibc0/n0/gaiacli/ node http://localhost:26657
gaiacli config --home ibc1/n0/gaiacli/ node http://localhost:26557

echo "Importing keys..."

SEED0=$(jq -r '.secret' ibc0/n0/gaiacli/key_seed.json)
SEED1=$(jq -r '.secret' ibc1/n0/gaiacli/key_seed.json)
echo -e "12345678\n" | gaiacli --home ibc1/n0/gaiacli keys delete n0

echo "Seed 0: ${SEED0}"
echo "Seed 1: ${SEED1}"

echo "Enter seed 1:"
gaiacli --home ibc0/n0/gaiacli keys add n1 --recover

echo "Enter seed 0:"
gaiacli --home ibc1/n0/gaiacli keys add n0 --recover

echo "Enter seed 1:"
gaiacli --home ibc1/n0/gaiacli keys add n1 --recover

#echo -e "12345678\n12345678\n$SEED1\n" | gaiacli --home ibc0/n0/gaiacli keys add n1 --recover
#echo -e "12345678\n12345678\n$SEED0\n" | gaiacli --home ibc1/n0/gaiacli keys add n0 --recover
#echo -e "12345678\n12345678\n$SEED1\n" | gaiacli --home ibc1/n0/gaiacli keys add n1 --recover

echo "Keys should match:"

gaiacli --home ibc0/n0/gaiacli keys list | jq '.[].address'
gaiacli --home ibc1/n0/gaiacli keys list | jq '.[].address'

DEST=$(gaiacli --home ibc0/n0/gaiacli keys list | jq -r '.[1].address')

echo "Destination: $DEST"

echo "Starting Gaiad instances..."

nohup gaiad --home ibc0/n0/gaiad --log_level="*:debug" start > ibc0.log &
nohup gaiad --home ibc1/n0/gaiad --log_level="*:debug" start > ibc1.log &

sleep 20

echo "Creating clients..."

echo -e "12345678\n" | gaiacli --home ibc0/n0/gaiacli \
  tx ibc client create ibconeclient \
  $(gaiacli --home ibc1/n0/gaiacli q ibc client node-state) \
  --from n0 -y -o text

echo -e "12345678\n" | gaiacli --home ibc1/n0/gaiacli \
  tx ibc client create ibczeroclient \
  $(gaiacli --home ibc0/n0/gaiacli q ibc client node-state) \
  --from n1 -y -o text

sleep 3

echo "Establishing a connection..."

gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc connection handshake \
  connectionzero ibconeclient $(gaiacli --home ibc1/n0/gaiacli q ibc client path) \
  connectionone ibczeroclient $(gaiacli --home ibc0/n0/gaiacli q ibc client path) \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557

echo "Establishing a channel..."

gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc channel handshake \
  ibconeclient bank channelzero connectionzero \
  ibczeroclient bank channelone connectionone \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557 \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1

echo "Sending token packets from ibc0..."

gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc transfer transfer \
  bankbankbank channelzero \
  $DEST 1stake

echo "Recieving token packets on ibc1..."

gaiacli \
  --home ibc1/n0/gaiacli \
  tx ibc transfer recv-packet \
  bankbankbank channelone \
  packet.json \
  proof.json
