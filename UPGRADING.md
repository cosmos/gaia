# How to upgrade your node

This upgrade is the governance gated upgrade

You can upgrade your node by replacing the binary manually or using Cosmovisor.

> [!IMPORTANT]
> Upgrade name: `v29.1.0`. Upgrade height: `<UPGRADE_HEIGHT>`. <br>
> `v29.1.0` is a state-breaking release. Keep running `v29.0.1` until your node stops at the upgrade height, then switch to `v29.1.0`. <br>
> Do not start `v29.1.0` before the upgrade height.

## CosmWasm notes

`v29.1.0` uses wasmvm `v2.3.6` (previously `v2.3.5`).

* The release binaries link `libwasmvm` statically. If you build from source against a shared `libwasmvm` library (for example `/usr/lib/libwasmvm.x86_64.so`), replace it with the `v2.3.6` build before the upgrade. `gaiad start` refuses to run with a mismatched library.
  * Check the library version with `gaiad query wasm libwasmvm-version`. It must print `2.3.6`.
* The compiled module cache moves from `3858fb5318-wasmer24` to `20bed02e1f-wasmer24`.
  * On the first start after the upgrade, the 08-wasm light client code is recompiled. CosmWasm contracts are recompiled the first time each one is executed, so the first blocks after the upgrade may process wasm transactions slightly slower. Gas is not affected.
  * The old cache directories are not removed automatically. After confirming that the node runs normally, you can remove them:
    ```
    <node home>/wasm/cache/modules/3858fb5318-wasmer24
    <node home>/data/08-light-client/cache/modules/3858fb5318-wasmer24
    ```

## Governance Gated Upgrade

A governance gated upgrade occurs when an upgrade plan goes on chain through a software upgrade proposal.
* Your node **will** stop at the height listed in the upgrade plan.

### Manual Binary Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Wait for the node to stop at the upgrade height.
  * The log will display something like this:
    ```
    ERR UPGRADE "v29.1.0" NEEDED at height: <UPGRADE_HEIGHT>: upgrade to v29.1.0 and applying upgrade "v29.1.0" at height:<UPGRADE_HEIGHT>
    ```
* If the node service remains active, you can stop it now.
3. Replace the binary listed in the unit file with the new release [v29.1.0](https://github.com/cosmos/gaia/releases/tag/v29.1.0)
4. Start the node service again.

### Cosmovisor Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Create a folder for the new binary in the relevant Cosmovisor directory.
  * If the upgrade name is `v29.1.0`, you would place the binary under `<node home>/cosmovisor/upgrades/v29.1.0/bin/gaiad`:
    ```
    .
    ├── current -> upgrades/v29.0.0
    ├── genesis
    │   └── bin
    │       └── gaiad
    └── upgrades
        ├── v29.0.0
        │   └── bin
        │       └── gaiad  # old: v29.0.1
        └── v29.1.0
            └── bin
                └── gaiad  # new: v29.1.0
    ```
3. Verify that Cosmovisor will use the binary you have prepared.
  * The Cosmovisor service should have the auto-download feature disabled. A sample Cosmovisor unit file will look like this:
    ```
    [Unit]
    Description=Cosmovisor service
    After=network-online.target

    [Service]
    User=gaia
    ExecStart=/gaia/go/bin/cosmovisor run start --home /gaia/.gaia
    Restart=no
    LimitNOFILE=50000
    Environment='DAEMON_NAME=gaiad'
    Environment='DAEMON_HOME=/gaia/.gaia'
    Environment='DAEMON_ALLOW_DOWNLOAD_BINARIES=false'
    Environment='DAEMON_RESTART_AFTER_UPGRADE=true'
    Environment='DAEMON_LOG_BUFFER_SIZE=512'
    Environment='UNSAFE_SKIP_BACKUP=true'

    [Install]
    WantedBy=multi-user.target
    ```
* If you must change the contents of the unit file, run `systemctl daemon-reload` before restarting the service.
4. Wait for the node to reach the upgrade height. Cosmovisor will restart the node using the new binary.
