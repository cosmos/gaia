#!/bin/bash
# Set up a Gaia service to join the Cosmos Hub public testnet.

# Configuration
# You should only have to modify the values in this block
# ***
NODE_HOME=~/.gaia
NODE_MONIKER=public-testnet
SERVICE_NAME=gaiad
CHAIN_BINARY_URL='https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-linux-amd64'
STATE_SYNC=true
# ***

CHAIN_BINARY='gaiad'
CHAIN_ID=theta-testnet-001
GENESIS_ZIPPED_URL=https://github.com/cosmos/testnets/raw/master/public/genesis.json.gz
SEEDS="639d50339d7045436c756a042906b9a69970913f@seed-01.theta-testnet.polypore.xyz:26656,3e506472683ceb7ed75c1578d092c79785c27857@seed-02.theta-testnet.polypore.xyz:26656"
SYNC_RPC="https://rpc.state-sync-01.theta-testnet.polypore.xyz:443,https://rpc.state-sync-02.theta-testnet.polypore.xyz:443"

# Install wget and jq
sudo apt-get install curl jq wget -y

# Install go 1.19.4
echo "Installing go..."
rm go*linux-amd64.tar.gz
wget https://go.dev/dl/go1.19.4.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.4.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Gaia binary
echo "Installing Gaia..."
mkdir -p $HOME/go/bin

# Download Linux amd64,
wget $CHAIN_BINARY_URL -O $HOME/go/bin/$CHAIN_BINARY
chmod +x $HOME/go/bin/$CHAIN_BINARY

# or build from source
# echo "Installing build-essential..."
# sudo apt install build-essential -y
# echo "Installing Gaia..."
# rm -rf gaia
# git clone https://github.com/cosmos/gaia.git
# cd gaia
# git checkout v8.0.0-rc3
# make install

export PATH=$PATH:$HOME/go/bin

# Initialize home directory
echo "Initializing $NODE_HOME..."
rm -rf $NODE_HOME
$CHAIN_BINARY config chain-id $CHAIN_ID --home $NODE_HOME
$CHAIN_BINARY config keyring-backend test --home $NODE_HOME
$CHAIN_BINARY config broadcast-mode block --home $NODE_HOME
$CHAIN_BINARY init $NODE_MONIKER --chain-id $CHAIN_ID --home $NODE_HOME

if $STATE_SYNC ; then
    echo "Configuring state sync..."
    CURRENT_BLOCK=$(curl -s https://rpc.sentry-01.theta-testnet.polypore.xyz/block | jq -r '.result.block.header.height')
    TRUST_HEIGHT=$[$CURRENT_BLOCK-1000]
    TRUST_BLOCK=$(curl -s https://rpc.sentry-01.theta-testnet.polypore.xyz/block\?height\=$TRUST_HEIGHT)
    TRUST_HASH=$(echo $TRUST_BLOCK | jq -r '.result.block_id.hash')
    sed -i -e '/enable =/ s/= .*/= true/' $NODE_HOME/config/config.toml
    sed -i -e "/trust_height =/ s/= .*/= $TRUST_HEIGHT/" $NODE_HOME/config/config.toml
    sed -i -e "/trust_hash =/ s/= .*/= \"$TRUST_HASH\"/" $NODE_HOME/config/config.toml
    sed -i -e "/rpc_servers =/ s^= .*^= \"$SYNC_RPC\"^" $NODE_HOME/config/config.toml
else
    echo "Skipping state sync..."
fi

# Replace genesis file
echo "Downloading genesis file..."
wget $GENESIS_ZIPPED_URL
gunzip genesis.json.gz -f
cp genesis.json $NODE_HOME/config/genesis.json

sudo rm /etc/systemd/system/$SERVICE_NAME.service
sudo touch /etc/systemd/system/$SERVICE_NAME.service

echo "[Unit]"                               | sudo tee /etc/systemd/system/$SERVICE_NAME.service
echo "Description=Gaia service"       | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "After=network-online.target"          | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo ""                                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "[Service]"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "User=$USER"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "ExecStart=$HOME/go/bin/$CHAIN_BINARY start --x-crisis-skip-assert-invariants --home $NODE_HOME --p2p.seeds $SEEDS" | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "Restart=always"                       | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "RestartSec=3"                         | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "LimitNOFILE=4096"                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo ""                                     | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "[Install]"                            | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a
echo "WantedBy=multi-user.target"           | sudo tee /etc/systemd/system/$SERVICE_NAME.service -a

# Start service
echo "Starting $SERVICE_NAME.service..."
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME.service
sudo systemctl start $SERVICE_NAME.service
sudo systemctl restart systemd-journald

# Add go and gaiad to the path
echo "Setting up paths for go and gaiad bin..."
echo "export PATH=$PATH:/usr/local/go/bin" >> .profile

echo "***********************"
echo "To see the Gaia log enter:"
echo "journalctl -fu $SERVICE_NAME.service"
echo "***********************"
