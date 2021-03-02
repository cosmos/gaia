# Upgrading to altheatestnet1v4

altheatestnet1v4 is the successor to the v3 testnet we brought online Sunday the 28th. Sadly after coming online that night it only ran for a few hours before hitting a slashing bug. I have confirmed that this new version does not have the same bug and I'm hopeful we can get online quickly and move into testing Gravity V2 features.

I've spent a significant amount of my time since the testnet went down refining the code that I was working on, so we should have a much better orchestrator experience, including automated waiting while the chain bootstraps.

If you are not a validator yet, please see [the instructions for setting up a validator](setting-up-a-validator.md)

When more than 66% of the voting power on Althea testnet returns the chain will start once again!

## Update your binaries

We have new releases of every binary, so take care to upgrade everything

In order to download ARM binaries change the names in the wget links from ‘client’ to ‘client-arm’. Repeat for all binaries. For the althea binary itself use -arm64 rather than amd.

```
cd althea-bin

# the althea chain binary itself
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/althea-0.0.3-4-g30eddc7-linux-amd64
mv althea-0.0.3-4-g30eddc7-linux-amd64 althea

# Tools for the gravity bridge from the gravity repo
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/client
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/orchestrator
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/register-delegate-keys
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/relayer
chmod +x *
sudo mv * /usr/bin/

```

## Update your genesis file

This is the exported genesis file of the chain history, we'll import it into our new updated chain keeping all balances and state

```
wget https://github.com/althea-net/althea-chain/releases/download/v0.0.4/althea-testnet1-v4-genesis.json
cp althea-testnet1-v4-genesis.json $HOME/.althea/config/genesis.json
```

## Start the chain

Unsafe reset all will reset the entire blockchain state in .althea allowing you to start althea-testnet1v4 using only the state from the genesis file

```
althea unsafe-reset-all
althea start
```

## Restart your Orchestrator

No argument changes are required, just ctrl-c and start it again.

The update orchestrator no longer needs to wait until after the chain has started, it will
stand ready until the chain itself starts.

If your orchestrator is working correctly you'll see a message like this

```
[2021-03-02T21:38:47Z INFO  peggy_utils::connection_prep] Cosmos node is syncing or waiting for the chain to start. Standing by
```

You may also want to check the status of your Geth node, no changes are required there, just make sure it's online.

If your Geth node is working correctly you'll see a message like this every few seconds

```
INFO [03-02|21:37:59.043] Imported new block headers               count=1 elapsed="347.865µs" number=4376135 hash="e05085…22afab"
```

## Wait for it

wait at this step for the chain to finish starting up again

If you've done the upgrade right expect to see this line

```
11:20PM INF Inbound Peer rejected err="incompatible: peer is on a different network. Got althea-testnet1v3, expected althea-testnet1v4" module=p2p numPeers=11
```

You are expecting the updated version of the chain. If you see 'expected althea-testnet1v3' then you are still running the out of date software and should double check these instructions.

Congrats you've finished upgrading! Keep an eye out for the upgrade signing progress message. You can safely leave your node unattended and everything should start when the chain starts
