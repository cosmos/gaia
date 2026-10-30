# How to upgrade your node

There are two major kinds of upgrades:
* Governance gated
* Non-governance gated

In both cases, you can upgrade your node by replacing the binary manually or using Cosmovisor.

The commands, paths, configuration snippets, and logs below are examples. Before using them, check the release notes and the approved on-chain upgrade plan (for governance gated upgrades), or the network upgrade announcement (for non-governance gated upgrades).

Replace the following placeholders with the values for your upgrade:

| Placeholder | Value |
| --- | --- |
| `<UPGRADE_NAME>` | The exact upgrade plan name. For governance gated upgrades, use the on-chain plan name; it may differ from the binary release version. |
| `<UPGRADE_HEIGHT>` | The block height specified in the on-chain plan or agreed upon for the network upgrade. |
| `<NEW_BINARY_PATH>` | The path to the new `gaiad` binary built or downloaded for your platform. |
| `<NODE_HOME>` | Your node's home directory. |
| `<PREVIOUS_UPGRADE_NAME>` | The upgrade directory currently targeted by Cosmovisor's `current` symlink, if it does not point to `genesis`. |

Do not copy placeholders literally into commands or configuration files. Adapt the sample service user and paths to your node's setup as well.


## Governance Gated Upgrade

A governance gated upgrade occurs when an upgrade plan goes on chain through a software upgrade proposal.
* Your node **will** stop at the height listed in the upgrade plan.

### Manual Binary Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Wait for the node to stop at the upgrade height.
  * The log will display something like this:
    ```
    ERR UPGRADE "<UPGRADE_NAME>" NEEDED at height: <UPGRADE_HEIGHT>: upgrade to <UPGRADE_NAME> and applying upgrade "<UPGRADE_NAME>" at height:<UPGRADE_HEIGHT>
    ```
* If the node service remains active, you can stop it now.
3. Replace the binary listed in the unit file with the new release.
4. Start the node service again.

### Cosmovisor Upgrade

1. Build or download the binary for the release you are upgrading to.
2. Create a folder for the new binary in the relevant Cosmovisor directory.
  * Place the binary under `<NODE_HOME>/cosmovisor/upgrades/<UPGRADE_NAME>/bin/gaiad`, using the exact upgrade plan name for the directory:
    ```
    .
    ├── current -> genesis or upgrades/<PREVIOUS_UPGRADE_NAME>
    ├── genesis
    │   └── bin
    │       └── gaiad  # genesis binary
    └── upgrades
        └── <UPGRADE_NAME>
            └── bin
                └── gaiad  # new binary
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
  * Set the agreed-upon upgrade height as an integer:
    ```
    halt-height = <UPGRADE_HEIGHT>
    ```
4. Restart the service.
5. Wait for the halt height.
  * The log will display something like this when height is reached:
    ```
    ERR CONSENSUS FAILURE!!! err="failed to apply block; error halt per configuration height <UPGRADE_HEIGHT> time 0"
    ```
6. Stop the node service.
7. Replace the binary listed in the unit file with the new release.
  * For Cosmovisor, this will be `<NODE_HOME>/cosmovisor/current/bin/gaiad`.
8. Update app.toml to set the `halt-height` back to 0.
9. Start the node service.

### Cosmovisor `add-upgrade`

#### ⚠️ Be aware that there are reported issues with using Cosmovisor for non-governance upgrades: the instructions below may result in unexpected behaviour, such as an immediate upgrade or failure to stop at the specified height.

1. Build or download the binary for the release you are upgrading to.
2. Run the `add-upgrade` command using the upgrade name and height from the network upgrade announcement and the path to your new binary:
   ```
   cosmovisor add-upgrade <UPGRADE_NAME> <NEW_BINARY_PATH> --upgrade-height <UPGRADE_HEIGHT> --force
   ```
3. Wait for the node to reach the upgrade height. Cosmovisor will restart the node using the new binary.
