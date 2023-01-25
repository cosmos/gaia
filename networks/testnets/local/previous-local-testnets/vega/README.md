# Cosmos-Hub Vega  Upgrade Local Testnet Instruction

This document describes the process on how to test the Cosmos Hub Vega upgrade locally.
This upgrade will integrate the new release of Cosmos-SDK v0.44.2 and IBC 1.2.1 into Gaia.

## Version
- Currently: running cosmoshub-4, Gaia version: v5.0.5
- After the upgrade: cosmoshub-4, Gaia version: v6.0.0

## Chain upgrade by cosmovisor

### Test Plan
This document uses the data exported from a live instance of cosmoshub-4 to mock the upgrade. We will run two nodes locally, with an exported genesis file to upgrade both nodes to Gaia v6.0.0-Vega using Cosmovisor. Specifically, we will do the following:
- We will modify the genesis file to take control of two nodes, by swapping out the accounts of Certus One (node 1) and Binance (node 2) and replacing them with ours.
- We will give node 2 more than 67% voting power, so that we can produce blocks ourselves.
- We will take control of a user account to let this user propose, deposit and vote for the upgrade proposal. This user is also a delegator so that they can use their own delegation to vote.
- We will change the voting parameters so that the upgrade proposal can be processed fast, and that it will pass without any issues.

### Build the binary of old version
```shell
cd gaia
git checkout release/v5.0.5
make install
```
```shell
gaiad version
# v5.0.5
```
### Change the genesis file

We have prepared a genesis file in this repo which was obtained by `gaiad export` on cosmoshub-4 network at height 7368387. Uncompress this genesis file and use it as the genesis data to mock the cosmoshub upgrade.

```shell
cd Vega-test
gunzip exported_unmodified_genesis.json.gz
# verify the hash
cat exported_unmodified_genesis.json | shasum -a 256
> 
86f29f23f9df51f5c58cb2c2f95e263f96f123801fc9844765f98eca49fe188f
```

Change one validator in the genesis file to be an account that you have the private key for, and make this validator own over 67% power, so that you can start to produce blocks on a local chain.

#### Change the addresses and keys
```shell
# change chain id
sed -i '' 's%"chain_id": "cosmoshub-4",%"chain_id": "test",%g' genesis.json

# substitute "Certus One", this is our  node1, you can find the key info. in priv_validator_key_val1.json in this repo.
sed -i '' 's%cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM=%qwiUMxz3llsy45fPvM0a8+XQeAJLvrX3QAEJmRMEEoU=%g' genesis.json
sed -i '' 's%B00A6323737F321EB0B8D59C6FD497A14B60938A%D5AB5E458FD9F9964EF50A80451B6F3922E6A4AA%g' genesis.json
sed -i '' 's%cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6%cosmosvalcons16k44u3v0m8uevnh4p2qy2xm08y3wdf92xsc3ve%g' genesis.json

# substitute "Binance Staking", this is our node2, also the validator who will own over 67% power. you can find the key info. in priv_validator_key_val1.json in this repo. 
# tendermint pub_key
sed -i '' 's%W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o=%oi55Dw+JjLQc4u1WlAS3FsGwh5fd5/N5cP3VOLnZ/H0=%g' genesis.json
# priv_val address
sed -i '' 's%83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0%7CB07B94FD743E2A8520C2B50DA4B03740643BF5%g' genesis.json
#  Validator consensus address, try command ` gaiad keys parse 83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0` to see if you can get the same addr.
sed -i '' 's%cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3%cosmosvalcons10jc8h98awslz4pfqc26smf9sxaqxgwl4vxpcrp%g' genesis.json

# substitute a user account,this user account is a delegator in the genesis file. This user account will be owned by node2(validator2) in the later setup.
sed -i '' 's%cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx%cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9%g' genesis.json
sed -i '' 's%A6apc7iThbRkwboKqPy6eXxxQvTH+0lNkXZvugDM9V4g%ApDOUyfcamDmnbEO7O4YKnKQQqQ93+gquLfGf7h5clX7%g' genesis.json
```
#### Fix delegation amount over 67%
```shell
# change one delegator's delegation. This delegator can be any delegator who is delegating to our validator2(Binance Staking) in the delegation list. Increase this stake by 6,000,000,000,000,000.
sed -i '' 's%"25390741.000000000000000000"%"6000000025390741.000000000000000000"%g' genesis.json

# fix power of the validator
# Binance Staking validator's "delegator_shares" and "tokens"
# Increase the "delegator_shares" by 6,000,000,000,000,000 correspondingly.
sed -i '' 's%13944328343563%6013944328343563%g' genesis.json
# Increase the validator power by 6,000,000,000
sed -i '' 's%"power": "13944328"%"power": "6013944328"%g' genesis.json

# fix last_total_power
# Increase total amounts of bonded tokens recorded during the previous end block by 6,000,000,000
sed -i '' 's%"194616038"%"6194616038"%g' genesis.json

# fix total supply of uatom
sed -i '' 's%277834757180509%6277834757180509%g' genesis.json

# fix balance of bonded_tokens_pool module account
#
# module account for recording Binance staking(validator 2)'s received delegations:
# cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh
# Increase the delegation account by 6,000,000,000,000,000
sed -i '' 's%194616098248861%6194616098248861%g' genesis.json
```

#### Modify some gov parameters to improve testing efficiency

