# `auth` subspace

The `auth` module is responsible for authenticating accounts and transactions. It has the following parameters:

<table>
    <tr>
        <th>Key</th>
        <th>Value</th>
    </tr>
    <tr v-for="(v,k) in $themeConfig.currentParameters.auth">
        <td><a :href="'#'+k"><code>{{ k }}</code></a></td>
        <td><code>{{ v }}</code></td>
    </tr>
</table>

The `auth` module is responsible for specifying the base transaction and account types for an application, since the SDK itself is agnostic to these particulars. It contains the ante handler, where all basic transaction validity checks (signatures, nonces, auxiliary fields) are performed, and exposes the account keeper, which allows other modules to read, write, and modify accounts.

## Governance notes on parameters

### `MaxMemoCharacters`
**The character limit for each transaction memo.**

There is an option to include a "memo," or additional information (data) to Cosmos Hub transactions, whether sending funds, delegating, voting, or other transaction types. This parameter limits the number of characters that may be included in the memo line of each transaction.

* on-chain value: `{{ $themeConfig.currentParameters.auth.MaxMemoCharacters }}`
* `cosmoshub-4` genesis: `512`
* `cosmoshub-3` genesis: `512`

#### Decreasing the value of `MaxMemoCharacters`
Decreasing the value of `MaxMemoCharacters` will decrease the character limit for each transaction memo. This may break the functionality of applications that rely upon the data in the memo field. For example, an exchange may use a common deposit address for all of its users, and then individualize account deposits using the memo field. If the memo field suddenly decreased, the exchange may no longer automatically sort its users' transactions.

#### Increasing the value of `MaxMemoCharacters`
Increasing the value of `MaxMemoCharacters` will increase the character limit for each transaction memo. This may enable new functionality for applications that use transaction memos. It may also enable an increase in the amount of data in each block, leading to an increased storage need for the blockchain and [state bloat](https://thecontrol.co/state-growth-a-look-at-the-problem-and-its-solutions-6de9d7634b0b).

### `TxSigLimit`
**The max number of signatures per transaction**

Users and applications may create multisignature (aka multisig) accounts. These accounts require more than one signature to generate a transaction. This parameter limits the number of signatures in a transaction.

* on-chain value: `{{ $themeConfig.currentParameters.auth.TxSigLimit }}`
* `cosmoshub-4` genesis: `7`
* `cosmoshub-3` genesis: `7`

#### Decreasing the value of `TxSigLimit`
Decreasing the value of `TxSigLimit` will decrease the maximum number of signatures possible. This may constrain stakeholders that want to use as many as seven signatures to authorize a transaction. It will also break the functionality of entities or applications dependent upon up to seven transactions, meaning that those transactions will no longer be able to be authorized. In this case, funds and functions controlled by a multisignature address will no longer be accessible, and funds may become stranded.

#### Increasing the value of `TxSigLimit`
Increasing the value of `TxSigLimit` will increase the maximum number of signatures possible. As this value increases, the network becomes more likely to be susceptible to attacks that slow block production, due to the burden of computational cost when verifying more signatures (since signature verification is costlier than other operations).

### `TxSizeCostPerByte`
**Sets the cost of transactions, in units of gas.**

`TxSizeCostPerByte` is used to compute the gas-unit consumption for each transaction.

* on-chain value: `{{ $themeConfig.currentParameters.auth.TxSizeCostPerByte }}`
* `cosmoshub-4` genesis: `10`
* `cosmoshub-3` genesis: `10`

#### Decreasing the value of `TxSizeCostPerByte`
Decreasing the value of `TxSizeCostPerByte` will reduce the number of gas units used per transaction. This may also reduce the fees that validators earn for processing transactions. There may be other effects that have not been detailed here.

#### Increasing the value of `TxSizeCostPerByte`
Increasing the value of `TxSizeCostPerByte` will raise the number of gas units used per transaction. This may also increase the fees that validators earn for processing transactions. There may be other effects that have not been detailed here.

### `SigVerifyCostED25519`
**The cost for verifying ED25519 signatures, in units of gas.**

Ed25519 is the EdDSA cryptographic signature scheme (using SHA-512 (SHA-2) and Curve25519) that is used by Cosmos Hub validators. `SigVerifyCostED25519` is the gas (ie. computational) cost for verifying ED25519 signatures, and ED25519-based transactions are not currently accepted by the Cosmos Hub.

* on-chain value: `{{ $themeConfig.currentParameters.auth.SigVerifyCostED25519 }}`
* `cosmoshub-4` genesis: `590`
* `cosmoshub-3` genesis: `590`

#### Decreasing the value of `SigVerifyCostED25519`
Decreasing the value of `SigVerifyCostED25519` will decrease the amount of gas that is consumed by transactions that require Ed25519 signature verifications. Since Ed25519-signed transactions are not currently accepted by Cosmos Hub, changing this parameter is unlikely to affect stakeholders at this time.

#### Increasing the value of `SigVerifyCostED25519`
Increasing the value of `SigVerifyCostED25519` will increase the amount of gas that is consumed by transactions that require Ed25519 signature verifications. Since Ed25519 signature transactions are not currently accepted by Cosmos Hub, changing this parameter is unlikely to affect stakeholders at this time.

#### Notes
Ed25519 signatures are not currently being accepted by the Cosmos Hub. Ed25519 signatures will be verified and can be considered valid, so the gas to verify them will be consumed. However, the transaction itself will be rejected. It could be that these signatures will be used for transactions a later time, such as after inter-blockchain communication (IBC) evidence upgrades happen.

### `SigVerifyCostSecp256k1`
**The cost for verifying Secp256k1 signatures, in units of gas.**

Secp256k1 is an elliptic curve domain parameter for cryptographic signatures used by user accounts in the Cosmos Hub. `SigVerifyCostSecp256k1` is the gas (ie. computational) cost for verifying Secp256k1 signatures. Practically all Cosmos Hub transactions require Secp256k1 signature verifications.

* on-chain value: `{{ $themeConfig.currentParameters.auth.SigVerifyCostSecp256k1 }}`
* `cosmoshub-4` default: `1000`
* `cosmoshub-3` default: `1000`

#### Decreasing the value of `SigVerifyCostSecp256k1`
Decreasing the value of `SigVerifyCostSecp256k1` will decrease the amount of gas that is consumed by practically all Cosmos Hub transactions, which require Secp256k1 signature verifications. Decreasing this parameter may have unintended effects on how the Cosmos Hub operates, since the computational cost of verifying a transaction may be greater than what the system's assumption is.

#### Increasing the value of `SigVerifyCostSecp256k1`
Increasing the value of `SigVerifyCostSecp256k1` will increase the amount of gas that is consumed by practically all Cosmos Hub transactions, which require Secp256k1 signature verifications. Increasing this parameter may have unintended effects on how the Cosmos Hub operates, since the computational cost of verifying a transaction may be less than what the system's assumption is.


#### Notes
There should be a better understanding of what the potential implications are for changing `SigVerifyCostSecp256k1`. For example, gas calculations are important because blocks have a gas limit. Transactions could be rejected for exceeding the block gas limit, breaking application functionality or perhaps preventing addresses controlled by multiple signatures from moving funds.