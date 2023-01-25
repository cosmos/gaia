# gaia-8000-rc1

Internal testnet only. You should be running the SDK version tagged `v0.24.0-rc1`:

```bash
$ gaiad version
0.24.0-1ae54661
```

Submit genesis transactions to this folder, as `[moniker].json`, in a new PR to this repo:

```bash
gaiad init gen-tx --name [name]
```

Make sure to only copy the `gen-tx-file` substructure!
