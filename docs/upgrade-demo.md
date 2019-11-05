# Upgrade demo

Compile gaia from https://github.com/regen-friends/gaia/tree/gaia-upgrade-rebased using `make install`

## Prepare demo genesis 

Setup local testnet. You can use the following script for this.

```
# You can run all of these commands from your home directory
cd $HOME

# Remove any existing gaiacli and gaiad folders - check there is no useful data there first
rm -rf .gaiacli/ .gaiad/

# Initialize the genesis.json file that will help you to bootstrap the network
gaiad init --chain-id=testing testing

# Create a key to hold your validator account
gaiacli keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
gaiad add-genesis-account $(gaiacli keys show validator -a) 1000000000stake,1000000000validatortoken

# Generate the transaction that creates your validator
gaiad gentx --name validator

# Add the generated bonding transaction to the genesis file
gaiad collect-gentxs
```

**Important** Change `voting_params.voting_period` in `.gaiad/config/genesis.json` to a reduced time ~5 minutes(300000000000)

## Start the network and trigger upgrade

```
# Start the testnet
gaiad start

# Set up the cli config
gaiacli config trust-node true
gaiacli config chain-id testing

# Create a proposal
gaiacli tx gov submit-proposal software-upgrade test1 --title "test1" --description "upgrade"  --from validator --upgrade-height 200 --deposit 10000000stake -y

# Ensure voting_start_time is set (sufficient deposit)
gaiacli query gov proposal 1

# Vote the proposal 
gaiacli tx gov vote 1 yes --from validator -y

#Wait for the voting period to pass...
sleep 500

# Query the proposal after voting time ends to see if it passed using 
gaiacli query gov proposal 1

# Query the pending plan
gaiacli query upgrade plan

# Verify the bonus account (see commented upgrade handler) is empty
gaiacli query account cosmos18cgkqduwuh253twzmhedesw3l7v3fm37sppt58
```

## Performing an upgrade

Assuming you voted properly above, the proposal will pass, and the chain will stop at given upgrade height.
You can stop and start the original binary all you want, but it will refuse to run after the upgrade height.
We need a new binary with the upgrade handler installed. The logs should look something like:

```
E[2019-11-05|12:44:18.913] UPGRADE "test1" NEEDED at height: 200:       module=main 
E[2019-11-05|12:44:18.914] CONSENSUS FAILURE!!!
...
```

Note that the process just hangs, doesn't exit to avoid restart loops. You must manually kill the process and replace it with a new
binary. Do so now with `Ctrl+C` or `killall gaiad`.

In `gaia/app/app.go`, uncomment upgrade handler on lines 170-176. Make sure the upgrade title in handler matches the title from proposal.

```
# Create a new binary of gaia with added upgrade handler using
make install

# Restart the chain using new binary, you should see  the chain resume from the upgrade height
# Like `I[2019-11-05|12:48:15.184] applying upgrade "test1" at height: 200      module=main`
gaiad start

# Verify no more pending plan
gaiacli query upgrade plan

# You can query the block header of the completed upgrade
gaiacli query upgrade applied test1

# Verify the bonus account (see commented upgrade handler) is now rich
gaiacli query account cosmos18cgkqduwuh253twzmhedesw3l7v3fm37sppt58
```