Change the minimum deposit amount, quorum, threshold, and voting period. These changes will allow the user account's vote to pass when voting for the upgrade proposal, because they are also a delegator.
```shell
# min deposition amount
sed -i '' 's%"amount": "64000000",%"amount": "1",%g' genesis.json
#   min voting power that a proposal requires in order to be a valid proposal
sed -i '' 's%"quorum": "0.400000000000000000",%"quorum": "0.000000000000000001",%g' genesis.json
# the minimum proportion of "yes" votes requires for the proposal to pass
sed -i '' 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' genesis.json
# voting period 
sed -i '' 's%"voting_period": "1209600s"%"voting_period": "60s"%g' genesis.json
```
<em>Please note: if you prefer not to replace the data in the genesis yourself as shown above, we have prepared a [modified genesis file](https://github.com/cosmos/vega-test/tree/master/local-testnet/modified_genesis_local_testnet) which can be directly used in local testnet.<em>

```
# verify the hash of the modified genesis:
cat genesis.json | shasum -a 256
> 
7374ea0ccceead23ff1a02cf041f4bf53dc7fe21692d3511b8812668720b4b8e
```

### Init the chain
#### Setup the environmental variables
```shell
export EXPORTED_GENESIS=genesis.json 
export BINARY=gaiad 
export CHAIN_ID=test 
export CHAIN_DIR=data 

export VAL_1_CHAIN_DIR=$CHAIN_DIR/$CHAIN_ID/val1 
export VAL_2_CHAIN_DIR=$CHAIN_DIR/$CHAIN_ID/val2 
export VAL_1_KEY_NAME="val1" 
export VAL_2_KEY_NAME="val2" 
export VAL_1_MONIKER="val1" 
export VAL_2_MONIKER="val2" 

export USER_CHAIN_DIR=$CHAIN_DIR/$CHAIN_ID/val2
export USER_MNEMONIC="junk appear guide guess bar reject vendor illegal script sting shock afraid detect ginger other theory relief dress develop core pull across hen float"
export USER_KEY_NAME="user"
```
#### Init the chain and setup the user account
```shell
$BINARY config chain-id test --home $VAL_1_CHAIN_DIR 
$BINARY config keyring-backend test --home $VAL_1_CHAIN_DIR 
$BINARY config broadcast-mode block --home $VAL_1_CHAIN_DIR 

$BINARY config chain-id test --home $VAL_2_CHAIN_DIR 
$BINARY config keyring-backend test --home $VAL_2_CHAIN_DIR 
$BINARY config broadcast-mode block --home $VAL_2_CHAIN_DIR 

# Validator 1
$BINARY init test --home $VAL_1_CHAIN_DIR --chain-id=$CHAIN_ID
# Validator 2
$BINARY --home $VAL_2_CHAIN_DIR init test --chain-id=$CHAIN_ID
#user
echo $USER_MNEMONIC | $BINARY --home $USER_CHAIN_DIR keys add $USER_KEY_NAME --recover --keyring-backend=test
```
#### Replace the genesis file and priv_validator_key.json
```shell
cp genesis.json $VAL_1_CHAIN_DIR/config/genesis.json
cp genesis.json $VAL_2_CHAIN_DIR/config/genesis.json
cp priv_validator_key_val1.json $VAL_1_CHAIN_DIR/config/priv_validator_key.json &&
cp priv_validator_key_val2.json $VAL_2_CHAIN_DIR/config/priv_validator_key.json
```
#### Setup configurations for synchronization
```shell
export VAL_1_P2P_PORT=26656 
export VAL_1_NODE_ID=$($BINARY tendermint --home $VAL_1_CHAIN_DIR show-node-id) 
export VAL_2_P2P_PORT=36656 
export VAL_2_RPC_PORT=36657 
export VAL_2_API_PORT=1327 
export VAL_2_GRPC_PORT=9080 
export VAL_2_GRPC_WEB_SERVER_PORT=9081 
export VAL_2_ROSETTA_API_PORT=8081 
export VAL_2_PPROF_PORT=6061 
export VAL_2_NODE_ID=$($BINARY tendermint --home $VAL_2_CHAIN_DIR show-node-id)
```
The following changes resolve the conflicts so that validator 1 and 2 will not use the same default ports, and let those two nodes be each other's peer node.
```shell
sed -i '' 's/enable = true/enable = false/g' $VAL_1_CHAIN_DIR/config/app.toml # disable all for val1 to prevent from colluding ports
sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$VAL_2_RPC_PORT"'"#g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$VAL_2_P2P_PORT"'"#g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$VAL_2_API_PORT"'"#g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's#"0.0.0.0:9090"#"0.0.0.0:'"$VAL_2_GRPC_PORT"'"#g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's#"0.0.0.0:9091"#"0.0.0.0:'"$VAL_2_GRPC_WEB_SERVER_PORT"'"#g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's#":8080"#":'"$VAL_2_ROSETTA_API_PORT"'"#g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's/enable = false/enable = true/g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's/swagger = false/swagger = true/g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/g' $VAL_1_CHAIN_DIR/config/app.toml
sed -i '' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/g' $VAL_2_CHAIN_DIR/config/app.toml
sed -i '' 's/persistent_peers = ""/persistent_peers = "'$VAL_2_NODE_ID'@'localhost':'$VAL_2_P2P_PORT'"/g' $VAL_1_CHAIN_DIR/config/config.toml
sed -i '' 's/persistent_peers = ""/persistent_peers = "'$VAL_1_NODE_ID'@'localhost':'$VAL_1_P2P_PORT'"/g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's/unconditional_peer_ids = ""/unconditional_peer_ids = "'$VAL_1_NODE_ID'"/g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's/unconditional_peer_ids = ""/unconditional_peer_ids = "'$VAL_2_NODE_ID'"/g' $VAL_1_CHAIN_DIR/config/config.toml
sed -i '' 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$VAL_2_PPROF_PORT'"/g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's/addr_book_strict = true/addr_book_strict = false/g' $VAL_2_CHAIN_DIR/config/config.toml
sed -i '' 's/addr_book_strict = true/addr_book_strict = false/g' $VAL_1_CHAIN_DIR/config/config.toml
```
Disable the Rosetta API server:
```shell
# Enable defines if the Rosetta API server should be enabled.
# enable = false
sed -i '' '/Enable defines if the Rosetta API server/,/Address defines the Rosetta API server/s/enable = true/enable = false/' $VAL_1_CHAIN_DIR/config/app.toml
sed -i '' '/Enable defines if the Rosetta API server/,/Address defines the Rosetta API server/s/enable = true/enable = false/' $VAL_2_CHAIN_DIR/config/app.toml
```
### Cosmovisor
Here we will show you two ways of using Cosmovisor to upgrade:
I. by manually preparing the new binary
II. by [auto-downloading](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor#auto-download) the new binary.

Method I requires node runners to manually build the old and new binary and put them into the `cosmovisor` folder (as shown below). Cosmovisor will then switch to the new binary upon upgrade height.
```shell
.
├── current -> genesis or upgrades/<name>
├── genesis
│   └── bin
│       └── gaiad
└── upgrades
    └── Vega
        ├── bin
        │   └── gaiad
        └── upgrade-info.json
```

Alternatively, by using the auto-download setup from method II, node runners do not need to prepare the new binaries manually. When proposing a new upgrade, the `--upgrade-info` flag should include a link of the new binary. The environmental variable `DAEMON_ALLOW_DOWNLOAD_BINARIES` is set to be true. Upon the upgrade height, if Cosmovisor cannot find the new binary locally, it will begin downloading the new binary according to the instructions found in `data/upgrade-info.json`(which is generated upon upgrade). After this it puts the new binary in `cosmovior/upgrade/Vega/bin` and starts running it.

*Note:*
- In general, auto-download comes with the risk that the verification of correct download is done automatically. If you want to have the highest guarantee you should confirm the check-sum manually. We hope more node operators will use the auto-download for this release but please be aware this is a risk you should take at your own discretion.
- For the Vega upgrade, Gaia will upgrade its dependency on Cosmos SDK v0.42 to Cosmos SDK v0.44, this will require [Cosmovisor v0.1](https://github.com/cosmos/cosmos-sdk/releases/tag/cosmovisor%2Fv0.1.0). Later versions of Cosmovisor do not support Cosmos SDK v0.42 or earlier if the auto-download option is enabled.
- By using Cosmovisor v0.1 you might experience a [node hanging issue](https://github.com/cosmos/cosmos-sdk/issues/9875) when querying a result with a large output size. For example, `gaiad q gov proposals` will hang the node being queried, this issue will not appear for Cosmovisor versions newer than v0.1.

#### Method I: manually prepare the new binary
##### Method I: setup Cosmosvisor
Create the Cosmovisor folders for validator 1 and validator 2, and put the old binary in `cosmovisor/genesis/bin`.
```shell
mkdir -p $VAL_1_CHAIN_DIR/cosmovisor/genesis/bin
mkdir -p $VAL_2_CHAIN_DIR/cosmovisor/genesis/bin
cp $(which gaiad) $VAL_1_CHAIN_DIR/cosmovisor/genesis/bin
cp $(which gaiad) $VAL_2_CHAIN_DIR/cosmovisor/genesis/bin
```

Build the new gaia binary
```shell
cd gaia
git checkout start-upgrade
make install
```
Create the folders for the two nodes and put the upgrade gaia binary into `cosmovisor/upgrades/Vega/bin`:
```shell
mkdir -p $VAL_1_CHAIN_DIR/cosmovisor/upgrades/Vega/bin
mkdir -p $VAL_2_CHAIN_DIR/cosmovisor/upgrades/Vega/bin
cp $(which gaiad) $VAL_1_CHAIN_DIR/cosmovisor/upgrades/Vega/bin
cp $(which gaiad) $VAL_2_CHAIN_DIR/cosmovisor/upgrades/Vega/bin
```
##### Method I: start Cosmovisor
For val1:
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME= $(pwd)/$VAL_1_CHAIN_DIR
export DAEMON_RESTART_AFTER_UPGRADE=true
cosmovisor start --x-crisis-skip-assert-invariants --home $VAL_1_CHAIN_DIR
```
For val2:

open a new terminal:
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME= $(pwd)/$VAL_2_CHAIN_DIR
export DAEMON_RESTART_AFTER_UPGRADE=true
cosmovisor start --x-crisis-skip-assert-invariants --home $VAL_2_CHAIN_DIR
```
##### Method I: propose upgrade
The user delegated to Validator 2 controls over 67% of the voting power. Normally a validator votes with the accumulated amount of stake from all delegators. But a user can also independently vote, and since this user is a whale, its vote will make any proposal pass.
```shell
cosmovisor tx gov submit-proposal software-upgrade Vega \
--title Vega \
--deposit 100uatom \
--upgrade-height 7368587 \
--upgrade-info "upgrade to Vega" \
--description "upgrade to Vega" \
--gas 400000 \
--from user \
--keyring-backend test \
--chain-id test \
--home data/test/val2 \
--node tcp://localhost:36657 \
--yes
```
##### Method I: Vote
Open a new terminal to vote by user.
```shell
cd vega-test
gaiad tx gov vote 54 yes \
--from user \
--keyring-backend test \
--chain-id test \
--home data/test/val2 \
--node tcp://127.0.0.1:36657 
--yes
```

After the voting period finishes, check the vote result by using:

```shell
$BINARY query gov proposal 54
```
The proposal status should be `PROPOSAL_STATUS_PASSED`.
```shell
content:
  '@type': /cosmos.upgrade.v1beta1.SoftwareUpgradeProposal
  description: upgrade to Vega
  plan:
    height: "7368587"
    info: upgrade to Vega
    name: Vega
    time: ""
    upgraded_client_state: null
  title: Vega
deposit_end_time: ""
final_tally_result:
  abstain: "0"
  "no": "0"
  no_with_veto: "0"
  "yes": "32099700"
proposal_id: "54"
status: PROPOSAL_STATUS_PASSED
submit_time: ""
total_deposit:
- amount: "100"
  denom: uatom
voting_end_time: ""
voting_start_time: ""
```
After `PROPOSAL_STATUS_PASSED`, wait until the upgrade height is reached, Cosmovisor will now apply the upgrade.

#### Method II: auto-download the new binary
##### Method II: setup Cosmosvisor

Create the Cosmovisor folder for for validator 1 and validator 2, and put the old binary in `cosmovisor/genesis/bin`.
```shell
mkdir -p $VAL_1_CHAIN_DIR/cosmovisor/genesis/bin
mkdir -p $VAL_2_CHAIN_DIR/cosmovisor/genesis/bin
cp $(which gaiad) $VAL_1_CHAIN_DIR/cosmovisor/genesis/bin
cp $(which gaiad) $VAL_2_CHAIN_DIR/cosmovisor/genesis/bin
```
##### Method II: start Cosmovisor
For validator 1:
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME= $(pwd)/$VAL_1_CHAIN_DIR
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
cosmovisor start --x-crisis-skip-assert-invariants --home $VAL_1_CHAIN_DIR
```
For validator 2:

open a new terminal:
```shell
export DAEMON_NAME=gaiad
export DAEMON_HOME= $(pwd)/$VAL_2_CHAIN_DIR
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
cosmovisor start --x-crisis-skip-assert-invariants --home $VAL_2_CHAIN_DIR
```
##### Method II: propose upgrade
With auto-download enabled, we can propose our upgrade with the `--upgrade-info` flag containing links to the new binaries from the [github releases](https://github.com/cosmos/gaia/releases). If you want to make sure that the binary downloaded is absolutely correct, we recommend you do your own checksum validation.

```shell
gaiad tx gov submit-proposal software-upgrade Vega \
--title Vega \
--deposit 100uatom \
--upgrade-height 7368587 \
--upgrade-info '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-linux-amd64?checksum=sha256:4b09b0466a6a8e1414e9f161b803023bec8c2fc8286bf1c4e4523a309e52a75c","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-linux-arm64?checksum=sha256:08b31fdd262d1576eae8bf0e7defb5ed680fafbf26bc85a1faec22c1e9d11a22","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-darwin-amd64?checksum=sha256:dddbaa50ab795e299ccb9f5bedab3f57965ab08d83103c5129bf7fc255950bb3","windows/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-windows-amd64.exe?checksum=sha256:e826dd52aee52e476da41b30fa488a50fa91b98e9937438eaa64d7b9cb4415dc"}}' \
--description "upgrade to Vega" \
--gas 400000 \
--from user \
--keyring-backend test \
--chain-id test \
--home data/test/val2 \
--node tcp://localhost:36657 \
--yes
```

##### Method II: Vote
Open a new terminal to vote by user.
```shell
cd vega-test
gaiad tx gov vote 54 yes \
--from user \
--keyring-backend test \
--chain-id test \
--home data/test/val2 \
--node tcp://127.0.0.1:36657 
--yes
```
After the voting period finishes, check the vote result by:

```shell
$BINARY query gov proposal 54
```
The proposal status should now be `PROPOSAL_STATUS_PASSED`.
```shell
content:
  '@type': /cosmos.upgrade.v1beta1.SoftwareUpgradeProposal
  description: upgrade to Vega
  plan:
    height: "7368587"
    info: '{"binaries":{"linux/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-linux-amd64?checksum=sha256:4b09b0466a6a8e1414e9f161b803023bec8c2fc8286bf1c4e4523a309e52a75c","linux/arm64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-linux-arm64?checksum=sha256:08b31fdd262d1576eae8bf0e7defb5ed680fafbf26bc85a1faec22c1e9d11a22","darwin/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-darwin-amd64?checksum=sha256:dddbaa50ab795e299ccb9f5bedab3f57965ab08d83103c5129bf7fc255950bb3","windows/amd64":"https://github.com/cosmos/gaia/releases/download/v6.0.0-rc3/gaiad-v6.0.0-rc3-windows-amd64.exe?checksum=sha256:e826dd52aee52e476da41b30fa488a50fa91b98e9937438eaa64d7b9cb4415dc"}}'
    name: Vega
    time: ""
    upgraded_client_state: null
  title: Vega
deposit_end_time: ""
final_tally_result:
  abstain: "0"
  "no": "0"
  no_with_veto: "0"
  "yes": "0"
proposal_id: "54"
status: PROPOSAL_STATUS_PASSED
submit_time: ""
total_deposit:
- amount: "100"
  denom: uatom
voting_end_time: ""
voting_start_time: ""
```

After `PROPOSAL_STATUS_PASSED`, wait until the upgrade height is reached. Cosmovisor will now auto-download the new binary specific to your platform and apply the upgrade. Please note, the upgrade info in method II does not contain the download link of the binary for GOOS=darwin GOARCH=arm64 (for mac M1 users), please use method I to upgrade.

## Upgrade result
### Method I:
When the upgrade block height is reached, you can find the following information in the log: `ERR UPGRADE "Vega" NEEDED at height: 7368587: upgrade to Vega` and `applying upgrade "Vega" at height:7368587`. The chain will continue to produce blocks after both nodes have completed the upgrade.
### Method II:
When the upgrade block height is reached, you can find the following information in the log: `INF applying upgrade "Vega" at height: 7368587`. The chain will continue to produce blocks after both nodes have completed the upgrade.

## Repeating the test

If you want to try running the test again, or if you made a mistake while running it the first time, make sure to do the following steps:

1. `cd vega-test`,`gaiad unsafe-reset-all --home data/test/val1` and `gaiad unsafe-reset-all --home data/test/val2`
2. If you ran the upgrade successfully with the Cosmovisor `--auto-download` flag enabled, make sure that you remove the downloaded binary from `data/test/val1/cosmovisor/upgrades/Vega` and `data/test/val2/cosmovisor/upgrades/Vega`. Also remove the symbolic link with `rm data/test/val1/cosmovisor/current` and `rm data/test/val2/cosmovisor/current`.
   Now you should be able to start the two binaries again and perform the upgrade from the start.

## Further info: test new modules
Now you can explore some of the functions of the new modules in Gaia.
For the AuthZ module, please refer to https://docs.cosmos.network/master/modules/authz/

## Reference

[Cosmovisor quick start](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)

[Changelog of cosmos-sdk v0.43.0](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/CHANGELOG.md#v0430---2021-08-10) and [Changelog of cosmos-sdk v0.44.0](https://github.com/cosmos/cosmos-sdk/blob/v0.44.0/CHANGELOG.md#v0440---2021-09-01) and [v0.44.1](https://github.com/cosmos/cosmos-sdk/blob/v0.44.1/CHANGELOG.md#v0441---2021-09-29)

[cosmos/ibc-go v1.2.1](https://github.com/cosmos/ibc-go/tree/v1.2.1)

[Gravity DEX Upgrade Simulation Test](https://github.com/b-harvest/gravity-dex-upgrade-test/blob/kogisin/v5.0.5-upgrade-simulation/v5.0.5/README.md)

[Cosmovisor](https://github.com/cosmos/cosmos-sdk/tree/master/cosmovisor)
## Changes in the genesis file

```diff
@@ -229195,10 +229195,10 @@
         {
           "@type": "/cosmos.auth.v1beta1.BaseAccount",
           "account_number": "27720",
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "pub_key": {
             "@type": "/cosmos.crypto.secp256k1.PubKey",
-            "key": "A6apc7iThbRkwboKqPy6eXxxQvTH+0lNkXZvugDM9V4g"
+            "key": "ApDOUyfcamDmnbEO7O4YKnKQQqQ93+gquLfGf7h5clX7"
           },
           "sequence": "221"
         },
@@ -3534263,7 +3534263,7 @@
           ]
         },
         {
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "coins": [
             {
               "amount": "10000000",
@@ -4039793,7 +4039793,7 @@
           "address": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
           "coins": [
             {
-              "amount": "194616098248861",
+              "amount": "6194616098248861",
               "denom": "uatom"
             }
           ]
@@ -5464401,7 +5464401,7 @@
           "denom": "poolF2805980C54E1474BDCCF70EF5FE881F3B8EFCF8BA3198765C01D91904521788"
         },
         {
-          "amount": "277834757180509",
+          "amount": "6277834757180509",
           "denom": "uatom"
         }
       ]
@@ -5565262,7 +5565262,7 @@
           "validator_address": "cosmosvaloper1pjmngrwcsatsuyy8m3qrunaun67sr9x7z5r2qs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "745",
@@ -5773162,7 +5773162,7 @@
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "465175",
@@ -6180718,7 +6180718,7 @@
           "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7274551",
             "previous_period": "82115",
@@ -6220192,7 +6220192,7 @@
           "validator_address": "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "1584",
@@ -6364934,7 +6364934,7 @@
           "starting_info": {
             "height": "0",
             "previous_period": "27294",
-            "stake": "25390741.000000000000000000"
+            "stake": "6000000025390741.000000000000000000"
           },
           "validator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf"
         },
@@ -6419947,7 +6419947,7 @@
           "validator_address": "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "5860",
@@ -6516184,7 +6516184,7 @@
           "validator_address": "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "80368",
@@ -6561175,7 +6561175,7 @@
           "validator_address": "cosmosvaloper1ey69r37gfxvxg62sh4r0ktpuc46pzjrm873ae8"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "269308",
@@ -6717784,7 +6717784,7 @@
           "validator_address": "cosmosvaloper1et77usu8q2hargvyusl4qzryev8x8t9wwqkxfs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "4225",
@@ -8636399,7 +8636399,7 @@
         "max_deposit_period": "1209600s",
         "min_deposit": [
           {
-            "amount": "64000000",
+            "amount": "1",
             "denom": "uatom"
           }
         ]
@@ -8637971,7 +8637971,7 @@
           "submit_time": "2021-06-02T17:30:15.614131648Z",
           "total_deposit": [
             {
-              "amount": "64000000",
+              "amount": "1",
               "denom": "uatom"
             }
           ],
@@ -8638092,7 +8638092,7 @@
           "submit_time": "2021-07-11T21:10:26.141197124Z",
           "total_deposit": [
             {
-              "amount": "64000000",
+              "amount": "1",
               "denom": "uatom"
             }
           ],
@@ -8638127,13 +8638127,13 @@
       ],
       "starting_proposal_id": "54",
       "tally_params": {
-        "quorum": "0.400000000000000000",
-        "threshold": "0.500000000000000000",
+        "quorum": "0.000000000000000001",
+        "threshold": "0.000000000000000001",
         "veto_threshold": "0.334000000000000000"
       },
       "votes": [],
       "voting_params": {
-        "voting_period": "1209600s"
+        "voting_period": "60s"
       }
     },
     "ibc": {
@@ -12473128,7 +12473128,7 @@
           ]
         },
         {
-          "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
+          "address": "cosmosvalcons10jc8h98awslz4pfqc26smf9sxaqxgwl4vxpcrp",
           "missed_blocks": [
             {
               "index": "10",
@@ -12976121,7 +12976121,7 @@
           ]
         },
         {
-          "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
+          "address": "cosmosvalcons16k44u3v0m8uevnh4p2qy2xm08y3wdf92xsc3ve",
           "missed_blocks": [
             {
               "index": "0",
@@ -14011121,7 +14011121,7 @@
           }
         },
         {
-          "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
+          "address": "cosmosvalcons10jc8h98awslz4pfqc26smf9sxaqxgwl4vxpcrp",
           "validator_signing_info": {
             "address": "",
             "index_offset": "7536348",
@@ -14011572,7 +14011572,7 @@
           }
         },
         {
-          "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
+          "address": "cosmosvalcons16k44u3v0m8uevnh4p2qy2xm08y3wdf92xsc3ve",
           "validator_signing_info": {
             "address": "",
             "index_offset": "10770382",
@@ -14062229,42 +14062229,42 @@
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "3000000.000000000000000000",
           "validator_address": "cosmosvaloper1pjmngrwcsatsuyy8m3qrunaun67sr9x7z5r2qs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "5000000.000000000000000000",
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "1000000.000000000000000000",
           "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "16100000.000000000000000000",
           "validator_address": "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "1000000.000000000000000000",
           "validator_address": "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "3000000.000000000000000000",
           "validator_address": "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "999999.999433247914197563",
           "validator_address": "cosmosvaloper1ey69r37gfxvxg62sh4r0ktpuc46pzjrm873ae8"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "2000000.000000000000000000",
           "validator_address": "cosmosvaloper1et77usu8q2hargvyusl4qzryev8x8t9wwqkxfs"
         },
@@ -14733575,7 +14733575,7 @@
--- original_v5_genesis.json    2021-09-02 11:59:06.000000000 +0200
+++ script/genesis.json 2021-09-02 12:29:51.000000000 +0200
@@ -229195,10 +229195,10 @@
         {
           "@type": "/cosmos.auth.v1beta1.BaseAccount",
           "account_number": "27720",
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "pub_key": {
             "@type": "/cosmos.crypto.secp256k1.PubKey",
-            "key": "A6apc7iThbRkwboKqPy6eXxxQvTH+0lNkXZvugDM9V4g"
+            "key": "ApDOUyfcamDmnbEO7O4YKnKQQqQ93+gquLfGf7h5clX7"
           },
           "sequence": "221"
         },
@@ -3534263,7 +3534263,7 @@
           ]
         },
         {
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "coins": [
             {
               "amount": "10000000",
@@ -4039793,7 +4039793,7 @@
           "address": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
           "coins": [
             {
-              "amount": "194616098248861",
+              "amount": "6194616098248861",
               "denom": "uatom"
             }
           ]
@@ -5464401,7 +5464401,7 @@
           "denom": "poolF2805980C54E1474BDCCF70EF5FE881F3B8EFCF8BA3198765C01D91904521788"
         },
         {
-          "amount": "277834757180509",
+          "amount": "6277834757180509",
--- original_v5_genesis.json    2021-09-02 11:59:06.000000000 +0200
+++ script/genesis.json 2021-09-02 12:29:51.000000000 +0200
@@ -229195,10 +229195,10 @@
         {
           "@type": "/cosmos.auth.v1beta1.BaseAccount",
           "account_number": "27720",
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "pub_key": {
             "@type": "/cosmos.crypto.secp256k1.PubKey",
-            "key": "A6apc7iThbRkwboKqPy6eXxxQvTH+0lNkXZvugDM9V4g"
+            "key": "ApDOUyfcamDmnbEO7O4YKnKQQqQ93+gquLfGf7h5clX7"
           },
           "sequence": "221"
         },
@@ -3534263,7 +3534263,7 @@
           ]
         },
         {
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "coins": [
             {
               "amount": "10000000",
@@ -4039793,7 +4039793,7 @@
           "address": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
           "coins": [
             {
-              "amount": "194616098248861",
+              "amount": "6194616098248861",
               "denom": "uatom"
             }
           ]
@@ -5464401,7 +5464401,7 @@
           "denom": "poolF2805980C54E1474BDCCF70EF5FE881F3B8EFCF8BA3198765C01D91904521788"
         },
         {
-          "amount": "277834757180509",
+          "amount": "6277834757180509",
           "denom": "uatom"
         }
       ]
@@ -5565262,7 +5565262,7 @@
           "validator_address": "cosmosvaloper1pjmngrwcsatsuyy8m3qrunaun67sr9x7z5r2qs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
         },
         {
           "delegator_address": "cosmos1ll705078lwg6yksn3flktpvzpe56gwvh7xmynw",
-          "shares": "25390741.000000000000000000",
+          "shares": "6000000025390741.000000000000000000",
           "validator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf"
         },
         {
@@ -14733640,7 +14733640,7 @@
         }
       ],
       "exported": true,
-      "last_total_power": "194616038",
+      "last_total_power": "6194616038",
       "last_validator_powers": [
         {
           "address": "cosmosvaloper1qwl879nx9t6kef4supyazayf7vjhennyh568ys",
@@ -14733952,7 +14733952,7 @@
         },
         {
           "address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf",
-          "power": "13944328"
+          "power": "6013944328"
         },
         {
           "address": "cosmosvaloper15urq2dtp9qce4fyc85m6upwm9xul3049e02707",
@@ -14734698,7 +14734698,7 @@
           "validator_src_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "entries": [
             {
               "completion_time": "2021-09-02T16:58:32.491727097Z",
@@ -14830334,7 +14830334,7 @@
           },
           "consensus_pubkey": {
             "@type": "/cosmos.crypto.ed25519.PubKey",
-            "key": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
+            "key": "qwiUMxz3llsy45fPvM0a8+XQeAJLvrX3QAEJmRMEEoU="
           },
           "delegator_shares": "2656249798904.000000000000000000",
           "description": {
@@ -14835612,9 +14835612,9 @@
           },
           "consensus_pubkey": {
             "@type": "/cosmos.crypto.ed25519.PubKey",
-            "key": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
+            "key": "oi55Dw+JjLQc4u1WlAS3FsGwh5fd5/N5cP3VOLnZ/H0="
           },
-          "delegator_shares": "13944328343563.000000000000000000",
+          "delegator_shares": "6013944328343563.000000000000000000",
           "description": {
             "details": "Exchange the world",
             "identity": "",
@@ -14835626,7 +14835626,7 @@
           "min_self_delegation": "1",
           "operator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf",
           "status": "BOND_STATUS_BONDED",
-          "tokens": "13944328343563",
+          "tokens": "6013944328343563",
           "unbonding_height": "0",
           "unbonding_time": "1970-01-01T00:00:00Z"
         },
@@ -14838926,7 +14838926,7 @@
     "upgrade": {},
     "vesting": {}
   },
-  "chain_id": "cosmoshub-4",
+  "chain_id": "test",
   "consensus_params": {
     "block": {
       "max_bytes": "200000",
@@ -14838949,12 +14838949,12 @@
   "initial_height": "7368387",
   "validators": [
     {
-      "address": "B00A6323737F321EB0B8D59C6FD497A14B60938A",
+      "address": "D5AB5E458FD9F9964EF50A80451B6F3922E6A4AA",
       "name": "Certus One",
       "power": "2656249",
       "pub_key": {
         "type": "tendermint/PubKeyEd25519",
-        "value": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
+        "value": "qwiUMxz3llsy45fPvM0a8+XQeAJLvrX3QAEJmRMEEoU="
       }
     },
     {
@@ -14839642,12 +14839642,12 @@
       }
     },
     {
-      "address": "83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0",
+      "address": "7CB07B94FD743E2A8520C2B50DA4B03740643BF5",
       "name": "Binance Staking",
-      "power": "13944328",
+      "power": "6013944328",
       "pub_key": {
         "type": "tendermint/PubKeyEd25519",
-        "value": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
+        "value": "oi55Dw+JjLQc4u1WlAS3FsGwh5fd5/N5cP3VOLnZ/H0="
       }
     },
     {
➜  gaia_data touch diff.md
➜  gaia_data diff -u original_v5_genesis.json script/genesis.json >diff.md
➜  gaia_data vim diff.md
➜  gaia_data cat diff.md
--- original_v5_genesis.json	2021-09-02 11:59:06.000000000 +0200
+++ script/genesis.json	2021-09-02 12:29:51.000000000 +0200
@@ -229195,10 +229195,10 @@
         {
           "@type": "/cosmos.auth.v1beta1.BaseAccount",
           "account_number": "27720",
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "pub_key": {
             "@type": "/cosmos.crypto.secp256k1.PubKey",
-            "key": "A6apc7iThbRkwboKqPy6eXxxQvTH+0lNkXZvugDM9V4g"
+            "key": "ApDOUyfcamDmnbEO7O4YKnKQQqQ93+gquLfGf7h5clX7"
           },
           "sequence": "221"
         },
@@ -3534263,7 +3534263,7 @@
           ]
         },
         {
-          "address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "coins": [
             {
               "amount": "10000000",
@@ -4039793,7 +4039793,7 @@
           "address": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
           "coins": [
             {
-              "amount": "194616098248861",
+              "amount": "6194616098248861",
               "denom": "uatom"
             }
           ]
@@ -5464401,7 +5464401,7 @@
           "denom": "poolF2805980C54E1474BDCCF70EF5FE881F3B8EFCF8BA3198765C01D91904521788"
         },
         {
-          "amount": "277834757180509",
+          "amount": "6277834757180509",
           "denom": "uatom"
         }
       ]
@@ -5565262,7 +5565262,7 @@
           "validator_address": "cosmosvaloper1pjmngrwcsatsuyy8m3qrunaun67sr9x7z5r2qs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "745",
@@ -5773162,7 +5773162,7 @@
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "465175",
@@ -6180718,7 +6180718,7 @@
           "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7274551",
             "previous_period": "82115",
@@ -6220192,7 +6220192,7 @@
           "validator_address": "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "1584",
@@ -6364934,7 +6364934,7 @@
           "starting_info": {
             "height": "0",
             "previous_period": "27294",
-            "stake": "25390741.000000000000000000"
+            "stake": "6000000025390741.000000000000000000"
           },
           "validator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf"
         },
@@ -6419947,7 +6419947,7 @@
           "validator_address": "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "5860",
@@ -6516184,7 +6516184,7 @@
           "validator_address": "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "80368",
@@ -6561175,7 +6561175,7 @@
           "validator_address": "cosmosvaloper1ey69r37gfxvxg62sh4r0ktpuc46pzjrm873ae8"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "269308",
@@ -6717784,7 +6717784,7 @@
           "validator_address": "cosmosvaloper1et77usu8q2hargvyusl4qzryev8x8t9wwqkxfs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "starting_info": {
             "height": "7248467",
             "previous_period": "4225",
@@ -8636399,7 +8636399,7 @@
         "max_deposit_period": "1209600s",
         "min_deposit": [
           {
-            "amount": "64000000",
+            "amount": "1",
             "denom": "uatom"
           }
         ]
@@ -8637971,7 +8637971,7 @@
           "submit_time": "2021-06-02T17:30:15.614131648Z",
           "total_deposit": [
             {
-              "amount": "64000000",
+              "amount": "1",
               "denom": "uatom"
             }
           ],
@@ -8638092,7 +8638092,7 @@
           "submit_time": "2021-07-11T21:10:26.141197124Z",
           "total_deposit": [
             {
-              "amount": "64000000",
+              "amount": "1",
               "denom": "uatom"
             }
           ],
@@ -8638127,13 +8638127,13 @@
       ],
       "starting_proposal_id": "54",
       "tally_params": {
-        "quorum": "0.400000000000000000",
-        "threshold": "0.500000000000000000",
+        "quorum": "0.000000000000000001",
+        "threshold": "0.000000000000000001",
         "veto_threshold": "0.334000000000000000"
       },
       "votes": [],
       "voting_params": {
-        "voting_period": "1209600s"
+        "voting_period": "60s"
       }
     },
     "ibc": {
@@ -12473128,7 +12473128,7 @@
           ]
         },
         {
-          "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
+          "address": "cosmosvalcons10jc8h98awslz4pfqc26smf9sxaqxgwl4vxpcrp",
           "missed_blocks": [
             {
               "index": "10",
@@ -12976121,7 +12976121,7 @@
           ]
         },
         {
-          "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
+          "address": "cosmosvalcons16k44u3v0m8uevnh4p2qy2xm08y3wdf92xsc3ve",
           "missed_blocks": [
             {
               "index": "0",
@@ -14011121,7 +14011121,7 @@
           }
         },
         {
-          "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
+          "address": "cosmosvalcons10jc8h98awslz4pfqc26smf9sxaqxgwl4vxpcrp",
           "validator_signing_info": {
             "address": "",
             "index_offset": "7536348",
@@ -14011572,7 +14011572,7 @@
           }
         },
         {
-          "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
+          "address": "cosmosvalcons16k44u3v0m8uevnh4p2qy2xm08y3wdf92xsc3ve",
           "validator_signing_info": {
             "address": "",
             "index_offset": "10770382",
@@ -14062229,42 +14062229,42 @@
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "3000000.000000000000000000",
           "validator_address": "cosmosvaloper1pjmngrwcsatsuyy8m3qrunaun67sr9x7z5r2qs"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "5000000.000000000000000000",
           "validator_address": "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "1000000.000000000000000000",
           "validator_address": "cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "16100000.000000000000000000",
           "validator_address": "cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "1000000.000000000000000000",
           "validator_address": "cosmosvaloper1hjct6q7npsspsg3dgvzk3sdf89spmlpfdn6m9d"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "3000000.000000000000000000",
           "validator_address": "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "999999.999433247914197563",
           "validator_address": "cosmosvaloper1ey69r37gfxvxg62sh4r0ktpuc46pzjrm873ae8"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "shares": "2000000.000000000000000000",
           "validator_address": "cosmosvaloper1et77usu8q2hargvyusl4qzryev8x8t9wwqkxfs"
         },
@@ -14733575,7 +14733575,7 @@
         },
         {
           "delegator_address": "cosmos1ll705078lwg6yksn3flktpvzpe56gwvh7xmynw",
-          "shares": "25390741.000000000000000000",
+          "shares": "6000000025390741.000000000000000000",
           "validator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf"
         },
         {
@@ -14733640,7 +14733640,7 @@
         }
       ],
       "exported": true,
-      "last_total_power": "194616038",
+      "last_total_power": "6194616038",
       "last_validator_powers": [
         {
           "address": "cosmosvaloper1qwl879nx9t6kef4supyazayf7vjhennyh568ys",
@@ -14733952,7 +14733952,7 @@
         },
         {
           "address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf",
-          "power": "13944328"
+          "power": "6013944328"
         },
         {
           "address": "cosmosvaloper15urq2dtp9qce4fyc85m6upwm9xul3049e02707",
@@ -14734698,7 +14734698,7 @@
           "validator_src_address": "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c"
         },
         {
-          "delegator_address": "cosmos1z98eg2ztdp2glyla62629nrlvczg8s7f0tm3dx",
+          "delegator_address": "cosmos1wvvhhfm387xvfnqshmdaunnpujjrdxznr5d5x9",
           "entries": [
             {
               "completion_time": "2021-09-02T16:58:32.491727097Z",
@@ -14830334,7 +14830334,7 @@
           },
           "consensus_pubkey": {
             "@type": "/cosmos.crypto.ed25519.PubKey",
-            "key": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
+            "key": "qwiUMxz3llsy45fPvM0a8+XQeAJLvrX3QAEJmRMEEoU="
           },
           "delegator_shares": "2656249798904.000000000000000000",
           "description": {
@@ -14835612,9 +14835612,9 @@
           },
           "consensus_pubkey": {
             "@type": "/cosmos.crypto.ed25519.PubKey",
-            "key": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
+            "key": "oi55Dw+JjLQc4u1WlAS3FsGwh5fd5/N5cP3VOLnZ/H0="
           },
-          "delegator_shares": "13944328343563.000000000000000000",
+          "delegator_shares": "6013944328343563.000000000000000000",
           "description": {
             "details": "Exchange the world",
             "identity": "",
@@ -14835626,7 +14835626,7 @@
           "min_self_delegation": "1",
           "operator_address": "cosmosvaloper156gqf9837u7d4c4678yt3rl4ls9c5vuursrrzf",
           "status": "BOND_STATUS_BONDED",
-          "tokens": "13944328343563",
+          "tokens": "6013944328343563",
           "unbonding_height": "0",
           "unbonding_time": "1970-01-01T00:00:00Z"
         },
@@ -14838926,7 +14838926,7 @@
     "upgrade": {},
     "vesting": {}
   },
-  "chain_id": "cosmoshub-4",
+  "chain_id": "test",
   "consensus_params": {
     "block": {
       "max_bytes": "200000",
@@ -14838949,12 +14838949,12 @@
   "initial_height": "7368387",
   "validators": [
     {
-      "address": "B00A6323737F321EB0B8D59C6FD497A14B60938A",
+      "address": "D5AB5E458FD9F9964EF50A80451B6F3922E6A4AA",
       "name": "Certus One",
       "power": "2656249",
       "pub_key": {
         "type": "tendermint/PubKeyEd25519",
-        "value": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
+        "value": "qwiUMxz3llsy45fPvM0a8+XQeAJLvrX3QAEJmRMEEoU="
       }
     },
     {
@@ -14839642,12 +14839642,12 @@
       }
     },
     {
-      "address": "83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0",
+      "address": "7CB07B94FD743E2A8520C2B50DA4B03740643BF5",
       "name": "Binance Staking",
-      "power": "13944328",
+      "power": "6013944328",
       "pub_key": {
         "type": "tendermint/PubKeyEd25519",
-        "value": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
+        "value": "oi55Dw+JjLQc4u1WlAS3FsGwh5fd5/N5cP3VOLnZ/H0="
       }
     },
     {
```
