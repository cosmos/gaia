# Live Upgrade Tutorial

This document demonstrates how a live upgrade can be performed on-chain through a
governance process.

1. Start the network and trigger upgrade

   ```bash
   # start a gaia application full-node
   $ gaiad start

   # set up the cli config
   $ gaiad config trust-node true
   $ gaiad config chain-id testing

   # create an upgrade governance proposal
   $ gaiad tx gov submit-proposal software-upgrade <plan-name> \
   --title <proposal-title> --description <proposal-description> \
   --from <name-or-key> --upgrade-height <desired-upgrade-height> --deposit 10000000stake

   # once the proposal passes you can query the pending plan
   $ gaiad query upgrade plan
   ```

2. Performing an upgrade

   Assuming the proposal passes the chain will stop at given upgrade height.

   You can stop and start the original binary all you want, but **it will refuse to
   run after the upgrade height**.

   We need a new binary with the upgrade handler installed. The logs should look
   something like:

   ```bash
   E[2019-11-05|12:44:18.913] UPGRADE "<plan-name>" NEEDED at height: <desired-upgrade-height>:       module=main
   E[2019-11-05|12:44:18.914] CONSENSUS FAILURE!!!
   ...
   ```

   Note that the process will hang indefinitely (doesn't exit to avoid restart loops). So, you must
   manually kill the process and replace it with a new binary. Do so now with `Ctrl+C` or `killall gaiad`.

   In `gaia/app/app.go`, after `upgrade.Keeper` is initialized and set in the app, set the the
   corresponding upgrade `Handler` with the correct `<plan-name>`:

   ```go
       app.upgradeKeeper.SetUpgradeHandler("<plan-name>", func(ctx sdk.Context, plan upgrade.Plan) {
           // custom logic after the network upgrade has been executed
       })
   ```

   Note that we panic on any error - this would cause the upgrade to fail if the
   migration could not be run, and no node would advance - allowing a manual recovery.
   If we ignored the errors, then we would proceed with an incomplete upgrade and
   have a very difficult time every recovering the proper state.

   Now, compile the new binary and run the upgraded code to complete the upgrade:

   ```bash
   # create a new binary of gaia with the added upgrade handler
   $ make install

   # Restart the chain using the new binary. You should see the chain resume from
   # the upgrade height:
   # `I[2019-11-05|12:48:15.184] applying upgrade <plan-name> at height: <desired-upgrade-height>      module=main`
   $ gaiad start

   # verify there is no pending plan
   $ gaiad query upgrade plan

   # verify you can query the block header of the completed upgrade
   $ gaiad query upgrade applied <plan-name>
   ```
