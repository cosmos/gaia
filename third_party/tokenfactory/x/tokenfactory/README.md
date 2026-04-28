# Token Factory

The tokenfactory module allows any account to create a new token with the name `factory/{creator address}/{subdenom}`. Because tokens are namespaced by creator address, this allows token minting to be permissionless, due to not needing to resolve name collisions. A single account can create multiple denoms, by providing a unique subdenom for each created denom. Once a denom is created, the original creator is given "admin" privileges over the asset. This allows them to:
* Mint their denom to any account
* Burn their denom from any account
* Create a transfer of their denom between any two accounts
* Change the admin.

## Messages

### MsgCreateDenom

The `MsgCreateDenom` message allows an account to create a new denom. It requires a sender address and a sub denomination. The (sender_address, sub_denomination) tuple must be unique and cannot be re-used.

The resulting denom created is defined as `factory/{creator address}/{subdenom}`. The denom's admin is originally set to be the creator, but this can be changed later.

Subdenoms can contain `[a-zA-Z0-9./]`.

```protobuf
message MsgCreateDenom {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/create-denom";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  string subdenom = 2 [ (gogoproto.moretags) = "yaml:\"subdenom\"" ];
}
```

This message returns a response containing the full string of the newly created denom:

```protobuf
message MsgCreateDenomResponse {
  string new_token_denom = 1
      [ (gogoproto.moretags) = "yaml:\"new_token_denom\"" ];
}
```

#### State Modifications

* Fund community pool with the denom creation fee from the creator address, set in `Params`.
* Set `DenomMetaData` via bank keeper.
* Set `AuthorityMetadata` for the given denom to store the admin for the created denom `factory/{creator address}/{subdenom}`. Admin is automatically set as the Msg sender.
* Add denom to the `CreatorPrefixStore`, where a state of denoms created per creator is kept.

This message is expected to fail if:

* The sender address is invalid
* The subdenom is invalid or exceeds 44 alphanumeric characters
* The (sender, subdenom) tuple already exists

When this message is processed the following actions occur:

* A new denom is created with the format `factory/{sender}/{subdenom}`
* The sender is set as the admin of the newly created denom
* The denom creation fee is charged to the sender
* The denom creation gas is consumed

### MsgMint

The `MsgMint` message allows an admin account to mint more of a token. The minted tokens can be sent to the sender's account or to a specified address.

```protobuf
message MsgMint {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/mint";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false,
    (amino.encoding)     = "legacy_coin"
  ];
  string mintToAddress = 3 [
    (gogoproto.moretags) = "yaml:\"mint_to_address\"",
    (amino.dont_omitempty) = true
  ];
}
```

#### State Modifications

* Safety check the following
  * Check that the denom minting is created via `tokenfactory` module
  * Check that the sender of the message is the admin of the denom
* Mint designated amount of tokens for the denom via `bank` module

This message is expected to fail if:

* The sender is not the admin of the denom
* The sender address is invalid
* The mint to address (if provided) is invalid
* The amount is invalid or zero

When this message is processed the following actions occur:

* The specified amount of tokens is minted
* The minted tokens are sent to the `mintToAddress` if specified, otherwise to the sender

### MsgBurn

The `MsgBurn` message allows an admin account to burn tokens. The burned tokens can be from the sender's account or from a specified address.

```protobuf
message MsgBurn {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/burn";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false,
    (amino.encoding)     = "legacy_coin"
  ];
  string burnFromAddress = 3 [
    (gogoproto.moretags) = "yaml:\"burn_from_address\"",
    (amino.dont_omitempty) = true
  ];
}
```

#### State Modifications

* Safety check the following
  * Check that the denom minting is created via `tokenfactory` module
  * Check that the sender of the message is the admin of the denom
* Burn designated amount of tokens for the denom via `bank` module

This message is expected to fail if:

* The sender is not the admin of the denom
* The sender address is invalid
* The burn from address (if provided) is invalid
* The amount is invalid or zero
* The account being burned from has insufficient balance

When this message is processed the following actions occur:

* The specified amount of tokens is burned from the `burnFromAddress` if specified, otherwise from the sender


### MsgChangeAdmin

The `MsgChangeAdmin` message allows an admin account to reassign adminship of a denom to a new account.

