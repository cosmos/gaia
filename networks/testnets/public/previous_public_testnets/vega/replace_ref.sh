#!/bin/bash
sed -i '' 's%"chain_id": "cosmoshub-4",%"chain_id": "vega-testnet",%g' genesis.json &&

# substitute "Binance Staking", this is our val1
# tendermint pub_key
sed -i '' 's%W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o=%bf5gFMl/dQxJFVE4jReOYxbVeux8UcFJ9lj1+qDZDGs=%g' genesis.json &&
# priv_val address
sed -i '' 's%83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0%7C9F0FADF306FED9663F811619141F99147E6722%g' genesis.json &&
#  Validator consensus address, try command ` gaiad keys parse 83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0` to see if you can get the same addr.
sed -i '' 's%cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3%cosmosvalcons10j0slt0nqmldje3lsytpj9qlny28ueezg92w6g%g' genesis.json &&


# substitute "Certus One", this is our  val2
sed -i '' 's%cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM=%p6ihCq31IZUeY6z00G9ROoHTphnhi1J7wFrZ+5F2epU=%g' genesis.json &&
sed -i '' 's%B00A6323737F321EB0B8D59C6FD497A14B60938A%17472F7923685922F8165C33F4E02176189AF253%g' genesis.json &&
sed -i '' 's%cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6%cosmosvalcons1zarj77frdpvj97qktselfcppwcvf4ujn35a2ps%g' genesis.json &&


# substitute "Coinbase Custody", this is our  val3
sed -i '' 's%NK3/1mb/ToXmxlcyCK8HYyudDn4sttz1sXyyD+42x7I=%un9hBl/53UOx5oFOu7+eOY1C0wOsdoVDfUW5VCH8TyA=%g' genesis.json &&
sed -i '' 's%F8C01C0681578AA700D736D675C9992065F65E3E%132BAA3FAD92DDB2A23BB1FC8144F2D04F16DCC8%g' genesis.json &&
sed -i '' 's%cosmosvalcons1lrqpcp5p2792wqxhxmt8tjveypjlvh378gkddu%cosmosvalcons1zv4650adjtwm9g3mk87gz38j6p83dhxgytprlj%g' genesis.json &&


# substitute the faucet-user, faucet-user delegates to binance
sed -i '' 's%cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4%cosmos10aak94tfdl3pgt8qe6ga75qh3zkf3anpq8aqg0%g' genesis.json &&
sed -i '' 's%AjEkAHzQakRnyUppiM5/hnA6h2D7NkdxExxgiCG+NiDh%A81DhG/5sB6RA8dl/6jtmX0svTc0xJL5NjPPI/q4jJWP%g' genesis.json

# fix faucet-user's balance, increase by 300,000,000,000,000 uatom
sed -i '' 's%"amount": "160896"%"amount": "300000000160896"%g' genesis.json &&


# change one delegator's(cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4) delegation. This delegator delegates to our val1(Binance Staking). Increase this stake by 6,000,000,000,000,000.
sed -i '' 's%"11316631.000000000000000000"%"6000000011316631.000000000000000000"%g' genesis.json &&

# fix power of the validator
# Binance Staking validator's "delegator_shares" and "tokens"
# Increase the "delegator_shares" by 6,000,000,000,000,000 correspondingly.
sed -i '' 's%13944328343563%6013944328343563%g' genesis.json &&
# Increase the validator power by 6,000,000,000
sed -i '' 's%"power": "13944328"%"power": "6013944328"%g' genesis.json &&

# fix last_total_power
# Increase total amounts of bonded tokens recorded during the previous end block by 6,000,000,000
sed -i '' 's%"194616038"%"6194616038"%g' genesis.json &&

# fix total supply of uatom
sed -i '' 's%277834757180509%6577834757180509%g' genesis.json &&

# fix balance of bonded_tokens_pool module account
# module account for recording Binance staking(val1)'s received delegations:
# cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh
# Increase the delegation account by 6,000,000,000,000,000
sed -i '' 's%194616098248861%6194616098248861%g' genesis.json &&




# change gov params
# make voting period 24h
sed -i '' 's%"voting_period": "1209600s"%"voting_period": "86400s"%g' genesis.json
