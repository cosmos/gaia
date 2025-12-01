---
title: tokenfactory
order: 2
---

The `tokenfactory` module used by the Hub allows any account to create a new token with the name `factory/{creator address}/{subdenom}`.

Gaia uses a fork of Strangelove's implementaton of the tokenfactory module: [cosmos/tokenfactory](https://github.com/cosmos/tokenfactory). The fork introduces the following changes with respect to the Strangelove implementation:
- The sudo mint capability was removed.

You can find more details in the [module documentation](https://github.com/cosmos/tokenfactory/blob/main/x/tokenfactory/README.md).

## Client

### CLI

A user can query and interact with the `tokenfactory` module using the CLI.

#### Query

The `query` commands allows users to query `tokenfactory` state.

```bash
gaiad query tokenfactory --help
```

##### params

The `params` command allows users to query the current module params.

Usage:

```bash
gaiad query tokenfactory params [flags]
```

Example:

```bash
gaiad query tokenfactory params
```

Example Output:

```bash
params:
  denom_creation_fee:
  - amount: "10000000"
    denom: uatom
  denom_creation_gas_consume: "2000000"
```

##### denom-authority-metadata

The `denom-authority-metadata` command allows users to query the authority metadata for a specific denom.

Usage:

```bash
gaiad query tokenfactory denom-authority-metadata [denom] [flags]
```

Example:

```bash
gaiad query tokenfactory denom-authority-metadata factory/cosmos1...addr.../subdenom
```

Example Output:

```bash
authority_metadata:
  admin: cosmos1...addr...
```

##### denoms-from-creator

The `denoms-from-creator` command allows users to query all denoms created by a specific creator address.

Usage:

```bash
gaiad query tokenfactory denoms-from-creator [creator-address] [flags]
```

Example:

```bash
gaiad query tokenfactory denoms-from-creator cosmos1...addr...
```

Example Output:

```bash
denoms:
- factory/cosmos1...addr.../subdenom1
- factory/cosmos1...addr.../subdenom2
```

##### denoms-from-admin

The `denoms-from-admin` command allows users to query all denoms owned by a specific admin address.

Usage:

```bash
gaiad query tokenfactory denoms-from-admin [admin-address] [flags]
```

Example:

```bash
gaiad query tokenfactory denoms-from-admin cosmos1...addr...
```

Example Output:

```bash
denoms:
- factory/cosmos1...addr.../subdenom1
- factory/cosmos1...addr.../subdenom2
```

#### Transactions

The `tx` commands allows users to interact with the `tokenfactory` module.

```bash
gaiad tx tokenfactory --help
```

##### create-denom

The command `create-denom` allows users to create a new denom.

Usage:

```bash
gaiad tx tokenfactory create-denom [subdenom] [flags]
```

Example:

```bash
gaiad tx tokenfactory create-denom mytoken --from=mykey
```

##### mint

The command `mint` allows denom admins to mint tokens to their address.

Usage:

```bash
gaiad tx tokenfactory mint [amount] [flags]
```

Example:

```bash
gaiad tx tokenfactory mint 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### mint-to

The command `mint-to` allows denom admins to mint tokens to a specific address.

Usage:

```bash
gaiad tx tokenfactory mint-to [address] [amount] [flags]
```

Example:

```bash
gaiad tx tokenfactory mint-to cosmos1...recipient... 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### burn

The command `burn` allows denom admins to burn tokens from their address.

Usage:

```bash
gaiad tx tokenfactory burn [amount] [flags]
```

Example:

```bash
gaiad tx tokenfactory burn 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### burn-from

The command `burn-from` allows denom admins to burn tokens from a specific address.

Usage:

```bash
gaiad tx tokenfactory burn-from [address] [amount] [flags]
```

Example:

```bash
gaiad tx tokenfactory burn-from cosmos1...addr... 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### force-transfer

The command `force-transfer` allows denom admins to transfer tokens from one address to another.

Usage:

```bash
gaiad tx tokenfactory force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]
```

Example:

```bash
gaiad tx tokenfactory force-transfer 1000factory/cosmos1...addr.../mytoken cosmos1...from... cosmos1...to... --from=mykey
```

##### change-admin

The command `change-admin` allows denom admins to change the admin of a denom.

* The admin address can be set to the gov module account.
* The admin address can be set to an empty string to renounce admin control entirely.

Usage:

```bash
gaiad tx tokenfactory change-admin [denom] [new-admin-address] [flags]
```

Example:

```bash
gaiad tx tokenfactory change-admin factory/cosmos1...addr.../mytoken cosmos1...newadmin... --from=mykey
```

##### modify-metadata

The command `modify-metadata` allows denom admins to modify the bank metadata of a denom.

Usage:

```bash
gaiad tx tokenfactory modify-metadata [denom] [ticker-symbol] [description] [exponent] [flags]
```

Example:

```bash
gaiad tx tokenfactory modify-metadata factory/cosmos1...addr.../mytoken MYTOKEN "My Token Description" 6 --from=mykey
```

### gRPC

A user can query the `tokenfactory` module using gRPC endpoints.

#### Params

The `Params` endpoint queries the module parameters.

```bash
osmosis.tokenfactory.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext localhost:9090 osmosis.tokenfactory.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "denomCreationFee": [
      {
        "denom": "uatom",
        "amount": "10000000"
      }
    ],
    "denomCreationGasConsume": "2000000"
  }
}
```

#### DenomAuthorityMetadata

The `DenomAuthorityMetadata` endpoint queries the authority metadata for a specific denom.

```bash
osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata
```

Example:

```bash
grpcurl -plaintext -d '{"denom": "factory/cosmos1...addr.../mytoken"}' \
localhost:9090 osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata
```

Example Output:

```bash
{
  "authorityMetadata": {
    "admin": "cosmos1...addr..."
  }
}
```

#### DenomsFromCreator

The `DenomsFromCreator` endpoint queries all denoms created by a specific creator address.

```bash
osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator
```

Example:

```bash
grpcurl -plaintext -d '{"creator": "cosmos1...addr..."}' \
localhost:9090 osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator
```

Example Output:

```bash
{
  "denoms": [
    "factory/cosmos1...addr.../subdenom1",
    "factory/cosmos1...addr.../subdenom2"
  ]
}
```

#### DenomsFromAdmin

The `DenomsFromAdmin` endpoint queries all denoms owned by a specific admin address.

```bash
osmosis.tokenfactory.v1beta1.Query/DenomsFromAdmin
```

Example:

```bash
grpcurl -plaintext -d '{"admin": "cosmos1...addr..."}' \
localhost:9090 osmosis.tokenfactory.v1beta1.Query/DenomsFromAdmin
```

Example Output:

```bash
{
  "denoms": [
    "factory/cosmos1...addr.../subdenom1",
    "factory/cosmos1...addr.../subdenom2"
  ]
}
```

### REST

A user can query the `tokenfactory` module using REST endpoints.

#### Params

The `Params` REST endpoint queries the module parameters.

```bash
/osmosis/tokenfactory/v1beta1/params
```

Example:

```bash
curl -X GET "http://localhost:1317/osmosis/tokenfactory/v1beta1/params" -H "accept: application/json"
```

Example Output:

```json
{
  "params": {
    "denom_creation_fee": [
      {
        "denom": "uatom",
        "amount": "10000000"
      }
    ],
    "denom_creation_gas_consume": "2000000"
  }
}
```

#### DenomAuthorityMetadata

The `DenomAuthorityMetadata` REST endpoint queries the authority metadata for a specific denom.

```bash
/osmosis/tokenfactory/v1beta1/denoms/{denom}/authority_metadata
```

Example:

```bash
curl -X GET "http://localhost:1317/osmosis/tokenfactory/v1beta1/denoms/factory%2Fcosmos1...addr...%2Fmytoken/authority_metadata" -H "accept: application/json"
```

Example Output:

```json
{
  "authority_metadata": {
    "admin": "cosmos1...addr..."
  }
}
```

#### DenomsFromCreator

The `DenomsFromCreator` REST endpoint queries all denoms created by a specific creator address.

```bash
/osmosis/tokenfactory/v1beta1/denoms_from_creator/{creator}
```

Example:

```bash
curl -X GET "http://localhost:1317/osmosis/tokenfactory/v1beta1/denoms_from_creator/cosmos1...addr..." -H "accept: application/json"
```

Example Output:

```json
{
  "denoms": [
    "factory/cosmos1...addr.../subdenom1",
    "factory/cosmos1...addr.../subdenom2"
  ]
}
```

#### DenomsFromAdmin

The `DenomsFromAdmin` REST endpoint queries all denoms owned by a specific admin address.

```bash
/osmosis/tokenfactory/v1beta1/denoms_from_admin/{admin}
```

Example:

```bash
curl -X GET "http://localhost:1317/osmosis/tokenfactory/v1beta1/denoms_from_admin/cosmos1...addr..." -H "accept: application/json"
```

Example Output:

```json
{
  "denoms": [
    "factory/cosmos1...addr.../subdenom1",
    "factory/cosmos1...addr.../subdenom2"
  ]
}
```

## Parameters

The `tokenfactory` module uses the following parameters, both of which can be updated via governance:
* Denom Creation Fee
  * Param: `denom_creation_fee`
  * This is the fee required to create one denom.
* Denom Creation Gas Consume
  * Param: `denom_creation_gas_consume`
  * This is the gas that will be consumed each time a denom is created.

The JSON below can be used as a reference proposal to update the tokenfactory module params.
```json
{
  "messages": [
    {
      "@type": "/osmosis.tokenfactory.v1beta1.MsgUpdateParams",
      "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
      "params": {
        "denom_creation_fee": [
          {
            "denom": "uatom",
            "amount": "50000000"
          }
        ],
        "denom_creation_gas_consume": "5000000"
      }
    }
  ],
  "metadata": "ipfs://CID",
  "deposit": "50000000uatom",
  "title": "Update tokenfactory params",
  "summary": "Set the tokenfactory params to 50ATOM denom creation fee and 5_000_000 denom creation gas consume.",
  "expedited": false
}
```

## Denoms

Each denom has an `authority metadata`, which lists the denom admin.
* The admin is allowed to submit mint, burn force transfer, and change admin transactions.
* If a denom admin is set to the governance module, transactions can only be processed through a governance proposal.
* If a denom admin is set to an empty string, the denom becomes a fixed supply token with no further mint and burn operations allowed.
