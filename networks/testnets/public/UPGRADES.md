# Scheduled Upgrades 🗓️ 

## v8-Rho

### Schedule

| Date | Testnet plan                             |
|------|------------------------------------------|
| January 20 2023 | ✅ v8-Rho upgrade ([Gaia v8.0.0-rc3](https://github.com/cosmos/gaia/releases/tag/v8.0.0-rc3)) is live on the testnet       |
| January 18 2023 | ✅ Submit and pass v8-Rho software upgrade [proposal](https://explorer.theta-testnet.polypore.xyz/proposals/112) |

* **Version before upgrade**: `v7.1.0`
* **Version after upgrade**: `v8.0.0-rc3`

### Upgrade height and binaries

* Upgrade height: `14175595` 
* Upgrade time: `2023-01-20 14:36:02 UTC`
  * The upgrade took under a minute, and block `14175596` was indexed at `14:36:53 UTC`.
* Unsigned proposal:

```json=
{"body":{"messages":[{"@type":"/cosmos.gov.v1beta1.MsgSubmitProposal","content":{"@type":"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal","title":"v8-Rho","description":"# v8-Rho Software Upgrade\r\n\r\n## Summary\r\n\r\nThis on-chain upgrade governance proposal is to adopt Gaia `v8.0.0-rc3`.By voting YES to this proposal, you approve of adding these updates to the Cosmos Hub Public Testnet.\r\n\r\n## Details\r\n\r\nSince the last v7-Theta upgrade at height 9283650, there have been a number of updates, fixes and new modules added to Gaia.\r\n\r\nThis is a proposal to adopt release candidate 3 for the Gaia v8-Rho upgrade on the public testnet. It contains the following changes:\r\n\r\n* (gaia) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v3.4.0](https://github.com/cosmos/ibc-go/releases/tag/v3.4.0) to fix a vulnerability in ICA. See [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v3.4.0/CHANGELOG.md) for details.\r\n* (gaia) Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.11](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.11). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/release/v0.45.x/CHANGELOG.md) for details.\r\n* (gaia) Bump [tendermint](https://github.com/tendermint/tendermint) to [0.34.24](https://github.com/tendermint/tendermint/tree/v0.34.24). See [CHANGELOG.md](https://github.com/tendermint/tendermint/blob/v0.34.24/CHANGELOG.md) for details.\r\n* (gaia) Bump [liquidity](https://github.com/Gravity-Devs/liquidity) to [v1.5.3](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.3).\r\n* (gaia) Bump [packet-forwarding-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) to [v3.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v3.1.1).\r\n* (feat) Add [globalfee](https://github.com/cosmos/gaia/tree/main/x/globalfee) module. See [globalfee docs](https://github.com/cosmos/gaia/blob/main/docs/modules/globalfee.md) for more details.\r\n* (feat) [#1845](https://github.com/cosmos/gaia/pull/1845) Add bech32-convert command to gaiad.\r\n* (fix) [Add new fee decorator](https://github.com/cosmos/gaia/pull/1961) to change MaxBypassMinFeeMsgGasUsage so importers of x/globalfee can change MaxGas.\r\n* (fix) [#1870](https://github.com/cosmos/gaia/issues/1870) Fix bank denom metadata in migration. See #1892 for more details.\r\n* (fix) [#1976](https://github.com/cosmos/gaia/pull/1976) Fix Quicksilver ICA exploit in migration. See the bug fix forum post for more details.\r\n* (tests) Add [E2E](https://github.com/cosmos/gaia/tree/main/tests/e2e) tests. The tests cover transactions/queries tests of different modules, including Bank, Distribution, Encode, Evidence, FeeGrant, Global Fee, Gov, IBC, packet forwarding middleware, Slashing, Staking, and Vesting module.\r\n* (tests) [#1941](https://github.com/cosmos/gaia/pull/1941) Fix packet forward configuration for e2e tests.\r\n* (tests) Use gaiad to swap out [Ignite](https://github.com/ignite/cli) in [liveness tests](https://github.com/cosmos/gaia/blob/main/.github/workflows/test.yml).\r\n\r\n## On-Chain Upgrade Process\r\n\r\nWhen the network reaches the upgrade height, the state machine program will be halted. One method for upgrading requires all validators and node operators to manually substitute the existing state machine binary with the new binary. Alternatively, Cosmovisor has the ability to download the binaries automatically before swapping them.\r\n\r\nTo test a simulated local upgrade please see the local testnet documentation. Because it is an on-chain upgrade process, the blockchain will be continued with all the accumulated history with continuous block height.","plan":{"name":"v8-Rho","time":"0001-01-01T00:00:00Z","height":"14175595","info":"{\"binaries\":{\"linux/amd64\":\"https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-linux-amd64?checksum=sha256:52236137b101de47dc392ce831c7d379d7a0dca35cdde997f6de61241d6cc71e\",\"linux/arm64\":\"https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-linux-arm64?checksum=sha256:1d118a0f8911c5039e07e1f90cd4bcc85ac1d22c2686e6c838c9157aa6d89031\",\"darwin/amd64\":\"https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-darwin-amd64?checksum=sha256:77c6b73b43ad583484b5d0373c5176583654f32477e9b47db2370ac30e34875a\",\"darwin/arm64\":\"https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-darwin-arm64?checksum=sha256:c32c36fc49a05f38916863c7a429decaab534552449c36657e58bd80917770ff\",\"windows/amd64\":\"https://github.com/cosmos/gaia/releases/download/v8.0.0-rc3/gaiad-v8.0.0-rc3-windows-amd64.exe?checksum=sha256:ca1b0ded75093862850a1beaffd8fd15c96bea701701130077926724228c27f9\"}}","upgraded_client_state":null}},"initial_deposit":[{"denom":"uatom","amount":"1"}],"proposer":"cosmos10v6wvdenee8r9l6wlsphcgur2ltl8ztkvhc8fw"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"uatom","amount":"5000"}],"gas_limit":"1000000","payer":"","granter":""}},"signatures":[]}
```

## v7-Theta

### Schedule

| Date          | Testnet plan                                                                                                        |
|---------------|---------------------------------------------------------------------------------------------------------------------|
| March 17 2022 | ✅  Theta upgrade ([Gaia v7.0.0-rc0](https://github.com/cosmos/gaia/releases/tag/v7.0.0-rc0)) is live on the testnet |
| March 16 2022 | ✅  Voting ends                                                                                                      |
| March 16 2022 | ✅  Submit v7-Theta software [upgrade proposal](https://testnet.cosmos.bigdipper.live/proposals/66)                  |
| March 10 2022 | ✅  Launch testnet chain with Gaia v6.0.0 (previous version)                                                         |


* **Version before upgrade**: `v6.0.x`
* **Version after upgrade**: `v7.0.0-rc0`

The v7-Theta upgrade was successfully completed on **March 17th, 2022 at 16:14 UTC (12:14 PM ET)**. The upgrade halt height was `9283650`, and blocks were being produced 7 minutes later.

Relevant log lines:
```
Mar 17 12:07:40 earth cosmovisor[822]: 12:07PM ERR UPGRADE "v7-Theta" NEEDED at height: 9283650
Mar 17 12:14:42 earth cosmovisor[13563]: 12:14PM INF finalizing commit of block hash=D83798E929BA7FB1F740C7E583EC2918EC40EDD3249B14BC72876130053BD0EE height=9283651 module=consensus num_txs=0 root=17F5C1754B53350A543A6BB29DE5E35A9DB9874AF89117220117213E53E38344
```

### Validators participating in upgrade testing

* 0base.vc
* 20MB Lab
* Cosmic Validator - Testnet
* Everstake
* Itachi
* KalpaTech
* P2P.ORG Validator
* Provalidator
* StakeWithUs
* Stakely.io
* WeStaking

Thank you to all participating validators! These validators received `THETA` tokens to their self-delegation addresses as part of our [collectables program](https://interchain-io.medium.com/cosmos-hub-theta-testnet-token-collectables-d08967ba2875)!

### Upgrade height and binaries

* Upgrade height: `9283650` 
* Estimated update time: 16:00 UTC
* On-chain proposal:

```bash=
gaiad tx gov submit-proposal software-upgrade v7-Theta \
--title v7-Theta \
--deposit 500000uatom \
--upgrade-height 9283650 \
--upgrade-info '{"binaries": {"linux/amd64": "https://github.com/cosmos/gaia/releases/download/v7.0.0-rc0/gaiad-v7.0.0-rc0-linux-amd64?checksum=sha256:4e95eaca51d6e0ab61b7a759aafc4b4674c270b8ffa764cb953d3808a1f9e264","linux/arm64": "https://github.com/cosmos/gaia/releases/download/v7.0.0-rc0/gaiad-v7.0.0-rc0-linux-arm64?checksum=sha256:574916076c6e0960fa980522ed9a404967a6f4c306448d09649a11e5626cd991","darwin/amd64": "https://github.com/cosmos/gaia/releases/download/v7.0.0-rc0/gaiad-v7.0.0-rc0-darwin-amd64?checksum=sha256:547182dd4456e8d71ff5256484458f0690a865d5c9f2185286dd9ab71dd11b10","windows/amd64": "https://github.com/cosmos/gaia/releases/download/v7.0.0-rc0/gaiad-v7.0.0-rc0-windows-amd64.exe?checksum=sha256:4eea1a32af3ed79632cfc8cca7088a10b3d89f767310e3c24fe31ad99492f6c8"}}' \
--description "This on-chain upgrade governance proposal is to adopt gaia v7.0.0 which includes a number of updates, fixes and new modules. By voting YES to this proposal, you approve of adding these updates to the Cosmos Hub.\n\n# Background\n\nSince the last v6-Vega upgrade at height 86950000 there have been a number of updates, fixes and new modules added to the Cosmos SDK, IBC and Tendermint.\n\nThis is a proposal to adopt the first release candiate for the [v7-Theta](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md#v7-theta-upgrade-expected-q1-2022) upgrade on the public testnet.\n\nIt contains the following changes:\n\n* (gaia) bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1). See [CHANGELOG.md](https://github.com/cosmos/cosmos-sdk/blob/v0.45.1/CHANGELOG.md#v0451---2022-02-03) for details.\n* (gaia) bump [ibc-go](https://github.com/cosmos/ibc-go) module to [v3.0.0](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0). See [CHANGELOG.md](https://github.com/cosmos/ibc-go/blob/v3.0.0/CHANGELOG.md#v300---2022-03-15) for details.\n* (gaia) add [interhcian account](https://github.com/cosmos/ibc-go/tree/main/modules/apps/27-interchain-accounts) module (interhchain-account module is part of ibc-go module).\n* (gaia) bump [liquidity](https://github.com/gravity-devs/liquidity/x/liquidity) module to [v1.5.0](https://github.com/Gravity-Devs/liquidity/releases/tag/v1.5.0). See [CHANGELOG.md](https://github.com/Gravity-Devs/liquidity/blob/v1.5.0/CHANGELOG.md#v150---20220223) for details.\n* (gaia) bump [packet-forward-middleware](https://github.com/strangelove-ventures/packet-forward-middleware) module to [v2.1.1](https://github.com/strangelove-ventures/packet-forward-middleware/releases/tag/v2.1.1).\n* (gaia) add migration logs for upgrade process.\n\n# On-Chain Upgrade Process\nWhen the network reaches the halt height, the state machine program of the Cosmos Hub will be halted. The classic method for upgrading requires all validators and node operators to manually substitute the existing state machine binary with the new binary. There is also a newer method that relies on Cosmovisor to swap the binaries automatically. Cosmovisor also includes the ability to download the binaries automatically before swapping them. To test a simulated local upgrade please see the local testnet documentation. Because it is an onchain upgrade process, the blockchain will be continued with all the accumulated history with continuous block height." \
--fees 1500uatom \
--gas auto \
--from <key_name> \
--chain-id theta-testnet-001 \
--node tcp://localhost:26657 \
--yes
```