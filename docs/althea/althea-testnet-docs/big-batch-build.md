# Stress test the Gravity bridge!

As part of our testnet we're going to try and put the maximum possible load on the bridge, Sunday February 21st at around 1pm west coast time. This isn't a hard time event, you can simply setup your client to send thousands of transactions and leave it alone for the rest of the day. We just want to get as many people as possible doing so.

This guide does not require that you be a validator on althea-testnet1v5, it is designed to run on any machine.

## Download Althea chain and the Gravity tools

The client binary has been updated for this guide.

```
mkdir althea-bin
cd althea-bin

# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/althea-0.0.4-16-g6812f87-linux-amd64
mv althea-0.0.4-16-g6812f87-linux-amd64 althea

# Tools for the gravity bridge from the gravity repo
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/client
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/orchestrator
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/register-delegate-keys
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.5/relayer
chmod +x *
sudo mv * /usr/bin/
```

## Generate your Althea testnet Cosmos address

If you are a validator you can skip this step as you already have an address ready. Or you can generate a new address there's not much real difference. Just remember to keep track of your seed phrases.

```
cd $HOME
althea init mymoniker --chain-id althea-testnet1v5
althea keys add mytestingname
```

## Request some ERC20 tokens on Cosmos

Paste your Cosmos key in the chat and request some ERC20 to test with (you may already have some if you've been testing as a validator)

## Build a big batch

The goal of our test is to build 3 large transaction batches (over 100 transactions) and then have them
flow through the Gravity bridge all at once. When you request tokens I'll send you 100 of each type of ERC20 which you'll then use these commands to send back. They will take quite some time to run, that's intentional we want to have the largest number of different people sending transactions at once, not the most spam a single person can send at once

```
RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc.althea.net:1317"  \
        --erc20-address="0xD7600ae27C99988A6CD360234062b540F88ECA43" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"

RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc.althea.net:1317"  \
        --erc20-address="0x7580bFE88Dd3d07947908FAE12d95872a260F2D8" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"

RUST_LOG=info client cosmos-to-eth \
        --cosmos-phrase="the cosmos phrase from before" \
        --cosmos-rpc="http://testnet1-rpc.althea.net:1317"  \
        --erc20-address="0xD50c0953a99325d01cca655E57070F1be4983b6b" \
        --amount=.5 \
        --times=200 \
        --no-batch \
        --eth-destination="any eth address, try your delegate eth address"
```

## Wait for it

the above commands have 'no-batch' set, which means the funds won't show up on Ethereum instead they are waiting for a relayer to deem the batch profitable enough to relay and request it. This is where our scheduled time comes in. I'll request batches for all three tokens at around 1pm West coast time.
