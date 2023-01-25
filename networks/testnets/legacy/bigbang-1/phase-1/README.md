# Phase-1 : Instructions

## Software Requirements:
- Go version v1.14.+
- Cosmos SDK version: [v0.39.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.1)
- Akash version : [v0.8.1](https://github.com/ovrclk/akash/releases/tag/v0.8.1)

### Install Akash
```
$ mkdir -p $GOPATH/src/github.com/ovrclk
$ cd $GOPATH/src/github.com/ovrclk
$ git clone https://github.com/ovrclk/akash && cd akash
$ git checkout v0.8.1
$ make install
```

To verify if the installation was successful, execute the following command:
```
$ akashd version --long
```
It will display the version of akashd currently installed:
```
name: akash
server_name: akashd
client_name: akashctl
version: 0.8.1
commit: 1f7e40ae25da683f9728eb36d28ace3b6f9b7604
build_tags: netgo,ledger
go: go version go1.15.2 linux/amd64
```

## Activity-1: Start your validator node
If you are looking to join the testnet post genesis time (i.e, _20-Oct-2020 16:00UTC_), skip to [Create Testnet Validator](#create-testnet-validator)

Below are the instructions to generate & submit your `GenTx`
### Generate GenTx
1. Initialize the akash directories and create the local genesis file with the correct
   chain-id

   ```shell
   $ akashd init <moniker-name> --chain-id=bigbang-1
   ```

2. Create a local key pair in the Keyring

   ```shell
   $ akashctl keys add <key-name>
   ```

3. Add your account to your local genesis file with a given amount and the key you
   just created. Use only `1000000000star`, other amounts will be ignored.

   ```shell
   $ akashd add-genesis-account $(akashctl keys show <key-name> -a) 1000000000star
   ```

4. Create the gentx

   ```shell
   $ akashd gentx --amount 900000000star --name=<key-name>
   ```

   If all goes well, you will see a message similar to the following:
    ```shell
    Genesis transaction written to "/home/user/.akashd/config/gentx/gentx-******.json"
    ```

### Submit Gentx
Submit your gentx in a PR [here](https://github.com/cosmos/testnets)

- Fork the testnets repo into your github account 

- Clone your repo using

    ```sh
    $ git clone https://github.com/<your-github-username>/testnets
    ```

- Copy the generated gentx json file to `<repo_path>/bigbang-1/gentx/`

    ```sh
    $ cd $GOPATH/src/github.com/cosmos/testnets
    $ cp ~/.akashd/config/gentx/gentx*.json ./bigbang-1/gentx/
    ```

- Commit and push to your repo
- Create a PR onto https://github.com/cosmos/testnets


### Start your validator node
Once the genesis is released (i.e., _19-Oct-2020 16:00UTC_), follow the instructions below to start your validator node.

#### Genesis & Seeds
Fetch `genesis.json` into `akashd`'s `config` directory.
```
$ curl https://raw.githubusercontent.com/cosmos/testnets/master/bigbang-1/genesis.json > $HOME/.akashd/config/genesis.json
```

Add seed nodes in `config.toml`.

  ```seeds = "<To be published here>"```
  ```persistent_peers = "<To be published here>"```
```
$ nano $HOME/.akashd/config/config.toml
```
Find the following section and add the seed nodes.
```
# Comma separated list of seed nodes to connect to
seeds = ""
```
```
# Comma separated list of persistent peers to connect to
persistent_peers = "7eb6b7cf07b47991c786994d48b7bb143300bc1b@157.230.185.206:26656" // Witval peer
```

#### Start Your Node

Create a systemd service

```shell
$ sudo nano /lib/systemd/system/akashd.service
```

Copy-Paste in the following and update `<your_username>` and `<go_workspace>` as required:

```
[Unit]
Description=akashd
After=network-online.target

[Service]
User=<your_username>
ExecStart=/home/<your_username>/<go_workspace>/bin/akashd start
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

**This tutorial assumes `$HOME/go_workspace` to be your Go workspace. Your actual workspace directory may vary.**

```
$ sudo systemctl enable akashd
$ sudo systemctl start akashd
```
Check node status
```
$ systemctl status akashd 
```
Check logs
```
$ sudo journalctl -u akashd -f
```

## Create Testnet Validator
This section applies to those who are looking to join the testnet post genesis.

1. Init Chain and start your node
   ```shell
   $ akashd init <moniker-name> --chain-id=bigbang-1
   ```

   After that, please follow all the instructions from [Start your validator node ](#start-your-validator-node) section.


2. Create a local key pair in the Keyring

   ```shell
   $ akashctl keys add <key-name>
   $ akashctl keys show <key-name> -a
   ```

3. Request tokens from faucet: https://faucet.bigbang.vitwit.com

4. Create validator

   ```shell
   $ akashctl tx staking create-validator \
   --amount 900000000star \
   --commission-max-change-rate "0.1" \
   --commission-max-rate "0.20" \
   --commission-rate "0.1" \
   --min-self-delegation "1" \
   --details "Some details about yourvalidator" \
   --pubkey=$(akashd tendermint show-validator) \
   --moniker <your_moniker> \
   --chain-id bigbang-1 \
   --from <key-name> 
   ```

## Useful Links
### Explorers: [Aneka](https://bigbang.aneka.io) [BigDipper](https://bigbang.bigdipper.live)
### Faucet: https://faucet.bigbang.vitwit.com
