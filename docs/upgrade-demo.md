# Upgrade demo

Compile gaia from https://github.com/regen-friends/gaia/tree/gaia-upgrade-rebased using 
```make install```

Setup local testnet. You can use the following script for this.
```
# You can run all of these commands from your home directory
cd $HOME

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


Change ```voting_params.voting_period``` in ```genesis.json``` to a reduced time ~5 minutes(300000000000)

Start the testnet using
```gaiad start```

Create a proposal using 
```gaiacli tx gov submit-proposal software-upgrade test1 --title "test1" --description "upgrade"  --from validator --chain-id testing --upgrade-height 200 --deposit 1000stake -y```

Vote the proposal using 
```gaiacli tx gov vote 1 yes --from validator --trust-node --chain-id testing -y```

Query the proposal after voting time ends to see if it passed using 
```gaiacli query gov proposal 1 --trust-node```

If the proposal passes, the chain will stop at given upgrade height.

In ```gaia/app/app.go```, uncomment upgrade handler on line:155. Make sure the upgrade title in handler matches the title from proposal.

Create a new binary of gaia with added upgrade handler using
```make install```

Restart the chain using new binary, you should see  the chain resume from the upgrade height 
```gaiad start```



