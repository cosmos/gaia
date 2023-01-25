#!/bin/bash
# Set up a service to join the consumer-1 chain.

# Configuration
# You should only have to modify the values in this block
# * Keys
#    The private validator key and node key operations must match those used in the provider-1 chain if you want to run a validator.
# ***
PRIV_VALIDATOR_KEY_FILE=~/priv_validator_key.json
NODE_KEY_FILE=~/node_key.json
NODE_HOME=~/.consumer-1
NODE_MONIKER=consumer-1
SERVICE_NAME=consumer-1
STATE_SYNC=false
# ***

CHAIN_BINARY='interchain-security-cd'
CHAIN_ID=consumer-1
SEEDS="08ec17e86dac67b9da70deb20177655495a55407@consumer-1-seed-01.rs-testnet.polypore.xyz:26656,4ea6e56300a2f37b90e58de5ee27d1c9065cf871@consumer-1-seed-02.rs-testnet.polypore.xyz:26656,"
SYNC_RPC_1=http://consumer-1-state-sync-01.rs-testnet.polypore.xyz:26657
SYNC_RPC_2=http://consumer-1-state-sync-02.rs-testnet.polypore.xyz:26657
SYNC_RPC_SERVERS="$SYNC_RPC_1,$SYNC_RPC_2"

# The genesis file that includes the CCV state will not be published until after the spawn time has been reached.
# GENESIS_URL=https://github.com/cosmos/testnets/raw/master/replicated-security/consumer-1/consumer-1-genesis.json

# Install wget and jq
sudo apt-get install curl jq wget -y

# Install go 1.19.4
echo "Installing go..."
rm go*linux-amd64.tar.gz
wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install interchain-security-cd binary
echo "Installing build-essential..."
sudo apt install build-essential -y
echo "Installing interchain-security-cd..."
cd $HOME
mkdir -p $HOME/go/bin
rm -rf interchain-security
git clone https://github.com/cosmos/interchain-security.git
cd interchain-security
git checkout v1.0.0-rc3
make install
export PATH=$PATH:$HOME/go/bin

# Initialize home directory
echo "Initializing $NODE_HOME..."
rm -rf $NODE_HOME
$CHAIN_BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$CHAIN_BINARY config keyring-backend test --home $NODE_HOME
$CHAIN_BINARY config broadcast-mode block --home $NODE_HOME
$CHAIN_BINARY init $NODE_MONIKER --chain-id $CHAIN_ID --home $NODE_HOME
sed -i -e "/seeds =/ s^= .*^= \"$SEEDS\"^" $NODE_HOME/config/config.toml

# Replace keys
echo "Replacing keys..."
cp $PRIV_VALIDATOR_KEY_FILE $NODE_HOME/config/priv_validator_key.json
cp $NODE_KEY_FILE $NODE_HOME/config/node_key.json

# Replace genesis file: only after the spawn time is reached
# echo "Replacing genesis file..."
# wget $GENESIS_URL -O genesis.json
# mv genesis.json $NODE_HOME/config/genesis.json

if $STATE_SYNC ; then
    echo "Configuring state sync..."
    CURRENT_BLOCK=$(curl -s $SYNC_RPC_1/block | jq -r '.result.block.header.height')
    TRUST_HEIGHT=$[$CURRENT_BLOCK-1000]
    TRUST_BLOCK=$(curl -s $SYNC_RPC_1/block\?height\=$TRUST_HEIGHT)
    TRUST_HASH=$(echo $TRUST_BLOCK | jq -r '.result.block_id.hash')
    sed -i -e '/enable =/ s/= .*/= true/' $NODE_HOME/config/config.toml
    sed -i -e "/trust_height =/ s/= .*/= $TRUST_HEIGHT/" $NODE_HOME/config/config.toml
    sed -i -e "/trust_hash =/ s/= .*/= \"$TRUST_HASH\"/" $NODE_HOME/config/config.toml
    sed -i -e "/rpc_servers =/ s^= .*^= \"$SYNC_RPC_SERVERS\"^" $NODE_HOME/config/config.toml
else
    echo "Skipping state sync..."
fi

echo "Creating $SERVICE_NAME.service..."
sudo rm /etc/systemd/system/$SERVICE_NAME.service
sudo touch /etc/systemd/system/$SERVICE_NAME.service

echo "[Unit]"                               | sudo tee /etc/systemd/system/$SERVICE_NAME.service
echo "Description=consumer-1 service"       | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "After=network-online.target"          | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo ""                                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "[Service]"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "User=$USER"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "ExecStart=$HOME/go/bin/$CHAIN_BINARY start --x-crisis-skip-assert-invariants --home $NODE_HOME" | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "Restart=always"                       | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "RestartSec=3"                         | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "LimitNOFILE=4096"                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo ""                                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "[Install]"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "WantedBy=multi-user.target"           | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a

# Start service
sudo systemctl daemon-reload

# Enable and start the service after the genesis that includes the CCV state is in place
# sudo systemctl enable $SERVICE_NAME.service
# sudo systemctl start $SERVICE_NAME.service
# sudo systemctl restart systemd-journald

# Add go and gaiad to the path
echo "Setting up paths for go and interchain-security-cd bin..."
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile

echo "***********************"
echo "To see the service log enter:"
echo "journalctl -fu $SERVICE_NAME.service"
echo "***********************"
