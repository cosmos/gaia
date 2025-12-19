# How to upgrade your node

There are two major kinds of upgrades:
* Governance gated
* Non-governance gated

In both cases, you can upgrade your node by replacing the binary manually or using Cosmovisor.


## Governance Gated Upgrade

A governance gated upgrade occurs when an upgrade plan goes on chain through a software upgrade proposal.
* Your node **will** stop at the height listed in the upgrade plan.

### Manual Binary Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Wait for the node to stop at the upgrade height.
  * The log will display something like this:
    ```
    ERR UPGRADE "v25.3.0" NEEDED at height: <UPGRADE_HEIGHT>: upgrade to v25.3.0 and applying upgrade "v25.3.0" at height:<UPGRADE_HEIGHT>
    ```
* If the node service remains active, you can stop it now.
3. Replace the binary listed in the unit file with the new release.
4. Start the node service again.

### Cosmovisor Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Create a folder for the new binary in the relevant Cosmovisor directory.
  * If the upgrade name is `v25.3.0`, you would place the binary under `<node home>/cosmovisor/upgrades/v25.3.0/bin/gaiad`:
    ```
    .
    ├── current -> genesis or upgrades/<name>
    ├── genesis
    │   └── bin
    │       └── gaiad  # old: v25.2.0
    └── upgrades
        └── v25.3.0
            └── bin
                └── gaiad  # new: v25.3.0
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


## Non-governance Gated Upgrade

A non-governance gated upgrade occurs when nodes upgrade their binary at an agreed-upon height.

* You **must** stop your node at the relevant height and replace the binary before starting it again.
* ⚠️ Monitor your logs when the halt height is about to be reached. If your node does not stop at the specified height using the instructions below, stop the service manually and replace the binary before starting the service again.

### Manual Binary Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Stop the node service.
3. Update the app.toml file in your node's config folder to set the `halt-height` variable.
  * A sample app.toml line will look like this for an upgrade height of `24666000`:
    ```
    halt-height = 24666000
    ```
4. Restart the service.
5. Wait for the halt height.
  * The log will display something like this when height is reached:
    ```
    ERR CONSENSUS FAILURE!!! err="failed to apply block; error halt per configuration height 24666000 time 0" 
    ```
7. Stop the node service.
8. Replace the binary listed in the unit file with the new release.
  * For Cosmovisor, this will be `<node home>/cosmovisor/current/bin/gaiad`.
9.  Update app.toml to set the `halt-height` back to 0.
10. Start the node service.

### Cosmovisor `add-upgrade`

#### ⚠️ Be aware that there are reported issues with using Cosmovisor for non-governance upgrades: the instructions below may result in unexpected behaviour, such as an immediate upgrade or failure to stop at the specified height.

1. Build or download the binary for the release you are upgrading to.
2. Run the `add-upgrade` command. This command will look as follows using the example values listed below.
  * New binary: `gaiad-v23.0.0-linux-amd64`
  * Upgrade name: `v23.0.0`
  * Halt height: `24666000`
   ```
   cosmovisor add-upgrade v23.0.0 gaiad-v23.0.0-linux-amd64 --upgrade-height 24666000 --force
   ```
3. Wait for the node to reach the upgrade height. Cosmovisor will restart the node using the new binary.