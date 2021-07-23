# Preparing for the Althea testnet1v5 upgrade

This Wednesday (March 17th 2021) the Gorli Ethereum testnet will be performing the Berlin hardfork. This Sunday (March 21st 2021) we will be upgrading to Althea testnet1v5. Big changes in v5 will include.

1. Fixes for sudden slashing
2. Registration for delegate keys in the 'althea' binary (allowing for Ledeger key use)
3. Fix for the invalid signatures orchestrator bug
4. Other improvements to slashing logic
5. Cosmos security fixes

While we are targeting 1pm US Pacific time for this halt and upgrade, but it may vary. For this upgrade we won't be meeting in a chat channel.

## Upgrade your Ethereum Node

Stop your existing geth node using ctrl-C

Then download and start the new one

```
wget https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.10.6-576681f2.tar.gz
tar -xvf geth-linux-amd64-1.10.6-576681f2.tar.gz
cd geth-linux-amd64-1.10.6-576681f2
./geth --syncmode "light" --goerli --http --cache 16
```

It may take a minute for Geth to start syncing blocks again. You'll see a message like this every few seconds

```
INFO [03-15|17:19:44.413] Imported new block headers               count=1 elapsed="130.872µs" number=4446212 hash="74914e…ebd43f"
```

Once that happens you can move on

## Prepare your Cosmos node to halt

We are targeting block height `455927` to halt

run ctrl-c on the terminal with your Althea blockchain node, then restart using

```
althea start --halt-height=455927
```

You should see blocks being made again after only a few seconds.

## Double check your Orchestrator

Finally make sure that your orchestrator is in good shape by restarting it.

If the Orchestrator makes it as far as printing this message, then you're in good shape

```
[2021-03-15T17:25:47Z INFO  orchestrator::main_loop] Oracle resync complete, Oracle now operational
```

## All done

Thank you for participating in the Althea blockchain testnet. I'm looking forward to wrapping up our Gravity fixes and moving onto Althea blockchain specific features within one or two upgrades.