```protobuf
message MsgChangeAdmin {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/change-admin";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  string denom = 2 [ (gogoproto.moretags) = "yaml:\"denom\"" ];
  string new_admin = 3 [ (gogoproto.moretags) = "yaml:\"new_admin\"" ];
}
```

#### State Modifications

* Check that sender of the message is the admin of denom
* Modify `AuthorityMetadata` state entry to change the admin of the denom

This message is expected to fail if:

* The sender is not the current admin of the denom
* The sender address is invalid
* The new admin address is invalid
* The denom is invalid or does not exist

When this message is processed the following actions occur:

* The admin of the specified denom is changed to the `new_admin` address

### MsgSetDenomMetadata

The `MsgSetDenomMetadata` message allows an admin account to set the denom's bank metadata.

```protobuf
message MsgSetDenomMetadata {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/set-denom-metadata";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.bank.v1beta1.Metadata metadata = 2 [
    (gogoproto.moretags) = "yaml:\"metadata\"",
    (gogoproto.nullable) = false
  ];
}
```

This message is expected to fail if:

* The sender is not the admin of the denom
* The sender address is invalid
* The metadata is invalid
* The base denom in the metadata is invalid

When this message is processed the following actions occur:

* The bank metadata for the specified denom is set or updated

### MsgForceTransfer

The `MsgForceTransfer` message allows an admin account to transfer tokens from one account to another.

```protobuf
// MsgForceTransfer allows an admin to transfer tokens from one account to another
message MsgForceTransfer {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name) = "osmosis/tokenfactory/force-transfer";

  string sender = 1 [ (gogoproto.moretags) = "yaml:\"sender\"" ];
  cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.moretags) = "yaml:\"amount\"",
    (gogoproto.nullable) = false,
    (amino.encoding)     = "legacy_coin"
  ];
  string transferFromAddress = 3
      [ (gogoproto.moretags) = "yaml:\"transfer_from_address\"" ];
  string transferToAddress = 4
      [ (gogoproto.moretags) = "yaml:\"transfer_to_address\"" ];
}
```

This message is expected to fail if:

* The sender is not the admin of the denom
* The sender address is invalid
* The transfer from address is invalid
* The transfer to address is invalid
* The amount is invalid
* The account being transferred from has insufficient balance

When this message is processed the following actions occur:

* The specified amount of tokens is transferred from `transferFromAddress` to `transferToAddress`

### MsgUpdateParams

The `MsgUpdateParams` message updates the tokenfactory module parameters.
The params are updated through a governance proposal where the signer is the gov module account address.

