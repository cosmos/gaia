# Token Factory

[Audited](https://hashlock.com/wp-content/uploads/2024/11/Manifest-Token-Factory-Smart-Contract-Audit-Report-Revised-Report-v2.pdf) by [HashLock](https://hashlock.com/audits/manifest) (November 2024)

</br>

![GitHub commit check runs](https://img.shields.io/github/check-runs/strangelove-ventures/tokenfactory/main)
![Codecov](https://img.shields.io/codecov/c/github/strangelove-ventures/tokenfactory)

The `tokenfactory` module allows any account to create a new token with the name `factory/{creator address}/{subdenom}`. Because tokens are namespaced by creator address, this allows token minting to be permissionless, due to not needing to resolve name collisions. A single account can create multiple denoms, by providing a unique subdenom for each created denom. Once a denom is created, the original creator is given "admin" privileges over the asset. This allows them to:

- Mint their denom to any account
- Burn their denom from any account
- Create a transfer of their denom between any two accounts
- Change the admin. The `ChangeAdmin` functionality allows changing the master admin account, or even setting it to "", meaning no account has admin privileges of the asset.

## Supported Versions

| Tokenfactory | Cosmos-SDK | wasmvm | branch
| ------------ | ---------- | ------ | ------ |
| v0.50.X-wasmvm2 | v0.50.X | v2 | [main](https://github.com/strangelove-ventures/tokenfactory/tree/main) |
| v0.50.X       | v0.50.X   | v1 | [main_wasmvm1](https://github.com/strangelove-ventures/tokenfactory/tree/main_wasmvm1) |

## References

- Osmosis Labs [TokenFactory](https://github.com/osmosis-labs/osmosis/tree/main/x/tokenfactory)

## Installation

This repository provides the `tokend` application, a simple command-line application that demonstrates the functionality of the `tokenfactory` module. To build and install the `tokend` application, run:

```bash
make install
```

The application can be run with the following command:

```bash
tokend
```

## Test Node

To run a test node with the `tokenfactory` module, run:

```bash
make install
CHAIN_ID="local-1" HOME_DIR="~/.tokenfactory" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh
```

## Usage

The `tokend` application provides the following `tokenfactory` transactions:

- `create-denom`: Create a new denom with the name `factory/{creator address}/{subdenom}`.
- `mint`: Mint tokens to your address. You must be the admin of the denom to mint tokens.
- `mint-to`: Mint tokens to another address. You must be the admin of the denom to mint tokens.
- `burn`: Burn tokens from your address. You must be the admin of the denom to burn tokens.
- `burn-from`: Burn tokens from another address. You must be the admin of the denom to burn tokens.
- `force-transfer`: Transfer tokens between two addresses. You must be the admin of the denom to transfer tokens.
- `change-admin`: Change the admin of the denom. You must be the admin of the denom to change the admin.
- `modify-metadata`: Modify the metadata of the denom. You must be the admin of the denom to modify the metadata.

and the following queries:

- `params`: Get the tokenfactory module parameters.
- `denom-authority-metadata`: Get the authority metadata of a denom.
- `denoms-from-creator`: Returns a list of all denoms created by a given creator.
- `denoms-from-admin`: Returns a list of all denoms for which a given address is the admin.

## Testing

To test the `tokenfactory` module, run:

```bash
make test
```

To run the `tokenfactory` module's end-to-end tests, run:
```bash
make local-image
make ictest-tokenfactory
```

## Coverage

To generate the coverage report of the `tokenfactory` module, run:

```bash
make local-image
make coverage
```

## Simulation

To run the `tokenfactory` module's full application simulation tests, run:

```bash
make sim-full-app
```

To run the `tokenfactory` simulation after state import test, run:

```bash
make sim-after-import
```

To run the `tokenfactory` application determinism simulation, run:

```bash
make sim-app-determinism
```

Append `-random` to the end of the commands above to run the simulation with a random seed, e.g., `make sim-full-app-random`.

## Examples

### Create Denom

```bash
# Usage:
#   tokend tx tokenfactory create-denom [subdenom] [flags]
tokend tx tokenfactory create-denom utest --from alice

# Query the newly created token
# cosmos1... is the creator address of the denom (alice)
tokend q tokenfactory denoms-from-creator cosmos1...
denoms:
- factory/cosmos1.../utest

```

### Modify Metadata

```bash
# Usage:
#   tokend tx tokenfactory modify-metadata [denom] [ticker-symbol] [description] [exponent] [flags]

# Modify the metadata of the utest denom
# cosmos1... is the creator address of the denom (alice)
# The ticker symbol is TST and the description is "My token description"
# The denom exponent is 6
tokend tx tokenfactory modify-metadata factory/cosmos1.../utest TST "My token description" 6 --from alice

# Query the authority metadata of the factory/cosmos1.../utest denom
tokend q tokenfactory denom-authority-metadata factory/cosmos1.../utest
authority_metadata:
  admin: cosmos1...

# Query the denom metadata from the bank module
tokend q bank denom-metadata factory/cosmos1.../utest
metadata:
  base: factory/cosmos1.../utest
  denom_units:
  - aliases:
    - TST
    denom: factory/cosmos1.../utest
  - aliases:
    - factory/cosmos1.../utest
    denom: TST
    exponent: 6
  description: My token description
  display: TST
  name: factory/cosmos1.../utest
  symbol: TST
```

### Mint

```bash
# Usage:
#   tokend tx tokenfactory mint [amount] [flags]

# Mint 1000 tokens to alice
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory mint 1000factory/cosmos1.../utest --from alice

# Query the account balance of alice
tokend q bank balance cosmos1... factory/cosmos1.../utest
balance:
  amount: "1000"
  denom: factory/cosmos1.../utest
```

### Mint To

```bash
# Usage:
#   tokend tx tokenfactory mint-to [address] [amount] [flags]

# Mint 2000 tokens to bob
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory mint-to [bob-addr] 2000factory/cosmos1.../utest --from alice

# Query the account balance of bob
tokend q bank balance [bob-addr] factory/cosmos1.../utest
balance:
  amount: "2000"
  denom: factory/cosmos1.../utest
```

### Burn

```bash
# Usage:
#   tokend tx tokenfactory burn [amount] [flags]

# Burn 500 tokens from alice
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory burn 500factory/cosmos1.../utest --from alice

# Query the account balance of alice
tokend q bank balance cosmos1... factory/cosmos1.../utest
balance:
  amount: "500"
  denom: factory/cosmos1.../utest
```

### Burn From

```bash
# Usage:
#   tokend tx tokenfactory burn-from [address] [amount] [flags]

# Burn 500 tokens from bob
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory burn-from [bob-addr] 500factory/cosmos1.../utest --from alice

# Query the account balance of bob
tokend q bank balance [bob-addr] factory/cosmos1.../utest
balance:
  amount: "1500"
  denom: factory/cosmos1.../utest
```

### Force Transfer

```bash
# Usage:
#   tokend tx tokenfactory force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]

# Transfer 500 tokens from bob to alice
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory force-transfer 50factory/cosmos1.../utest [bob-addr] cosmos1... --from alice

# Query the account balance of alice
tokend q bank balance cosmos1... factory/cosmos1.../utest
balance:
  amount: "550"
  denom: factory/cosmos1.../utest

# Query the account balance of bob
tokend q bank balance [bob-addr] factory/cosmos1.../utest
balance:
  amount: "1450"
  denom: factory/cosmos1.../utest
```

### Change Admin

```bash
# Usage:
#   tokend tx tokenfactory change-admin [denom] [new-admin-address] [flags]

# Change the admin of the utest denom to bob
# cosmos1... is the creator address of the denom (alice)
tokend tx tokenfactory change-admin factory/cosmos1.../utest [bob-addr] --from alice

# Query the authority metadata of the factory/cosmos1.../utest denom
tokend q tokenfactory denom-authority-metadata factory/cosmos1.../utest
authority_metadata:
  admin: [bob-addr]
```
