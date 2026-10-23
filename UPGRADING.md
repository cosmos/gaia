# How to upgrade your node

This upgrade is the governance gated upgrade (proposal [#1059](https://www.mintscan.io/cosmos/proposals/1059))

You can upgrade your node by replacing the binary manually or using Cosmovisor.

> [!IMPORTANT]
> Use `v29.0.1` instead of `v29.0.0` for the `v29.0.0` upgrade at height `33,618,500`. <br>
> Nodes already running `v29.0.0` can switch to `v29.0.1` at any time by replacing the binary and restarting. <br>
> `v29.0.1` - is a non-state-breaking patch release that rebuilds Gaia with `Go 1.26.10`

## Governance Gated Upgrade

A governance gated upgrade occurs when an upgrade plan goes on chain through a software upgrade proposal.
* Your node **will** stop at the height listed in the upgrade plan.

### Manual Binary Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Wait for the node to stop at the upgrade height.
  * The log will display something like this:
    ```
    ERR UPGRADE "v29.0.0" NEEDED at height: <UPGRADE_HEIGHT>: upgrade to v29.0.0 and applying upgrade "v29.0.0" at height:<UPGRADE_HEIGHT>
    ```
* If the node service remains active, you can stop it now.
3. Replace the binary listed in the unit file with the new release [v29.0.1](https://github.com/cosmos/gaia/releases/tag/v29.0.1)
4. Start the node service again.

### Cosmovisor Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Create a folder for the new binary in the relevant Cosmovisor directory.
  * If the upgrade name is `v29.0.0`, you would place the binary under `<node home>/cosmovisor/upgrades/v29.0.0/bin/gaiad`:
    ```
    .
    ├── current -> genesis or upgrades/<name>
    ├── genesis
    │   └── bin
    │       └── gaiad  # old: v28.3.1
    └── upgrades
        └── v29.0.0
            └── bin
                └── gaiad  # new: v29.0.1
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