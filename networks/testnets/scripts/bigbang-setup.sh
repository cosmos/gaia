#!/bin/bash

command_exists () {
    type "$1" &> /dev/null ;
}

if command_exists go ; then
    echo "Golang is already installed"
else
  echo "Install dependencies"
  sudo apt update
  sudo apt install build-essential jq -y

  wget https://dl.google.com/go/go1.15.5.linux-amd64.tar.gz
  tar -xvf go1.15.5.linux-amd64.tar.gz
  sudo mv go /usr/local

  echo "" >> ~/.bashrc
  echo 'export GOPATH=$HOME/go' >> ~/.bashrc
  echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
  echo 'export GOBIN=$GOPATH/bin' >> ~/.bashrc
  echo 'export PATH=$PATH:/usr/local/go/bin:$GOBIN' >> ~/.bashrc

  #source ~/.bashrc
  . ~/.bashrc
  
  
fi

echo "-- Stopping any previous system service of akashd"

sudo systemctl stop akashd

akashd unsafe-reset-all

echo "-- Clear old akash data and install akashd and setup the node --"

rm -rf $GOBIN/akashctl
rm -rf $GOBIN/akashd
rm -rf ~/.akashd
rm -rf ~/.akashctl
rm -rf ~/akash

YOUR_KEY_NAME=$1
YOUR_NAME=$2
DAEMON=akashd
DENOM=uakt
CHAIN_ID=bigbang-3
PERSISTENT_PEERS="6205fb3c05d0dccb5451507112601af454f4d059@104.131.69.13:26656"

echo "installing akashd"
git clone https://github.com/ovrclk/akash.git
cd akash
git fetch
git checkout bigbang
make install

echo "Creating keys"
$DAEMON keys add $YOUR_KEY_NAME

echo "Setting up your validator"
$DAEMON init $YOUR_NAME --chain-id $CHAIN_ID 
curl http://104.131.69.13:26657/genesis | jq .result.genesis > ~/.$DAEMON/config/genesis.json

echo "----------Setting config for seed node---------"
sed -i 's#tcp://127.0.0.1:26657#tcp://0.0.0.0:26657#g' ~/.$DAEMON/config/config.toml
sed -i '/persistent_peers =/c\persistent_peers = "'"$PERSISTENT_PEERS"'"' ~/.$DAEMON/config/config.toml


echo "---------Creating system file---------"

echo "[Unit]
Description=Akashd daemon
After=network-online.target
[Service]
User=${USER}
ExecStart=${GOBIN}/$DAEMON start
Restart=always
RestartSec=3
LimitNOFILE=4096
[Install]
WantedBy=multi-user.target
" > akashd.service

sudo mv akashd.service /etc/systemd/system/akashd.service
sudo systemctl daemon-reload
sudo systemctl start akashd

echo
echo "Your account address is :"
$DAEMON keys show $YOUR_KEY_NAME -a
echo "Your node setup is done. You would need some tokens to start your validator. You can get some tokens from the faucet: https://faucet.bigbang.vitwit.com"
echo
echo
echo "After receiving tokens, you can create your validator by running"
echo "$DAEMON tx staking create-validator --amount 90000000$DENOM --commission-max-change-rate \"0.1\" --commission-max-rate \"0.20\" --commission-rate \"0.1\" --details \"Some details about yourvalidator\" --from $YOUR_KEY_NAME   --pubkey=\"$($DAEMON tendermint show-validator)\" --moniker $YOUR_NAME --min-self-delegation \"1\" --chain-id $CHAIN_ID"