```protobuf
// MsgUpdateParams is the Msg/UpdateParams request type.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name) = "osmosis/tokenfactory/update-params";

  // authority is the address of the governance account.
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params defines the x/tokenfactory parameters to update.
  //
  // NOTE: All parameters must be supplied.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

The message handling can fail if:

* Signer is not the authority defined in the tokenfactory keeper (usually the gov module account)

## State

### Params

The tokenfactory module stores its params in state with the prefix of `0x00`,
they can be updated via governance.

* Params: `0x00 | ProtocolBuffer(Params)`

```protobuf
message Params {
  repeated cosmos.base.v1beta1.Coin denom_creation_fee = 1 [
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (gogoproto.moretags) = "yaml:\"denom_creation_fee\"",
    (gogoproto.nullable) = false
  ];

  uint64 denom_creation_gas_consume = 2 [
    (gogoproto.moretags) = "yaml:\"denom_creation_gas_consume\"",
    (gogoproto.nullable) = true
  ];
}
```

### DenomAuthorityMetadata

DenomAuthorityMetadata stores the admin address for each denom created via the tokenfactory module. The admin has permissions to mint, burn, force transfer, and change the admin of the denom.

* DenomAuthorityMetadata: `denoms|{denom}|authoritymetadata -> ProtocolBuffer(DenomAuthorityMetadata)`

```protobuf
message DenomAuthorityMetadata {
  option (gogoproto.equal) = true;
  // Can be empty for no admin, or a valid address
  string admin = 1 [ (gogoproto.moretags) = "yaml:\"admin\"" ];
}
```

### Denoms by Creator

The tokenfactory module maintains an index of all denoms created by each creator address. This allows for efficient querying of all denoms created by a specific account.

* Creator to Denoms: `creator|{creatorAddress}|{denom} -> []byte{}`

This index is used to:
- Query all denoms created by a specific address
- Validate denom ownership and permissions
- Support efficient lookups in queries and transactions


## Events

The tokenfactory module emits the following events:

### Msg's

### MsgCreateDenom

| Type         | Attribute Key   | Attribute Value  |
| ------------ | --------------- | ---------------- |
| create_denom | creator         | {creatorAddress} |
| create_denom | new_token_denom | {newTokenDenom}  |
| message      | module          | tokenfactory     |
| message      | action          | create_denom     |
| message      | sender          | {senderAddress}  |

### MsgMint

| Type    | Attribute Key   | Attribute Value |
| ------- | --------------- | --------------- |
| tf_mint | mint_to_address | {mintToAddress} |
| tf_mint | amount          | {amount}        |
| message | module          | tokenfactory    |
| message | action          | tf_mint         |
| message | sender          | {senderAddress} |

### MsgBurn

| Type    | Attribute Key     | Attribute Value   |
| ------- | ----------------- | ----------------- |
| tf_burn | burn_from_address | {burnFromAddress} |
| tf_burn | amount            | {amount}          |
| message | module            | tokenfactory      |
| message | action            | tf_burn           |
| message | sender            | {senderAddress}   |

### MsgForceTransfer

| Type           | Attribute Key         | Attribute Value       |
| -------------- | --------------------- | --------------------- |
| force_transfer | transfer_from_address | {transferFromAddress} |
| force_transfer | transfer_to_address   | {transferToAddress}   |
| force_transfer | amount                | {amount}              |
| message        | module                | tokenfactory          |
| message        | action                | force_transfer        |
| message        | sender                | {senderAddress}       |

### MsgChangeAdmin

| Type         | Attribute Key | Attribute Value   |
| ------------ | ------------- | ----------------- |
| change_admin | denom         | {denom}           |
| change_admin | new_admin     | {newAdminAddress} |
| message      | module        | tokenfactory      |
| message      | action        | change_admin      |
| message      | sender        | {senderAddress}   |

### MsgSetDenomMetadata

| Type               | Attribute Key  | Attribute Value    |
| ------------------ | -------------- | ------------------ |
| set_denom_metadata | denom          | {denom}            |
| set_denom_metadata | denom_metadata | {metadata}         |
| message            | module         | tokenfactory       |
| message            | action         | set_denom_metadata |
| message            | sender         | {senderAddress}    |


## Parameters

The liquid module contains the following parameters:

| Key                     | Type           | Example                                  |
| ----------------------- | -------------- | ---------------------------------------- |
| DenomCreationFee        | SDK coin array | `[{"denom":"token","amount":"1000000"}]` |
| DenomCreationGasConsume | string         | `"100000"`                               |

## Client

### CLI

A user can query and interact with the `tokenfactory` module using the CLI.

#### Query

The `query` commands allows users to query `tokenfactory` state.

```bash
tokend query tokenfactory --help
```

##### params

The `params` command allows users to query the current module params.

Usage:

```bash
tokend query tokenfactory params [flags]
```

Example:

```bash
tokend query tokenfactory params
```

Example Output:

```bash
params:
  denom_creation_fee:
  - amount: "10000000"
    denom: utoken
  denom_creation_gas_consume: "2000000"
```

##### denom-authority-metadata

The `denom-authority-metadata` command allows users to query the authority metadata for a specific denom.

Usage:

```bash
tokend query tokenfactory denom-authority-metadata [denom] [flags]
```

Example:

```bash
tokend query tokenfactory denom-authority-metadata factory/cosmos1...addr.../subdenom
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
tokend query tokenfactory denoms-from-creator [creator-address] [flags]
```

Example:

```bash
tokend query tokenfactory denoms-from-creator cosmos1...addr...
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
tokend query tokenfactory denoms-from-admin [admin-address] [flags]
```

Example:

```bash
tokend query tokenfactory denoms-from-admin cosmos1...addr...
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
tokend tx tokenfactory --help
```

##### create-denom

The command `create-denom` allows users to create a new denom.

Usage:

```bash
tokend tx tokenfactory create-denom [subdenom] [flags]
```

Example:

```bash
tokend tx tokenfactory create-denom mytoken --from=mykey
```

##### mint

The command `mint` allows denom admins to mint tokens to their address.

Usage:

```bash
tokend tx tokenfactory mint [amount] [flags]
```

Example:

```bash
tokend tx tokenfactory mint 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### mint-to

