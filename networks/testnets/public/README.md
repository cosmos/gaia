# Cosmos Hub Public Testnet

The public testnet will be used to test Theta, Rho, and Lambda upgrades. It mirrors the state of mainnet, aside from a few modifications to the exported genesis file. These adjustments help provide liveness and streamlined governance-permissioned software upgrades.

Visit the [Scheduled Upgrades](UPGRADES.md) page for details on current and upcoming versions. 

## Testnet Details

- **Chain-ID**: `theta-testnet-001`
- **Launch date**: 2022-03-10
- **Current Gaia Version:** `v8.0.0-rc3` (upgraded to v8 at height `14175595`)
- **Launch Gaia Version:** `release/v6.0.0`
- **Genesis File:**  Zipped and included [in this repository](genesis.json.gz), unzip and verify with `shasum -a 256 genesis.json`
- **Genesis sha256sum**: `522d7e5227ca35ec9bbee5ab3fe9d43b61752c6bdbb9e7996b38307d7362bb7e`

### Endpoints

Endpoints are exposed as subdomains for the sentry and snapshot nodes (described below) as follows:

* `https://rpc.<node-name>.theta-testnet.polypore.xyz:443`
* `https://rest.<node-name>.theta-testnet.polypore.xyz:443`
* `https://grpc.<node-name>.theta-testnet.polypore.xyz:443`
* `p2p.<node-name>.theta-testnet.polypore.xyz:26656`

Sentries:

1. `sentry-01.theta-testnet.polypore.xyz`
2. `sentry-02.theta-testnet.polypore.xyz`

Seed nodes:

1. `seed-01.theta-testnet.polypore.xyz`
2. `seed-02.theta-testnet.polypore.xyz`

The following state sync nodes serve snapshots every 1000 blocks:

1. `state-sync-01.theta-testnet.polypore.xyz`
2. `state-sync-02.theta-testnet.polypore.xyz`

### Seeds

You can add these in your seeds list.

```
639d50339d7045436c756a042906b9a69970913f@seed-01.theta-testnet.polypore.xyz:26656
3e506472683ceb7ed75c1578d092c79785c27857@seed-02.theta-testnet.polypore.xyz:26656
```

### Block Explorers

  - https://explorer.theta-testnet.polypore.xyz
  - https://cosmoshub-testnet.mintscan.io/cosmoshub-testnet
  - https://testnet.cosmos.bigdipper.live/

### Faucet

