package ethereum_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/ethereum"
)

func TestGetEthAddressFromStdout(t *testing.T) {
	exampleOutput := `Compiling 72 files with Solc 0.8.25
Solc 0.8.25 finished in 2.43s
Compiler run successful!
Traces:
  [13132696] E2ETestDeploy::run()
    ├─ [0] VM::projectRoot() [staticcall]
    │   └─ ← [Return] "/Users/gg/Code/solidity-ibc-eureka"
    ├─ [0] VM::readFile("/Users/gg/Code/solidity-ibc-eureka/e2e/genesis.json") [staticcall]
    │   └─ ← [Return] <file>
    ├─ [0] VM::parseJsonBytes("<stringified JSON>", ".trustedClientState") [staticcall]
    │   └─ ← [Return] 0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000013000000000000000000000000000000000000000000000000000000000012754500000000000000000000000000000000000000000000000000000000001baf800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000673696d642d310000000000000000000000000000000000000000000000000000
    ├─ [0] VM::parseJsonBytes("<stringified JSON>", ".trustedConsensusState") [staticcall]
    │   └─ ← [Return] 0x000000000000000000000000000000000000000000000000000000006708142974c0681a0dbf3e340cedb56abacacec3f01106cd13d1d240e6eb7fc60a34c31250d828ba2c80c290ee3edb5d12ec232dee758f20214af6fbd73a1f636f79d4d7
    ├─ [0] VM::parseJsonBytes32("<stringified JSON>", ".updateClientVkey") [staticcall]
    │   └─ ← [Return] 0x00787fed71dce3ac5685c23cb727f0bbe41c9c40977506f33530f52eabab6c86
    ├─ [0] VM::parseJsonBytes32("<stringified JSON>", ".membershipVkey") [staticcall]
    │   └─ ← [Return] 0x00a8f49c50bcef3ffaef3e8018be59281904d5951d6e20e0a824b9504da18c7f
    ├─ [0] VM::parseJsonBytes32("<stringified JSON>", ".ucAndMembershipVkey") [staticcall]
    │   └─ ← [Return] 0x0037bf750186b666717a245012144eb63d1c8675134c9a9c81133535c4f76656
    ├─ [0] VM::envUint("PRIVATE_KEY") [staticcall]
    │   └─ ← [Return] <env var value>
    ├─ [0] VM::startBroadcast(<pk>)
    │   └─ ← [Return]
    ├─ [2404781] → new SP1Verifier@0x867EBEE8fB04ef90a4161fe21b89420B0aeEF8f2
    │   └─ ← [Return] 12011 bytes of code
    ├─ [2442427] → new SP1ICS07Tendermint@0x65cE09e5864dD1f45F4ae50396A307291AaD6631
    │   └─ ← [Return] 11624 bytes of code
    ├─ [1171142] → new ICS02Client@0xEBC7C68E032d765e392CFf5B2a11E76C2C43BbbF
    │   ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: 0x51A4283eBaeC10B9B764AE8B021BcAA30C0631Ff)
    │   └─ ← [Return] 5730 bytes of code
    ├─ [2953720] → new ICS26Router@0xfcf4c2FAc206cFABE9C2B68AefE5D0a9fA038501
    │   ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: 0x51A4283eBaeC10B9B764AE8B021BcAA30C0631Ff)
    │   └─ ← [Return] 14411 bytes of code
    ├─ [3299657] → new ICS20Transfer@0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22
    │   ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: ICS26Router: [0xfcf4c2FAc206cFABE9C2B68AefE5D0a9fA038501])
    │   └─ ← [Return] 16250 bytes of code
    ├─ [531853] → new TestERC20@0x022b667cC0D57836CCb12669ce93Ae1e15d4f8BC
    │   └─ ← [Return] 2431 bytes of code
    ├─ [28083] ICS26Router::addIBCApp("transfer", ICS20Transfer: [0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22])
    │   ├─ emit IBCAppAdded(portId: "transfer", app: ICS20Transfer: [0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22])
    │   └─ ← [Stop]
    ├─ [0] VM::stopBroadcast()
    │   └─ ← [Return]
    ├─ [0] VM::serializeString("<stringified JSON>", "ics07Tendermint", "0x65ce09e5864dd1f45f4ae50396a307291aad6631")
    │   └─ ← [Return] "{\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\"}"
    ├─ [0] VM::serializeString("<stringified JSON>", "ics02Client", "0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf")
    │   └─ ← [Return] "{\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\"}"
    ├─ [0] VM::serializeString("<stringified JSON>", "ics26Router", "0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501")
    │   └─ ← [Return] "{\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\",\"ics26Router\":\"0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501\"}"
    ├─ [0] VM::serializeString("<stringified JSON>", "ics20Transfer", "0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22")
    │   └─ ← [Return] "{\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\",\"ics20Transfer\":\"0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22\",\"ics26Router\":\"0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501\"}"
    ├─ [0] VM::serializeString("<stringified JSON>", "erc20", "0x022b667cc0d57836ccb12669ce93ae1e15d4f8bc")
    │   └─ ← [Return] "{\"erc20\":\"0x022b667cc0d57836ccb12669ce93ae1e15d4f8bc\",\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\",\"ics20Transfer\":\"0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22\",\"ics26Router\":\"0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501\"}"
    └─ ← [Return] "{\"erc20\":\"0x022b667cc0d57836ccb12669ce93ae1e15d4f8bc\",\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\",\"ics20Transfer\":\"0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22\",\"ics26Router\":\"0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501\"}"


Script ran successfully.

== Return ==
0: string "{\"erc20\":\"0x022b667cc0d57836ccb12669ce93ae1e15d4f8bc\",\"ics02Client\":\"0xebc7c68e032d765e392cff5b2a11e76c2c43bbbf\",\"ics07Tendermint\":\"0x65ce09e5864dd1f45f4ae50396a307291aad6631\",\"ics20Transfer\":\"0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22\",\"ics26Router\":\"0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501\"}"

## Setting up 1 EVM.
==========================
Simulated On-chain Traces:

  [521379] → new ICS20Lib@0xDcC1FffeC88e1Fa830604047fad1Aa8D957A7DC7
    └─ ← [Return] 2604 bytes of code

  [2404781] → new SP1Verifier@0x867EBEE8fB04ef90a4161fe21b89420B0aeEF8f2
    └─ ← [Return] 12011 bytes of code

  [2442427] → new SP1ICS07Tendermint@0x65cE09e5864dD1f45F4ae50396A307291AaD6631
    └─ ← [Return] 11624 bytes of code

  [1171142] → new ICS02Client@0xEBC7C68E032d765e392CFf5B2a11E76C2C43BbbF
    ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: 0x51A4283eBaeC10B9B764AE8B021BcAA30C0631Ff)
    └─ ← [Return] 5730 bytes of code

  [2953720] → new ICS26Router@0xfcf4c2FAc206cFABE9C2B68AefE5D0a9fA038501
    ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: 0x51A4283eBaeC10B9B764AE8B021BcAA30C0631Ff)
    └─ ← [Return] 14411 bytes of code

  [3299657] → new ICS20Transfer@0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22
    ├─ emit OwnershipTransferred(previousOwner: 0x0000000000000000000000000000000000000000, newOwner: ICS26Router: [0xfcf4c2FAc206cFABE9C2B68AefE5D0a9fA038501])
    └─ ← [Return] 16250 bytes of code

  [531853] → new TestERC20@0x022b667cC0D57836CCb12669ce93Ae1e15d4f8BC
    └─ ← [Return] 2431 bytes of code

  [30083] ICS26Router::addIBCApp("transfer", ICS20Transfer: [0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22])
    ├─ emit IBCAppAdded(portId: "transfer", app: ICS20Transfer: [0xD6D4C57D09bA13C9535Ee2d6BdB100231d793a22])
    └─ ← [Stop]


==========================

Chain 3151908

Estimated gas price: 0.460566435 gwei

Estimated total gas used for script: 19196301

Estimated amount required: 0.008841171916756935 ETH

==========================


==========================

ONCHAIN EXECUTION COMPLETE & SUCCESSFUL.

Transactions saved to: /Users/gg/Code/solidity-ibc-eureka/broadcast/E2ETestDeploy.s.sol/3151908/run-latest.json

Sensitive values saved to: /Users/gg/Code/solidity-ibc-eureka/cache/E2ETestDeploy.s.sol/3151908/run-latest.json
`

	deployedContracts, err := ethereum.GetEthContractsFromDeployOutput(exampleOutput)
	require.NoError(t, err)

	require.Equal(t, "0x65ce09e5864dd1f45f4ae50396a307291aad6631", deployedContracts.Ics07Tendermint)
	require.Equal(t, "0xfcf4c2fac206cfabe9c2b68aefe5d0a9fa038501", deployedContracts.Ics26Router)
	require.Equal(t, "0xd6d4c57d09ba13c9535ee2d6bdb100231d793a22", deployedContracts.Ics20Transfer)
	require.Equal(t, "0x022b667cc0d57836ccb12669ce93ae1e15d4f8bc", deployedContracts.Erc20)
}
