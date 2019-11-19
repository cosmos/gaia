# Upgrade Module

## Start the network and trigger upgrade

```bash
# Start the testnet
gaiad start

# Set up the cli config
gaiacli config trust-node true
gaiacli config chain-id testing

# Create a proposal
gaiacli tx gov submit-proposal software-upgrade test1 \
--title "cosmoshub-x" --description "upgrade to latest Gaia release" --from validator \
--upgrade-height 200 --deposit 10000000stake -y

# Once the proposal passes you can query the pending plan
gaiacli query upgrade plan
```

## Performing an upgrade

Assuming the proposal passes the chain will stop at given upgrade height.

You can stop and start the original binary all you want, but **it will refuse to
run after the upgrade height**.

We need a new binary with the upgrade handler installed. The logs should look
something like:

```shell
E[2019-11-05|12:44:18.913] UPGRADE "cosmoshub-x" NEEDED at height: 200:       module=main
E[2019-11-05|12:44:18.914] CONSENSUS FAILURE!!!
...
```

Note that the process just hangs, doesn't exit to avoid restart loops. You must
manually kill the process and replace it with a new binary. Do so now with
`Ctrl+C` or `killall gaiad`.

In `gaia/app/app.go`, add the following lines after line 169 (once
`app.upgradeKeeper` is initialized). Make sure the upgrade title in handler matches the title from proposal.

```go
    app.upgradeKeeper.SetUpgradeHandler("cosmoshub-x", func(ctx sdk.Context, plan upgrade.Plan) {
        // custom logic after the network upgrade has been executed
    })
```

Note that we panic on any error - this would cause the upgrade to fail if the
migration could not be run, and no node would advance - allowing a manual recovery.
If we ignored the errors, then we would proceed with an incomplete upgrade and
have a very difficult time every recovering the proper state.

Now, compile the new binary and run the upgraded code to complete the upgrade:

```bash
# Create a new binary of gaia with added upgrade handler using
make install

# Restart the chain using new binary, you should see  the chain resume from the upgrade height
# Like `I[2019-11-05|12:48:15.184] applying upgrade "test1" at height: 200      module=main`
gaiad start

# Verify no more pending plan
gaiacli query upgrade plan

# You can query the block header of the completed upgrade
gaiacli query upgrade applied cosmoshub-x
```
