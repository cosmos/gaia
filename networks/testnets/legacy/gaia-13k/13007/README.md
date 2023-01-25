# Gaia-13007

- Go version: [v1.13+](https://golang.org/dl/)
- Gaia version: [v2.0.3+](https://github.com/cosmos/gaia/releases)
- Seeds:

```
055a315b20c847813535d7c2b4cedba5756e3d79@207.180.204.112:26656
444d209bd0f89d7bf18cf389a74872e7082b237e@44.230.205.153:26656
30e46db6f9e6f5f19d1c08785faec03616024759@51.68.102.106:26656
04c28a44dd4eac4961c748bbe5451f7cdd12205c@18.217.97.195:26656
```

## GenTx Generation

1. Initialize the gaia directories and create the local genesis file with the correct
   chain-id

   ```shell
   $ gaiad init monikername --chain-id=gaia-13007
   ```

2. Create a local key pair in the Keybase

   ```shell
   $ gaiacli keys add <key-name>
   ```

3. Add your account to your local genesis file with a given amount and the key you
   just created.

   ```shell
   $ gaiad add-genesis-account $(gaiacli keys show <key-name> -a) 50000000000umuon
   ```

4. Create the gentx

   ```shell
   $ gaiad gentx --amount 50000000000umuon \
     --commission-rate=<rate> \
     --commission-max-rate=<max-rate> \
     --commission-max-change-rate=<max-change-rate-rate> \
     --pubkey $(gaiad tendermint show-validator) \
     --name=<key-name>
   ```