Visit the [🚰・testnet-faucet](https://discord.com/channels/669268347736686612/953697793476821092) channel in the Cosmos Developers Discord.


## Add to Keplr

Use this [jsfiddle](https://jsfiddle.net/kht96uvo/1/).

## How to Join

Both of the methods shown below will install Gaia and set up a Cosmovisor service with the auto-download feature enabled on your machine.

You can choose to (not) use state sync both ways. Your node will sync much faster if you use state sync, but it will not keep all the state locally.

### Ansible Playbook

Use the example inventory file from the [cosmos-ansible](https://github.com/hyphacoop/cosmos-ansible) repo to set up a node using state sync:

```
git clone https://github.com/hyphacoop/cosmos-ansible.git
cd cosmos-ansible
ansible-playbook node.yml -i examples/inventory-public-testnet.yml -e 'target=SERVER_IP_OR_DOMAIN'
```

The video below provides an overview of how the playbook sets up the node.

[![Join the Cosmos Theta Testnet](https://img.youtube.com/vi/SYt0EC5pcY0/0.jpg)](https://www.youtube.com/watch?v=SYt0EC5pcY0)

If you want to sync from genesis, set the following variables in the inventory file:
* `gaiad_version: v6.0.4`
* `statesync_enabled: false`

For additional information, visit the [examples page](https://github.com/hyphacoop/cosmos-ansible/tree/main/examples#join-the-theta-testnet).

### Bash Script

Run either one of the scripts provided in this repo to join the provider chain:
* `join-public-testnet.sh` will create a `gaiad` service.
* `join-public-testnet-cv.sh` will create a `cosmovisor` service.
* Both scripts must be run either as root or from a sudoer account.
* Both scripts will attempt to download the amd64 binary from the Gaia [releases](https://github.com/cosmos/gaia/releases) page. You can modify the `CHAIN_BINARY_URL` to match your target architecture if needed.

#### State Sync Option

* By default, the scripts will attempt to use state sync to catch up quickly to the current height. To turn off state sync, set `STATE_SYNC` to `false`.
* If you want to sync from genesis:
  * Turn off state sync.
  * Start with gaiad v6.0.4 and upgrade at the block heights described in the [Scheduled Upgrades](UPGRADES.md) page.
    * To run gaiad v6.0.4, you can download the appropriate binary or build from source.
    * To build from source, uncomment the below the binary download and use `git checkout v6.0.4` prior to `make install`.

### Cosmovisor

#### Ugrading with auto-download vs. manually preparing your binary

If you want to use Cosmovisor's **auto-download** feature, please set the environment variable `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`

If you are **manually preparing your binary**, please set the environement variable `DAEMON_ALLOW_DOWNLOAD_BINARIES=false` and download a copy of the v8.0.0-rc3 binary to the v8-Rho upgrade directory in your cosmovisor directory (`upgrades/v8-rho/bin/gaiad`). **If you are using Cosmovisor `v1.0.0`, the version name is not lowercased (use `upgrades/v8-Rho/bin/gaiad` instead)**.

```
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad
└── upgrades
    └── v8-rho
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

#### Upgrade Backup

Cosmovisor will attempt to make a backup of the home folder before upgrading, which will consume time and considerable disk space. If you want to skip this step, set the environment variable `UNSAFE_SKIP_BACKUP` to `true`.

## Public testnet modifications

The following modifications were made using the [cosmos-genesis-tinker script](https://github.com/hyphacoop/cosmos-genesis-tinkerer/blob/main/example_stateful_genesis.py):

1. Autoloading ./exported_genesis.json.preprocessed.json
2. Loading genesis from file ./exported_genesis.json.preprocessed.json
3. Swapping chain id to theta-testnet-001
4. Increasing balance of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 175000000000000 uatom
5. Increasing supply of uatom by 175000000000000
6. Increasing balance of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 1000 theta
7. Increasing supply of theta by 1000
8. Creating new coin theta valued at 1000
9. Increasing balance of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 1000 rho
10. Increasing supply of rho by 1000
11. Creating new coin rho valued at 1000
12. Increasing balance of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 1000 lambda
13. Increasing supply of lambda by 1000
14. Creating new coin lambda valued at 1000
15. Increasing balance of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 1000 epsilon
16. Increasing supply of epsilon by 1000
17. Creating new coin epsilon valued at 1000
18. Increasing balance of cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh by 550000000000000 uatom
19. Increasing supply of uatom by 550000000000000
20. Increasing delegator stake of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw by 550000000000000
21. Increasing validator stake of cosmosvaloper10v6wvdenee8r9l6wlsphcgur2ltl8ztkfrvj9a by 550000000000000
22. Increasing validator power of A8A7A64D1F8FFAF2A5332177F777A5816036D65A by 550000000
23. Increasing delegations of cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw with cosmosvaloper10v6wvdenee8r9l6wlsphcgur2ltl8ztkfrvj9a by 550000000000000.0
24. Swapping min governance deposit amount to 1uatom
25. Swapping tally parameter quorum to 0.000000000000000001
26. Swapping tally parameter threshold to 0.000000000000000001
27. Swapping governance voting period to 60s
28. Swapping staking unbonding_time to 1s

SHA256SUM: `522d7e5227ca35ec9bbee5ab3fe9d43b61752c6bdbb9e7996b38307d7362bb7e`
