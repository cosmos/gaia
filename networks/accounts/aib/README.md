# AiB

Fetch the allocation prepared by AiB. This should be on `master` branch but for
the avoidance of doubt, it's commit `d3671fce2b754445eefb815b2d56ba666e2bc4c4`:

```
curl https://raw.githubusercontent.com/cosmos/fundraiser-lib/d3671fce2b754445eefb815b2d56ba666e2bc4c4/src/atom_query/data/aib_atoms.final.json > orig.json
```

Copy the multisig piece into `multisig.json`, and add the address
(`cosmos176m2p8l3fps3dal7h8gf9jvrv98tu3rqfdht86`). 
Copy the rest to `employees.json` as a list.

Run `go run main.go` to get a summary, check values are positive, check duplicates, and check the multisig
address.
