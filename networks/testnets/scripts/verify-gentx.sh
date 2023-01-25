#!/bin/sh
DAEMON_HOME="/tmp/app$(date +%s)"
CLI_HOME="/tmp/appcli$(date +%s)"
DAEMON=akashd
CLI=akashctl
DENOM=star
RANDOM_KEY="randomvalidatorkeyxx"
CHAIN_ID=bigbang-1

GENTX_FILE=$(ls $CHAIN_ID/gentx | head -1)
LEN_GENTX=$(echo ${#GENTX_FILE})

GENTX_DEADLINE=$(date -d '2020-10-19 15:00:00' '+%d/%m/%Y %H:%M:%S');
now=$(date +"%d/%m/%Y %H:%M:%S")

# if [ $GENTX_DEADLINE < $now ]; then
#     echo 'Gentx submission is closed'
# el
if [ $LEN_GENTX -eq 0 ]; then
    echo "No new gentx file found."
else
    set -e

    ./scripts/check-gentx-amount.sh "./$CHAIN_ID/gentx/$GENTX_FILE" || exit 1

    echo "...........Install & Init Chain.............."
    curl -L https://github.com/ovrclk/akash/releases/download/v0.8.1/akash_0.8.1_linux_amd64.zip -o akash_linux.zip && unzip akash_linux.zip
    rm akash_linux.zip
    cd akash_0.8.1_linux_amd64

    echo "12345678" | ./$CLI keys add $RANDOM_KEY --keyring-backend test --home $CLI_HOME

    ./$DAEMON init --chain-id $CHAIN_ID dummyvalidator --home $DAEMON_HOME -o

    echo "..........Fetching genesis......."
    rm -rf $DAEMON_HOME/config/genesis.json
    curl -s https://raw.githubusercontent.com/cosmos/testnets/master/$CHAIN_ID/genesis.json > $DAEMON_HOME/config/genesis.json

    sed -i '/genesis_time/c\   \"genesis_time\" : \"2020-09-20T00:00:00Z\",' $DAEMON_HOME/config/genesis.json

    GENACC=$(cat ../$CHAIN_ID/gentx/$GENTX_FILE | sed -n 's|.*"delegator_address":"\([^"]*\)".*|\1|p')

    echo $GENACC

    echo "12345678" | ./$DAEMON add-genesis-account $RANDOM_KEY 1000000000000$DENOM --home $DAEMON_HOME \
        --keyring-backend test --home-client $CLI_HOME
    ./$DAEMON add-genesis-account $GENACC 1000000000$DENOM --home $DAEMON_HOME

    echo "12345678" | ./$DAEMON gentx --name $RANDOM_KEY --amount 900000000000$DENOM --home $DAEMON_HOME \
        --keyring-backend test --home-client $CLI_HOME
    cp ../$CHAIN_ID/gentx/$GENTX_FILE $DAEMON_HOME/config/gentx/

    echo "..........Collecting gentxs......."
    ./$DAEMON collect-gentxs --home $DAEMON_HOME
    sed -i '/persistent_peers =/c\persistent_peers = ""' $DAEMON_HOME/config/config.toml

    ./$DAEMON validate-genesis --home $DAEMON_HOME

    echo "..........Starting node......."
    ./$DAEMON start --home $DAEMON_HOME &

    sleep 5s

    echo "...checking network status.."

    ./$CLI status --chain-id $CHAIN_ID --node http://localhost:26657

    sleep 5s

    echo "...Cleaning the stuff..."
    killall $DAEMON >/dev/null 2>&1
    rm -rf $DAEMON_HOME >/dev/null 2>&1
fi
