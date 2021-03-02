# Appendix

## Increase your stake

To increase your ualtg stake, if you have extra tokens lying around. The first command will show an output like this, you want to take the key starting with cosmosvaloper1 in the 'address' field.

```
- name: jkilpatr
  type: local
  address: cosmosvaloper1jpz0ahls2chajf78nkqczdwwuqcu97w6z3plt4
  pubkey: cosmosvaloperpub1addwnpepqvl0qgfqewmuqvyaskmr4pwkr5fwzuk8286umwrfnxqkgqceg6ksu359m5q
  mnemonic: ""
  threshold: 0
  pubkeys: []

```

```
althea keys show myvalidatorkeyname --bech val
althea tx staking delegate <the address from the above command> 99000000ualtg --from myvalidatorkeyname --chain-id althea-testnet1v4 --fees 50ualtg --broadcast-mode block
```

## Unjail your validator

You can be jailed for several different reasons. As part of the Althea testnet we are testing slashing conditions for the Gravity bridge, so you will be slashed if the Orchestrator is not running properly, in addition to the usual Cosmos double sign and downtime slashing parameters. To unjail your validator run

```
althea tx slashing unjail --from myvalidatorkeyname --chain-id=althea-testnet1v4
```

## Unjail yourself

This command will unjail you, completing the process of getting the chain back online!

_replace 'myvalidatorkeyname' with your validator keys name, if you don't remember run `althea keys list`_

```
althea tx slashing unjail --from myvalidatorkeyname --chain-id=althea-testnet1v4
```
