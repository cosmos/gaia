# Public

Here we recompute the results from the fundraiser using `blockchain.info` and
`etherscan.io` to access the Bitcoin and Ethereum blockchains.

The result is the same data that's been in
`https://github.com/cosmos/fundraiser-lib/blob/master/src/atom_query/data/fundraiser_atoms.json`
since shortly after the fundraiser.

## Bitcoin

Fetch block data from blockchain.info and save in `btc_donations.json`

```
go run btc_main.go data
```

Convert the raw donation data into an `address->atom` map and save in
`btc_atoms.json`:

```
go run btc_main.go atoms
```

## Eth


Fetch event data from etherscan.io and save in `eth_donations.json`

```
go run eth_main.go data
```

Convert the raw donation data into an `address->atom` map and save in
`eth_atoms.json`:

```
go run eth_main.go atoms
```

## Combine

Combine them into `contributors.json`:

```
go run main.go
```