The command `mint-to` allows denom admins to mint tokens to a specific address.

Usage:

```bash
tokend tx tokenfactory mint-to [address] [amount] [flags]
```

Example:

```bash
tokend tx tokenfactory mint-to cosmos1...recipient... 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### burn

The command `burn` allows denom admins to burn tokens from their address.

Usage:

```bash
tokend tx tokenfactory burn [amount] [flags]
```

Example:

```bash
tokend tx tokenfactory burn 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### burn-from

The command `burn-from` allows denom admins to burn tokens from a specific address.

Usage:

```bash
tokend tx tokenfactory burn-from [address] [amount] [flags]
```

Example:

```bash
tokend tx tokenfactory burn-from cosmos1...addr... 1000factory/cosmos1...addr.../mytoken --from=mykey
```

##### force-transfer

The command `force-transfer` allows denom admins to transfer tokens from one address to another.

Usage:

```bash
tokend tx tokenfactory force-transfer [amount] [transfer-from-address] [transfer-to-address] [flags]
```

Example:

```bash
tokend tx tokenfactory force-transfer 1000factory/cosmos1...addr.../mytoken cosmos1...from... cosmos1...to... --from=mykey
```

##### change-admin

The command `change-admin` allows denom admins to change the admin of a denom.
* The admin address can be set to the gov module account.

Usage:

```bash
tokend tx tokenfactory change-admin [denom] [new-admin-address] [flags]
```

Example:

```bash
tokend tx tokenfactory change-admin factory/cosmos1...addr.../mytoken cosmos1...newadmin... --from=mykey
```

##### modify-metadata

The command `modify-metadata` allows denom admins to modify the metadata of a denom.

Usage:

```bash
tokend tx tokenfactory modify-metadata [denom] [ticker-symbol] [description] [exponent] [flags]
```

Example:

```bash
tokend tx tokenfactory modify-metadata factory/cosmos1...addr.../mytoken MYTOKEN "My Token Description" 6 --from=mykey
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
        "denom": "utoken",
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

## Expectations from the chain

The chain's bech32 prefix for addresses can be at most 16 characters long.

This comes from denoms having a 128 byte maximum length, enforced from the SDK,
and us setting longest_subdenom to be 44 bytes.

A token factory token's denom is: `factory/{creator address}/{subdenom}`

Splitting up into sub-components, this has:

- `len(factory) = 7`
- `2 * len("/") = 2`
- `len(longest_subdenom)`
- `len(creator_address) = len(bech32(longest_addr_length, chain_addr_prefix))`.

Longest addr length at the moment is `32 bytes`. Due to SDK error correction
settings, this means `len(bech32(32, chain_addr_prefix)) = len(chain_addr_prefix) + 1 + 58`.
Adding this all, we have a total length constraint of `128 = 7 + 2 + len(longest_subdenom) + len(longest_chain_addr_prefix) + 1 + 58`.
Therefore `len(longest_subdenom) + len(longest_chain_addr_prefix) = 128 - (7 + 2 + 1 + 58) = 60`.

The choice between how we standardized the split these 60 bytes between maxes
from longest_subdenom and longest_chain_addr_prefix is somewhat arbitrary.
Considerations going into this:

- Per [BIP-0173](https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki#bech32)
  the technically longest HRP for a 32 byte address ('data field') is 31 bytes.
  (Comes from encode(data) = 59 bytes, and max length = 90 bytes)
- subdenom should be at least 32 bytes so hashes can go into it
- longer subdenoms are very helpful for creating human readable denoms
- chain addresses should prefer being smaller. The longest HRP in cosmos to date is 11 bytes. (`persistence`)

For explicitness, the limits are set to `len(longest_subdenom) = 44` and `len(longest_chain_addr_prefix) = 16`.

If the Cosmos SDK increases the maximum length of a denom from 128 bytes,
these caps should increase.