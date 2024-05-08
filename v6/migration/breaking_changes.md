# Breaking Changes

This document collects all of the *breaking changes* from the CHANGELOG.md files
located in the [Tendermint](https://github.com/tendermint/tendermint), [Cosmos SDK](https://github.com/cosmos/cosmos-sdk),
and [Gaia](https://github.com/cosmos/gaia) Github repositories.

Its purpose is to provide a checklist for potential impact on deployments; however, the changelog located in each repository
serves other important details, such as bug fixes and feature improvements.

## Table of Contents

- [Tendermint Changelog](#tendermint-changelog); Tendermint versions v0.32.13 - v0.34.1
- [Cosmos Changelog](#cosmos-sdk-changelog); Cosmos SDK versions v0.37.14 - v0.40.0

# Tendermint Changelog

## v0.34.1


- CLI/RPC/Config
    - [cli] [\#5786](https://github.com/tendermint/tendermint/issues/5786) deprecate snake_case commands for hyphen-case (@cmwaters)

- Go API
    - [libs/protoio] [\#5868](https://github.com/tendermint/tendermint/issues/5868) Return number of bytes read in `Reader.ReadMsg()` (@erikgrinaker)

## v0.34.0


- CLI/RPC/Config

    - [config] [\#5315](https://github.com/tendermint/tendermint/pull/5315) Rename `prof_laddr` to `pprof_laddr` and move it to `rpc` section (@melekes)
    - [evidence] [\#4959](https://github.com/tendermint/tendermint/pull/4959) Add JSON tags to `DuplicateVoteEvidence` (@marbar3778)
    - [light] [\#4946](https://github.com/tendermint/tendermint/pull/4946) `tendermint lite` command has been renamed to `tendermint light` (@marbar3778)
    - [privval] [\#4582](https://github.com/tendermint/tendermint/pull/4582) `round` in private_validator_state.json is no longer JSON string; instead it is a number (@marbar3778)
    - [rpc] [\#4792](https://github.com/tendermint/tendermint/pull/4792) `/validators` are now sorted by voting power (@melekes)
    - [rpc] [\#4947](https://github.com/tendermint/tendermint/pull/4947) Return an error when `page` pagination param is 0 in `/validators`, `tx_search` (@melekes)
    - [rpc] [\#5137](https://github.com/tendermint/tendermint/pull/5137) JSON tags of `gasWanted` and `gasUsed` in `ResponseCheckTx` and `ResponseDeliverTx` have been made snake_case (`gas_wanted` and `gas_used`) (@marbar3778)
    - [rpc] [\#5315](https://github.com/tendermint/tendermint/pull/5315) Remove `/unsafe_start_cpu_profiler`, `/unsafe_stop_cpu_profiler` and `/unsafe_write_heap_profile`. Please use pprof functionality instead (@melekes)
    - [rpc/client, rpc/jsonrpc/client] [\#5347](https://github.com/tendermint/tendermint/pull/5347) All client methods now accept `context.Context` as 1st param (@melekes)

- Apps

    - [abci] [\#4704](https://github.com/tendermint/tendermint/pull/4704) Add ABCI methods `ListSnapshots`, `LoadSnapshotChunk`, `OfferSnapshot`, and `ApplySnapshotChunk` for state sync snapshots. `ABCIVersion` bumped to 0.17.0. (@erikgrinaker)
    - [abci] [\#4989](https://github.com/tendermint/tendermint/pull/4989) `Proof` within `ResponseQuery` has been renamed to `ProofOps`  (@marbar3778)
    - [abci] [\#5096](https://github.com/tendermint/tendermint/pull/5096) `CheckTxType` Protobuf enum names are now uppercase, to follow Protobuf style guide (@erikgrinaker)
    - [abci] [\#5324](https://github.com/tendermint/tendermint/pull/5324) ABCI evidence type is now an enum with two types of possible evidence (@cmwaters)

- P2P Protocol

    - [blockchain] [\#4637](https://github.com/tendermint/tendermint/pull/4637) Migrate blockchain reactor(s) to Protobuf encoding (@marbar3778)
    - [evidence] [\#4949](https://github.com/tendermint/tendermint/pull/4949) Migrate evidence reactor to Protobuf encoding (@marbar3778)
    - [mempool] [\#4940](https://github.com/tendermint/tendermint/pull/4940) Migrate mempool from to Protobuf encoding (@marbar3778)
    - [mempool] [\#5321](https://github.com/tendermint/tendermint/pull/5321) Batch transactions when broadcasting them to peers (@melekes)
        - `MaxBatchBytes` new config setting defines the max size of one batch.
    - [p2p/pex] [\#4973](https://github.com/tendermint/tendermint/pull/4973) Migrate `p2p/pex` reactor to Protobuf encoding (@marbar3778)
    - [statesync] [\#4943](https://github.com/tendermint/tendermint/pull/4943) Migrate state sync reactor to Protobuf encoding (@marbar3778)

- Blockchain Protocol

    - [evidence] [\#4725](https://github.com/tendermint/tendermint/pull/4725) Remove `Pubkey` from `DuplicateVoteEvidence` (@marbar3778)
    - [evidence] [\#5499](https://github.com/tendermint/tendermint/pull/5449) Cap evidence to a maximum number of bytes (supercedes [\#4780](https://github.com/tendermint/tendermint/pull/4780)) (@cmwaters)
    - [merkle] [\#5193](https://github.com/tendermint/tendermint/pull/5193) Header hashes are no longer empty for empty inputs, notably `DataHash`, `EvidenceHash`, and `LastResultsHash` (@erikgrinaker)
    - [state] [\#4845](https://github.com/tendermint/tendermint/pull/4845) Include `GasWanted` and `GasUsed` into `LastResultsHash` (@melekes)
    - [types] [\#4792](https://github.com/tendermint/tendermint/pull/4792) Sort validators by voting power to enable faster commit verification (@melekes)

- On-disk serialization

    - [state] [\#4679](https://github.com/tendermint/tendermint/pull/4679) Migrate state module to Protobuf encoding (@marbar3778)
        - `BlockStoreStateJSON` is now `BlockStoreState` and is encoded as binary in the database
    - [store] [\#4778](https://github.com/tendermint/tendermint/pull/4778) Migrate store module to Protobuf encoding (@marbar3778)

- Light client, private validator

    - [light] [\#4964](https://github.com/tendermint/tendermint/pull/4964) Migrate light module migration to Protobuf encoding (@marbar3778)
    - [privval] [\#4985](https://github.com/tendermint/tendermint/pull/4985) Migrate `privval` module to Protobuf encoding (@marbar3778)

- Go API

    - [consensus] [\#4582](https://github.com/tendermint/tendermint/pull/4582) RoundState: `Round`, `LockedRound` & `CommitRound` are now `int32` (@marbar3778)
    - [consensus] [\#4582](https://github.com/tendermint/tendermint/pull/4582) HeightVoteSet: `round` is now `int32` (@marbar3778)
    - [crypto] [\#4721](https://github.com/tendermint/tendermint/pull/4721) Remove `SimpleHashFromMap()` and `SimpleProofsFromMap()` (@erikgrinaker)
    - [crypto] [\#4940](https://github.com/tendermint/tendermint/pull/4940) All keys have become `[]byte` instead of `[<size>]byte`. The byte method no longer returns the marshaled value but just the `[]byte` form of the data. (@marbar3778)
    - [crypto] [\#4988](https://github.com/tendermint/tendermint/pull/4988) Removal of key type multisig (@marbar3778)
        - The key has been moved to the [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/types/multisig/multisignature.go)
    - [crypto] [\#4989](https://github.com/tendermint/tendermint/pull/4989) Remove `Simple` prefixes from `SimpleProof`, `SimpleValueOp` & `SimpleProofNode`. (@marbar3778)
        - `merkle.Proof` has been renamed to `ProofOps`.
        - Protobuf messages `Proof` & `ProofOp` has been moved to `proto/crypto/merkle`
        - `SimpleHashFromByteSlices` has been renamed to `HashFromByteSlices`
        - `SimpleHashFromByteSlicesIterative` has been renamed to `HashFromByteSlicesIterative`
        - `SimpleProofsFromByteSlices` has been renamed to `ProofsFromByteSlices`
    - [crypto] [\#4941](https://github.com/tendermint/tendermint/pull/4941) Remove suffixes from all keys. (@marbar3778)
        - ed25519: type `PrivKeyEd25519` is now `PrivKey`
        - ed25519: type `PubKeyEd25519` is now `PubKey`
        - secp256k1: type`PrivKeySecp256k1` is now `PrivKey`
        - secp256k1: type`PubKeySecp256k1` is now `PubKey`
        - sr25519: type `PrivKeySr25519` is now `PrivKey`
        - sr25519: type `PubKeySr25519` is now `PubKey`
    - [crypto] [\#5214](https://github.com/tendermint/tendermint/pull/5214) Change `GenPrivKeySecp256k1` to `GenPrivKeyFromSecret` to be consistent with other keys (@marbar3778)
    - [crypto] [\#5236](https://github.com/tendermint/tendermint/pull/5236) `VerifyBytes` is now `VerifySignature` on the `crypto.PubKey` interface (@marbar3778)
    - [evidence] [\#5361](https://github.com/tendermint/tendermint/pull/5361) Add LightClientAttackEvidence and change evidence interface (@cmwaters)
    - [libs] [\#4831](https://github.com/tendermint/tendermint/pull/4831) Remove `Bech32` pkg from Tendermint. This pkg now lives in the [cosmos-sdk](https://github.com/cosmos/cosmos-sdk/tree/4173ea5ebad906dd9b45325bed69b9c655504867/types/bech32) (@marbar3778)
    - [light] [\#4946](https://github.com/tendermint/tendermint/pull/4946) Rename `lite2` pkg to `light`. Remove `lite` implementation. (@marbar3778)
    - [light] [\#5347](https://github.com/tendermint/tendermint/pull/5347) `NewClient`, `NewHTTPClient`, `VerifyHeader` and `VerifyLightBlockAtHeight` now accept `context.Context` as 1st param (@melekes)
    - [merkle] [\#5193](https://github.com/tendermint/tendermint/pull/5193) `HashFromByteSlices` and `ProofsFromByteSlices` now return a hash for empty inputs, following RFC6962 (@erikgrinaker)
    - [proto] [\#5025](https://github.com/tendermint/tendermint/pull/5025) All proto files have been moved to `/proto` directory. (@marbar3778)
        - Using the recommended the file layout from buf, [see here for more info](https://buf.build/docs/lint-checkers#file_layout)
    - [rpc/client] [\#4947](https://github.com/tendermint/tendermint/pull/4947) `Validators`, `TxSearch` `page`/`per_page` params become pointers (@melekes)
        - `UnconfirmedTxs` `limit` param is a pointer
    - [rpc/jsonrpc/server] [\#5141](https://github.com/tendermint/tendermint/pull/5141) Remove `WriteRPCResponseArrayHTTP` (use `WriteRPCResponseHTTP` instead) (@melekes)
    - [state] [\#4679](https://github.com/tendermint/tendermint/pull/4679) `TxResult` is a Protobuf type defined in `abci` types directory (@marbar3778)
    - [state] [\#5191](https://github.com/tendermint/tendermint/pull/5191) Add `State.InitialHeight` field to record initial block height, must be `1` (not `0`) to start from 1 (@erikgrinaker)
    - [state] [\#5231](https://github.com/tendermint/tendermint/pull/5231) `LoadStateFromDBOrGenesisFile()` and `LoadStateFromDBOrGenesisDoc()` no longer saves the state in the database if not found, the genesis state is simply returned (@erikgrinaker)
    - [state] [\#5348](https://github.com/tendermint/tendermint/pull/5348) Define an Interface for the state store. (@marbar3778)
    - [types] [\#4939](https://github.com/tendermint/tendermint/pull/4939)  `SignedMsgType` has moved to a Protobuf enum types (@marbar3778)
    - [types] [\#4962](https://github.com/tendermint/tendermint/pull/4962) `ConsensusParams`, `BlockParams`, `EvidenceParams`, `ValidatorParams` & `HashedParams` are now Protobuf types (@marbar3778)
    - [types] [\#4852](https://github.com/tendermint/tendermint/pull/4852) Vote & Proposal `SignBytes` is now func `VoteSignBytes` & `ProposalSignBytes` (@marbar3778)
    - [types] [\#4798](https://github.com/tendermint/tendermint/pull/4798) Simplify `VerifyCommitTrusting` func + remove extra validation (@melekes)
    - [types] [\#4845](https://github.com/tendermint/tendermint/pull/4845) Remove `ABCIResult` (@melekes)
    - [types] [\#5029](https://github.com/tendermint/tendermint/pull/5029) Rename all values from `PartsHeader` to `PartSetHeader` to have consistency (@marbar3778)
    - [types] [\#4939](https://github.com/tendermint/tendermint/pull/4939) `Total` in `Parts` & `PartSetHeader` has been changed from a `int` to a `uint32` (@marbar3778)
    - [types] [\#4939](https://github.com/tendermint/tendermint/pull/4939) Vote: `ValidatorIndex` & `Round` are now `int32` (@marbar3778)
    - [types] [\#4939](https://github.com/tendermint/tendermint/pull/4939) Proposal: `POLRound` & `Round` are now `int32` (@marbar3778)
    - [types] [\#4939](https://github.com/tendermint/tendermint/pull/4939) Block: `Round` is now `int32` (@marbar3778)

## v0.33.8

## v0.33.7

## v0.33.6

**Please note that the fix for the False Witness issue renames the `VerifyCommitTrusting`
function to `VerifyCommitLightTrusting`. If you were relying on the light client, you may
need to update your code.**

## v0.33.5

### BREAKING CHANGES:

- Go API

    - [privval] [\#4744](https://github.com/tendermint/tendermint/pull/4744) Remove deprecated `OldFilePV` (@melekes)
    - [mempool] [\#4759](https://github.com/tendermint/tendermint/pull/4759) Modify `Mempool#InitWAL` to return an error (@melekes)
    - [node] [\#4832](https://github.com/tendermint/tendermint/pull/4832) `ConfigureRPC` returns an error (@melekes)
    - [rpc] [\#4836](https://github.com/tendermint/tendermint/pull/4836) Overhaul `lib` folder (@melekes)
      Move lib/ folder to jsonrpc/.
      Rename:
      rpc package -> jsonrpc package
      rpcclient package -> client package
      rpcserver package -> server package
      JSONRPCClient to Client
      JSONRPCRequestBatch to RequestBatch
      JSONRPCCaller to Caller
      StartHTTPServer to Serve
      StartHTTPAndTLSServer to ServeTLS
      NewURIClient to NewURI
      NewJSONRPCClient to New
      NewJSONRPCClientWithHTTPClient to NewWithHTTPClient
      NewWSClient to NewWS
      Unexpose ResponseWriterWrapper
      Remove unused http_params.go

## v0.33.4

- Nodes are no longer guaranteed to contain all blocks up to the latest height. The ABCI app can now control which blocks to retain through the ABCI field `ResponseCommit.retain_height`, all blocks and associated data below this height will be removed.

### BREAKING CHANGES:

- Go API

    - [lite2] [\#4616](https://github.com/tendermint/tendermint/pull/4616) Make `maxClockDrift` an option `Verify/VerifyAdjacent/VerifyNonAdjacent` now accept `maxClockDrift time.Duration` (@melekes).
    - [rpc/client] [\#4628](https://github.com/tendermint/tendermint/pull/4628) Split out HTTP and local clients into `http` and `local` packages (@erikgrinaker).

## v0.33.3

## v0.33.2

### BREAKING CHANGES:

- CLI/RPC/Config
    - [cli] [\#4505](https://github.com/tendermint/tendermint/pull/4505) `tendermint lite` sub-command new syntax (@melekes):
      `lite cosmoshub-3 -p 52.57.29.196:26657 -w public-seed-node.cosmoshub.certus.one:26657
      --height 962118 --hash 28B97BE9F6DE51AC69F70E0B7BFD7E5C9CD1A595B7DC31AFF27C50D4948`

- Go API
    - [lite2] [\#4535](https://github.com/tendermint/tendermint/pull/4535) Remove `Start/Stop` (@melekes)
    - [lite2] [\#4469](https://github.com/tendermint/tendermint/issues/4469) Remove `RemoveNoLongerTrustedHeaders` and `RemoveNoLongerTrustedHeadersPeriod` option (@cmwaters)
    - [lite2] [\#4473](https://github.com/tendermint/tendermint/issues/4473) Return height as a 2nd param in `TrustedValidatorSet` (@melekes)
    - [lite2] [\#4536](https://github.com/tendermint/tendermint/pull/4536) `Update` returns a signed header (1st param) (@melekes)


## v0.33.1

## v0.33

This release contains breaking changes to the `Block#Header`, specifically
`NumTxs` and `TotalTxs` were removed (\#2521). Here's how this change affects
different modules:

- apps: it breaks the ABCI header field numbering
- state: it breaks the format of `State` on disk
- RPC: all RPC requests which expose the header broke
- Go API: the `Header` broke
- P2P: since blocks go over the wire, technically the P2P protocol broke

### BREAKING CHANGES:

- CLI/RPC/Config

    - [rpc] [\#3471](https://github.com/tendermint/tendermint/issues/3471) Paginate `/validators` response (default: 30 vals per page)
    - [rpc] [\#3188](https://github.com/tendermint/tendermint/issues/3188) Remove `BlockMeta` in `ResultBlock` in favor of `BlockId` for `/block`
    - [rpc] `/block_results` response format updated (see RPC docs for details)
      ```
      {
        "jsonrpc": "2.0",
        "id": "",
        "result": {
          "height": "2109",
          "txs_results": null,
          "begin_block_events": null,
          "end_block_events": null,
          "validator_updates": null,
          "consensus_param_updates": null
        }
      }
      ```
    - [rpc] [\#4141](https://github.com/tendermint/tendermint/pull/4141) Remove `#event` suffix from the ID in event responses.
      `{"jsonrpc": "2.0", "id": 0, "result": ...}`
    - [rpc] [\#4141](https://github.com/tendermint/tendermint/pull/4141) Switch to integer IDs instead of `json-client-XYZ`
      ```
      id=0 method=/subscribe
      id=0 result=...
      id=1 method=/abci_query
      id=1 result=...
      ```
        - ID is unique for each request;
        - Request.ID is now optional. Notification is a Request without an ID. Previously ID="" or ID=0 were considered as notifications.

    - [config] [\#4046](https://github.com/tendermint/tendermint/issues/4046) Rename tag(s) to CompositeKey & places where tag is still present it was renamed to event or events. Find how a compositeKey is constructed [here](https://github.com/tendermint/tendermint/blob/6d05c531f7efef6f0619155cf10ae8557dd7832f/docs/app-dev/indexing-transactions.md)
        - You will have to generate a new config for your Tendermint node(s)
    - [genesis] [\#2565](https://github.com/tendermint/tendermint/issues/2565) Add `consensus_params.evidence.max_age_duration`. Rename
      `consensus_params.evidence.max_age` to `max_age_num_blocks`.
    - [cli] [\#1771](https://github.com/tendermint/tendermint/issues/1771) `tendermint lite` now uses new light client package (`lite2`)
      and has 3 more flags: `--trusting-period`, `--trusted-height` and
      `--trusted-hash`

- Apps

    - [tm-bench] Removed tm-bench in favor of [tm-load-test](https://github.com/informalsystems/tm-load-test)

- Go API

    - [rpc] [\#3953](https://github.com/tendermint/tendermint/issues/3953) Modify NewHTTP, NewXXXClient functions to return an error on invalid remote instead of panicking (@mrekucci)
    - [rpc/client] [\#3471](https://github.com/tendermint/tendermint/issues/3471) `Validators` now requires two more args: `page` and `perPage`
    - [libs/common] [\#3262](https://github.com/tendermint/tendermint/issues/3262) Make error the last parameter of `Task` (@PSalant726)
    - [cs/types] [\#3262](https://github.com/tendermint/tendermint/issues/3262) Rename `GotVoteFromUnwantedRoundError` to `ErrGotVoteFromUnwantedRound` (@PSalant726)
    - [libs/common] [\#3862](https://github.com/tendermint/tendermint/issues/3862) Remove `errors.go` from `libs/common`
    - [libs/common] [\#4230](https://github.com/tendermint/tendermint/issues/4230) Move `KV` out of common to its own pkg
    - [libs/common] [\#4230](https://github.com/tendermint/tendermint/issues/4230) Rename `cmn.KVPair(s)` to `kv.Pair(s)`s
    - [libs/common] [\#4232](https://github.com/tendermint/tendermint/issues/4232) Move `Service` & `BaseService` from `libs/common` to `libs/service`
    - [libs/common] [\#4232](https://github.com/tendermint/tendermint/issues/4232) Move `common/nil.go` to `types/utils.go` & make the functions private
    - [libs/common] [\#4231](https://github.com/tendermint/tendermint/issues/4231) Move random functions from `libs/common` into pkg `rand`
    - [libs/common] [\#4237](https://github.com/tendermint/tendermint/issues/4237) Move byte functions from `libs/common` into pkg `bytes`
    - [libs/common] [\#4237](https://github.com/tendermint/tendermint/issues/4237) Move throttletimer functions from `libs/common` into pkg `timer`
    - [libs/common] [\#4237](https://github.com/tendermint/tendermint/issues/4237) Move tempfile functions from `libs/common` into pkg `tempfile`
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move os functions from `libs/common` into pkg `os`
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move net functions from `libs/common` into pkg `net`
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move mathematical functions and types out of `libs/common` to `math` pkg
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move string functions out of `libs/common` to `strings` pkg
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move async functions out of `libs/common` to `async` pkg
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move bit functions out of `libs/common` to `bits` pkg
    - [libs/common] [\#4240](https://github.com/tendermint/tendermint/issues/4240) Move cmap functions out of `libs/common` to `cmap` pkg
    - [libs/common] [\#4258](https://github.com/tendermint/tendermint/issues/4258) Remove `Rand` from all `rand` pkg functions
    - [types] [\#2565](https://github.com/tendermint/tendermint/issues/2565) Remove `MockBadEvidence` & `MockGoodEvidence` in favor of `MockEvidence`

- Blockchain Protocol

    - [abci] [\#2521](https://github.com/tendermint/tendermint/issues/2521) Remove `TotalTxs` and `NumTxs` from `Header`
    - [types] [\#4151](https://github.com/tendermint/tendermint/pull/4151) Enforce ordering of votes in DuplicateVoteEvidence to be lexicographically sorted on BlockID
    - [types] [\#1648](https://github.com/tendermint/tendermint/issues/1648) Change `Commit` to consist of just signatures

- P2P Protocol

    - [p2p] [\#3668](https://github.com/tendermint/tendermint/pull/3668) Make `SecretConnection` non-malleable

- [proto] [\#3986](https://github.com/tendermint/tendermint/pull/3986) Prefix protobuf types to avoid name conflicts.
    - ABCI becomes `tendermint.abci.types` with the new API endpoint `/tendermint.abci.types.ABCIApplication/`
    - core_grpc becomes `tendermint.rpc.grpc` with the new API endpoint `/tendermint.rpc.grpc.BroadcastAPI/`
    - merkle becomes `tendermint.crypto.merkle`
    - libs.common becomes `tendermint.libs.common`
    - proto3 becomes `tendermint.types.proto3`

## v0.32.13

# Cosmos SDK Changelog

## v0.41.0 - From release notes

Support Amino JSON for IBC MsgTransfer
This change breaks state backward compatibility.

At the moment hardware wallets are unable to sign messages using SIGN_MODE_DIRECT because the cosmos ledger app does not support proto encoding andSIGN_MODE_TEXTUAL is not available yet.

In order to enable hardware wallets users to interact with IBC, amino JSON support was added to MsgTransfer only.

Counterparty.ChannelID not available in OnChanOpenAck callback implementation.
This change breaks state backward compatibility.

In a previous version the Counterparty.ChannelID was available for an OnChanOpenAck callback implementation (read via channelKeeper.GetChannel(). Due to a regression, the channelID is currently empty.

The issue has been fixed by reordering IBC ChanOpenAck and ChanOpenConfirm to execute the core handlers logic first, followed by application callbacks.

It breaks state backward compatibility because the current change consumes more gas, which means that in an updated node a TX might fail because it ran out of gas whilst in older versions it would be successful.

## [v0.40.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.0) - 2021-01-08

v0.40.0, known as the Stargate release of the Cosmos SDK, is one of the largest releases
of the Cosmos SDK since launch. Please read through this changelog and [release notes](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0/RELEASE_NOTES.md) to make
sure you are aware of any relevant breaking changes.

### Client Breaking Changes

* __CLI__
    * (client/keys) [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) remove `keys update` command.
    * (x/auth) [\#5844](https://github.com/cosmos/cosmos-sdk/pull/5844) `tx sign` command now returns an error when signing is attempted with offline/multisig keys.
    * (x/auth) [\#6108](https://github.com/cosmos/cosmos-sdk/pull/6108) `tx sign` command's `--validate-signatures` flag is migrated into a `tx validate-signatures` standalone command.
    * (x/auth) [#7788](https://github.com/cosmos/cosmos-sdk/pull/7788) Remove `tx auth` subcommands, all auth subcommands exist as `tx <subcommand>`
    * (x/genutil) [\#6651](https://github.com/cosmos/cosmos-sdk/pull/6651) The `gentx` command has been improved. No longer are `--from` and `--name` flags required. Instead, a single argument, `name`, is required which refers to the key pair in the Keyring. In addition, an optional
      `--moniker` flag can be provided to override the moniker found in `config.toml`.
    * (x/upgrade) [#7697](https://github.com/cosmos/cosmos-sdk/pull/7697) Rename flag name "--time" to "--upgrade-time", "--info" to "--upgrade-info", to keep it consistent with help message.
* __REST / Queriers__
    * (api) [\#6426](https://github.com/cosmos/cosmos-sdk/pull/6426) The ability to start an out-of-process API REST server has now been removed. Instead, the API server is now started in-process along with the application and Tendermint. Configuration options have been added to `app.toml` to enable/disable the API server along with additional HTTP server options.
    * (client) [\#7246](https://github.com/cosmos/cosmos-sdk/pull/7246) The rest server endpoint `/swagger-ui/` is replaced by `/swagger/`, and contains swagger documentation for gRPC Gateway routes in addition to legacy REST routes. Swagger API is exposed only if set in `app.toml`.
    * (x/auth) [\#5702](https://github.com/cosmos/cosmos-sdk/pull/5702) The `x/auth` querier route has changed from `"acc"` to `"auth"`.
    * (x/bank) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) The `/bank/balances/{address}` endpoint now returns all account balances or a single balance by denom when the `denom` query parameter is present.
    * (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove CLI and REST handlers for querying `x/evidence` parameters.
    * (x/gov) [#6295](https://github.com/cosmos/cosmos-sdk/pull/6295) Fix typo in querying governance params.
* __General__
    * (baseapp) [\#6384](https://github.com/cosmos/cosmos-sdk/pull/6384) The `Result.Data` is now a Protocol Buffer encoded binary blob of type `TxData`. The `TxData` contains `Data` which contains a list of Protocol Buffer encoded message data and the corresponding message type.
    * (client) [\#5783](https://github.com/cosmos/cosmos-sdk/issues/5783) Unify all coins representations on JSON client requests for governance proposals.
    * (crypto) [\#7419](https://github.com/cosmos/cosmos-sdk/pull/7419) The SDK doesn't use Tendermint's `crypto.PubKey`
      interface anymore, and uses instead it's own `PubKey` interface, defined in `crypto/types`. Replace all instances of
      `crypto.PubKey` by `cryptotypes.Pubkey`.
    * (store/rootmulti) [\#6390](https://github.com/cosmos/cosmos-sdk/pull/6390) Proofs of empty stores are no longer supported.
    * (store/types) [\#5730](https://github.com/cosmos/cosmos-sdk/pull/5730) store.types.Cp() is removed in favour of types.CopyBytes().
    * (x/auth) [\#6054](https://github.com/cosmos/cosmos-sdk/pull/6054) Remove custom JSON marshaling for base accounts as multsigs cannot be bech32 decoded.
    * (x/auth/vesting) [\#6859](https://github.com/cosmos/cosmos-sdk/pull/6859) Custom JSON marshaling of vesting accounts was removed. Vesting accounts are now marshaled using their default proto or amino JSON representation.
    * (x/bank) [\#5785](https://github.com/cosmos/cosmos-sdk/issues/5785) In x/bank errors, JSON strings coerced to valid UTF-8 bytes at JSON marshalling time
      are now replaced by human-readable expressions. This change can potentially break compatibility with all those client side tools
      that parse log messages.
    * (x/evidence) [\#7538](https://github.com/cosmos/cosmos-sdk/pull/7538) The ABCI's `Result.Data` field for
      `MsgSubmitEvidence` responses does not contain the raw evidence's hash, but the protobuf encoded
      `MsgSubmitEvidenceResponse` struct.
    * (x/gov) [\#7533](https://github.com/cosmos/cosmos-sdk/pull/7533) The ABCI's `Result.Data` field for
      `MsgSubmitProposal` responses does not contain a raw binary encoding of the `proposalID`, but the protobuf encoded
      `MsgSubmitSubmitProposalResponse` struct.
    * (x/gov) [\#6859](https://github.com/cosmos/cosmos-sdk/pull/6859) `ProposalStatus` and `VoteOption` are now JSON serialized using its protobuf name, so expect names like `PROPOSAL_STATUS_DEPOSIT_PERIOD` as opposed to `DepositPeriod`.
    * (x/staking) [\#7499](https://github.com/cosmos/cosmos-sdk/pull/7499) `BondStatus` is now a protobuf `enum` instead
      of an `int32`, and JSON serialized using its protobuf name, so expect names like `BOND_STATUS_UNBONDING` as opposed
      to `Unbonding`.
    * (x/staking) [\#7556](https://github.com/cosmos/cosmos-sdk/pull/7556) The ABCI's `Result.Data` field for
      `MsgBeginRedelegate` and `MsgUndelegate` responses does not contain custom binary marshaled `completionTime`, but the
      protobuf encoded `MsgBeginRedelegateResponse` and `MsgUndelegateResponse` structs respectively

### API Breaking Changes

* __Baseapp / Client__
    * (AppModule) [\#7518](https://github.com/cosmos/cosmos-sdk/pull/7518) [\#7584](https://github.com/cosmos/cosmos-sdk/pull/7584) Rename `AppModule.RegisterQueryServices` to `AppModule.RegisterServices`, as this method now registers multiple services (the gRPC query service and the protobuf Msg service). A `Configurator` struct is used to hold the different services.
    * (baseapp) [\#5865](https://github.com/cosmos/cosmos-sdk/pull/5865) The `SimulationResponse` returned from tx simulation is now JSON encoded instead of Amino binary.
    * (client) [\#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `CLIContext` is renamed to `Context`. `Context` and all related methods have been moved from package context to client.
    * (client) [\#6525](https://github.com/cosmos/cosmos-sdk/pull/6525) Removed support for `indent` in JSON responses. Clients should consider piping to an external tool such as `jq`.
    * (client) [\#8107](https://github.com/cosmos/cosmos-sdk/pull/8107) Renamed `PrintOutput` and `PrintOutputLegacy`
      methods of the `context.Client` object to `PrintProto` and `PrintObjectLegacy`.
    * (client/flags) [\#6632](https://github.com/cosmos/cosmos-sdk/pull/6632) Remove NewCompletionCmd(), the function is now available in tendermint.
    * (client/input) [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) Removal of unnecessary `GetCheckPassword`, `PrintPrefixed` functions.
    * (client/keys) [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) Rename `NewKeyBaseFromDir()` -> `NewLegacyKeyBaseFromDir()`.
    * (client/keys) [\#5820](https://github.com/cosmos/cosmos-sdk/pull/5820/) Removed method CloseDB from Keybase interface.
    * (client/rpc) [\#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `client` package and subdirs reorganization.
    * (client/lcd) [\#6290](https://github.com/cosmos/cosmos-sdk/pull/6290) `CliCtx` of struct `RestServer` in package client/lcd has been renamed to `ClientCtx`.
    * (codec) [\#6330](https://github.com/cosmos/cosmos-sdk/pull/6330) `codec.RegisterCrypto` has been moved to the `crypto/codec` package and the global `codec.Cdc` Amino instance has been deprecated and moved to the `codec/legacy_global` package.
    * (codec) [\#8080](https://github.com/cosmos/cosmos-sdk/pull/8080) Updated the `codec.Marshaler` interface
        * Moved `MarshalAny` and `UnmarshalAny` helper functions to `codec.Marshaler` and renamed to `MarshalInterface` and
          `UnmarshalInterface` respectively. These functions must take interface as a parameter (not a concrete type nor `Any`
          object). Underneath they use `Any` wrapping for correct protobuf serialization.
    * (crypto) [\#6780](https://github.com/cosmos/cosmos-sdk/issues/6780) Move ledger code to its own package.
    * (crypto/types/multisig) [\#6373](https://github.com/cosmos/cosmos-sdk/pull/6373) `multisig.Multisignature` has been renamed  to `AminoMultisignature`
    * (codec) `*codec.LegacyAmino` is now a wrapper around Amino which provides backwards compatibility with protobuf `Any`. ALL legacy code should use `*codec.LegacyAmino` instead of `*amino.Codec` directly
    * (crypto) [\#5880](https://github.com/cosmos/cosmos-sdk/pull/5880) Merge `crypto/keys/mintkey` into `crypto`.
    * (crypto/hd) [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `crypto/keys/hd` moved to `crypto/hd`.
    * (crypto/keyring):
        * [\#5866](https://github.com/cosmos/cosmos-sdk/pull/5866) Rename `crypto/keys/` to `crypto/keyring/`.
        * [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `Keybase` -> `Keyring` interfaces migration. `LegacyKeybase` interface is added in order
          to guarantee limited backward compatibility with the old Keybase interface for the sole purpose of migrating keys across the new keyring backends. `NewLegacy`
          constructor is provided [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) to allow for smooth migration of keys from the legacy LevelDB based implementation
          to new keyring backends. Plus, the package and the new keyring no longer depends on the sdk.Config singleton. Please consult the [package documentation](https://github.com/cosmos/cosmos-sdk/tree/master/crypto/keyring/doc.go) for more
          information on how to implement the new `Keyring` interface.
        * [\#5858](https://github.com/cosmos/cosmos-sdk/pull/5858) Make Keyring store keys by name and address's hexbytes representation.
    * (export) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) `AppExporter` now returns ABCI consensus parameters to be included in marshaled exported state. These parameters must be returned from the application via the `BaseApp`.
    * (simapp) Deprecating and renaming `MakeEncodingConfig` to `MakeTestEncodingConfig` (both in `simapp` and `simapp/params` packages).
    * (store) [\#5803](https://github.com/cosmos/cosmos-sdk/pull/5803) The `store.CommitMultiStore` interface now includes the new `snapshots.Snapshotter` interface as well.
    * (types) [\#5579](https://github.com/cosmos/cosmos-sdk/pull/5579) The `keepRecent` field has been removed from the `PruningOptions` type.
      The `PruningOptions` type now only includes fields `KeepEvery` and `SnapshotEvery`, where `KeepEvery`
      determines which committed heights are flushed to disk and `SnapshotEvery` determines which of these
      heights are kept after pruning. The `IsValid` method should be called whenever using these options. Methods
      `SnapshotVersion` and `FlushVersion` accept a version arugment and determine if the version should be
      flushed to disk or kept as a snapshot. Note, `KeepRecent` is automatically inferred from the options
      and provided directly the IAVL store.
    * (types) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Refactored `AppModuleBasic` and `AppModuleGenesis`
      to now accept a `codec.JSONMarshaler` for modular serialization of genesis state.
    * (types/rest) [\#5779](https://github.com/cosmos/cosmos-sdk/pull/5779) Drop unused Parse{Int64OrReturnBadRequest,QueryParamBool}() functions.
* __Modules__
    * (modules) [\#7243](https://github.com/cosmos/cosmos-sdk/pull/7243) Rename `RegisterCodec` to `RegisterLegacyAminoCodec` and `codec.New()` is now renamed to `codec.NewLegacyAmino()`
    * (modules) [\#6564](https://github.com/cosmos/cosmos-sdk/pull/6564) Constant `DefaultParamspace` is removed from all modules, use ModuleName instead.
    * (modules) [\#5989](https://github.com/cosmos/cosmos-sdk/pull/5989) `AppModuleBasic.GetTxCmd` now takes a single `CLIContext` parameter.
    * (modules) [\#5664](https://github.com/cosmos/cosmos-sdk/pull/5664) Remove amino `Codec` from simulation `StoreDecoder`, which now returns a function closure in order to unmarshal the key-value pairs.
    * (modules) [\#5555](https://github.com/cosmos/cosmos-sdk/pull/5555) Move `x/auth/client/utils/` types and functions to `x/auth/client/`.
    * (modules) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Move account balance logic and APIs from `x/auth` to `x/bank`.
    * (modules) [\#6326](https://github.com/cosmos/cosmos-sdk/pull/6326) `AppModuleBasic.GetQueryCmd` now takes a single `client.Context` parameter.
    * (modules) [\#6336](https://github.com/cosmos/cosmos-sdk/pull/6336) `AppModuleBasic.RegisterQueryService` method was added to support gRPC queries, and `QuerierRoute` and `NewQuerierHandler` were deprecated.
    * (modules) [\#6311](https://github.com/cosmos/cosmos-sdk/issues/6311) Remove `alias.go` usage
    * (modules) [\#6447](https://github.com/cosmos/cosmos-sdk/issues/6447) Rename `blacklistedAddrs` to `blockedAddrs`.
    * (modules) [\#6834](https://github.com/cosmos/cosmos-sdk/issues/6834) Add `RegisterInterfaces` method to `AppModuleBasic` to support registration of protobuf interface types.
    * (modules) [\#6734](https://github.com/cosmos/cosmos-sdk/issues/6834) Add `TxEncodingConfig` parameter to `AppModuleBasic.ValidateGenesis` command to support JSON tx decoding in `genutil`.
    * (modules) [#7764](https://github.com/cosmos/cosmos-sdk/pull/7764) Added module initialization options:
        * `server/types.AppExporter` requires extra argument: `AppOptions`.
        * `server.AddCommands` requires extra argument: `addStartFlags types.ModuleInitFlags`
        * `x/crisis.NewAppModule` has a new attribute: `skipGenesisInvariants`. [PR](https://github.com/cosmos/cosmos-sdk/pull/7764)
    * (types) [\#6327](https://github.com/cosmos/cosmos-sdk/pull/6327) `sdk.Msg` now inherits `proto.Message`, as a result all `sdk.Msg` types now use pointer semantics.
    * (types) [\#7032](https://github.com/cosmos/cosmos-sdk/pull/7032) All types ending with `ID` (e.g. `ProposalID`) now end with `Id` (e.g. `ProposalId`), to match default Protobuf generated format. Also see [\#7033](https://github.com/cosmos/cosmos-sdk/pull/7033) for more details.
    * (x/auth) [\#6029](https://github.com/cosmos/cosmos-sdk/pull/6029) Module accounts have been moved from `x/supply` to `x/auth`.
    * (x/auth) [\#6443](https://github.com/cosmos/cosmos-sdk/issues/6443) Move `FeeTx` and `TxWithMemo` interfaces from `x/auth/ante` to `types`.
    * (x/auth) [\#7006](https://github.com/cosmos/cosmos-sdk/pull/7006) All `AccountRetriever` methods now take `client.Context` as a parameter instead of as a struct member.
    * (x/auth) [\#6270](https://github.com/cosmos/cosmos-sdk/pull/6270) The passphrase argument has been removed from the signature of the following functions and methods: `BuildAndSign`, ` MakeSignature`, ` SignStdTx`, `TxBuilder.BuildAndSign`, `TxBuilder.Sign`, `TxBuilder.SignStdTx`
    * (x/auth) [\#6428](https://github.com/cosmos/cosmos-sdk/issues/6428):
        * `NewAnteHandler` and `NewSigVerificationDecorator` both now take a `SignModeHandler` parameter.
        * `SignatureVerificationGasConsumer` now has the signature: `func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error`.
        * The `SigVerifiableTx` interface now has a `GetSignaturesV2() ([]signing.SignatureV2, error)` method and no longer has the `GetSignBytes` method.
    * (x/auth/tx) [\#8106](https://github.com/cosmos/cosmos-sdk/pull/8106) change related to missing append functionality in
      client transaction signing
        + added `overwriteSig` argument to `x/auth/client.SignTx` and `client/tx.Sign` functions.
        + removed `x/auth/tx.go:wrapper.GetSignatures`. The `wrapper` provides `TxBuilder` functionality, and it's a private
          structure. That function was not used at all and it's not exposed through the `TxBuilder` interface.
    * (x/bank) [\#7327](https://github.com/cosmos/cosmos-sdk/pull/7327) AddCoins and SubtractCoins no longer return a resultingValue and will only return an error.
    * (x/capability) [#7918](https://github.com/cosmos/cosmos-sdk/pull/7918) Add x/capability safety checks:
        * All outward facing APIs will now check that capability is not nil and name is not empty before performing any state-machine changes
        * `SetIndex` has been renamed to `InitializeIndex`
    * (x/evidence) [\#7251](https://github.com/cosmos/cosmos-sdk/pull/7251) New evidence types and light client evidence handling. The module function names changed.
    * (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove APIs for getting and setting `x/evidence` parameters. `BaseApp` now uses a `ParamStore` to manage Tendermint consensus parameters which is managed via the `x/params` `Substore` type.
    * (x/gov) [\#6147](https://github.com/cosmos/cosmos-sdk/pull/6147) The `Content` field on `Proposal` and `MsgSubmitProposal`
      is now `Any` in concordance with [ADR 019](docs/architecture/adr-019-protobuf-state-encoding.md) and `GetContent` should now
      be used to retrieve the actual proposal `Content`. Also the `NewMsgSubmitProposal` constructor now may return an `error`
    * (x/ibc) [\#6374](https://github.com/cosmos/cosmos-sdk/pull/6374) `VerifyMembership` and `VerifyNonMembership` now take a `specs []string` argument to specify the proof format used for verification. Most SDK chains can simply use `commitmenttypes.GetSDKSpecs()` for this argument.
    * (x/params) [\#5619](https://github.com/cosmos/cosmos-sdk/pull/5619) The `x/params` keeper now accepts a `codec.Marshaller` instead of
      a reference to an amino codec. Amino is still used for JSON serialization.
    * (x/staking) [\#6451](https://github.com/cosmos/cosmos-sdk/pull/6451) `DefaultParamspace` and `ParamKeyTable` in staking module are moved from keeper to types to enforce consistency.
    * (x/staking) [\#7419](https://github.com/cosmos/cosmos-sdk/pull/7419) The `TmConsPubKey` method on ValidatorI has been
      removed and replaced instead by `ConsPubKey` (which returns a SDK `cryptotypes.PubKey`) and `TmConsPublicKey` (which
      returns a Tendermint proto PublicKey).
    * (x/staking/types) [\#7447](https://github.com/cosmos/cosmos-sdk/issues/7447) Remove bech32 PubKey support:
        * `ValidatorI` interface update. `GetConsPubKey` renamed to `TmConsPubKey` (consensus public key must be a tendermint key). `TmConsPubKey`, `GetConsAddr` methods return error.
        * `Validator` update. Methods changed in `ValidatorI` (as described above) and `ToTmValidator` return error.
        * `Validator.ConsensusPubkey` type changed from `string` to `codectypes.Any`.
        * `MsgCreateValidator.Pubkey` type changed from `string` to `codectypes.Any`.
    * (x/supply) [\#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) All `x/supply` types and APIs have been moved to `x/bank`.
    * [\#6409](https://github.com/cosmos/cosmos-sdk/pull/6409) Rename all IsEmpty methods to Empty across the codebase and enforce consistency.
    * [\#6231](https://github.com/cosmos/cosmos-sdk/pull/6231) Simplify `AppModule` interface, `Route` and `NewHandler` methods become only `Route`
      and returns a new `Route` type.
    * (x/slashing) [\#6212](https://github.com/cosmos/cosmos-sdk/pull/6212) Remove `Get*` prefixes from key construction functions
    * (server) [\#6079](https://github.com/cosmos/cosmos-sdk/pull/6079) Remove `UpgradeOldPrivValFile` (deprecated in Tendermint Core v0.28).
    * [\#5719](https://github.com/cosmos/cosmos-sdk/pull/5719) Bump Go requirement to 1.14+


### State Machine Breaking

* __General__
    * (client) [\#7268](https://github.com/cosmos/cosmos-sdk/pull/7268) / [\#7147](https://github.com/cosmos/cosmos-sdk/pull/7147) Introduce new protobuf based PubKeys, and migrate PubKey in BaseAccount to use this new protobuf based PubKey format

* __Modules__
    * (modules) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Separate balance from accounts per ADR 004.
        * Account balances are now persisted and retrieved via the `x/bank` module.
        * Vesting account interface has been modified to account for changes.
        * Callers to `NewBaseVestingAccount` are responsible for verifying account balance in relation to
          the original vesting amount.
        * The `SendKeeper` and `ViewKeeper` interfaces in `x/bank` have been modified to account for changes.
    * (x/auth) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/auth` module to use Protocol Buffers for state
      serialization instead of Amino.
        * The `BaseAccount.PubKey` field is now represented as a Bech32 string instead of a `crypto.Pubkey`.
        * `NewBaseAccountWithAddress` now returns a reference to a `BaseAccount`.
        * The `x/auth` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
          requiring a concrete codec to know how to serialize accounts.
        * The `AccountRetriever` type now accepts a `Codec` in its constructor in order to know how to
          serialize accounts.
    * (x/bank) [\#6518](https://github.com/cosmos/cosmos-sdk/pull/6518) Support for global and per-denomination send enabled flags.
        * Existing send_enabled global flag has been moved into a Params structure as `default_send_enabled`.
        * An array of: `{denom: string, enabled: bool}` is added to bank Params to support per-denomination override of global default value.
    * (x/distribution) [\#5610](https://github.com/cosmos/cosmos-sdk/pull/5610) Migrate the `x/distribution` module to use Protocol Buffers for state
      serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
      for JSON encoding.
        * `ValidatorHistoricalRewards.ReferenceCount` is now of types `uint32` instead of `uint16`.
        * `ValidatorSlashEvents` is now a struct with `slashevents`.
        * `ValidatorOutstandingRewards` is now a struct with `rewards`.
        * `ValidatorAccumulatedCommission` is now a struct with `commission`.
        * The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
          provided is specified by `ModuleCdc`.
    * (x/evidence) [\#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/evidence` module to use Protocol Buffers for state
      serialization instead of Amino.
        * The `internal` sub-package has been removed in order to expose the types proto file.
        * The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
          requiring a concrete codec to know how to serialize `Evidence` types.
        * The `MsgSubmitEvidence` message has been removed in favor of `MsgSubmitEvidenceBase`. The application-level
          codec must now define the concrete `MsgSubmitEvidence` type which must implement the module's `MsgSubmitEvidence`
          interface.
    * (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove parameters from `x/evidence` genesis and module state. The `x/evidence` module now solely uses Tendermint consensus parameters to determine of evidence is valid or not.
    * (x/gov) [\#5737](https://github.com/cosmos/cosmos-sdk/pull/5737) Migrate the `x/gov` module to use Protocol
      Buffers for state serialization instead of Amino.
        * `MsgSubmitProposal` will be removed in favor of the application-level proto-defined `MsgSubmitProposal` which
          implements the `MsgSubmitProposalI` interface. Applications should extend the `NewMsgSubmitProposalBase` type
          to define their own concrete `MsgSubmitProposal` types.
        * The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
          requiring a concrete codec to know how to serialize `Proposal` types.
    * (x/mint) [\#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/mint` module to use Protocol Buffers for state
      serialization instead of Amino.
        * The `internal` sub-package has been removed in order to expose the types proto file.
    * (x/slashing) [\#5627](https://github.com/cosmos/cosmos-sdk/pull/5627) Migrate the `x/slashing` module to use Protocol Buffers for state
      serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
      for JSON encoding.
        * The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
          provided is specified by `ModuleCdc`.
    * (x/staking) [\#6844](https://github.com/cosmos/cosmos-sdk/pull/6844) Validators are now inserted into the unbonding queue based on their unbonding time and height. The relevant keeper APIs are modified to reflect these changes by now also requiring a height.
    * (x/staking) [\#6061](https://github.com/cosmos/cosmos-sdk/pull/6061) Allow a validator to immediately unjail when no signing info is present due to
      falling below their minimum self-delegation and never having been bonded. The validator may immediately unjail once they've met their minimum self-delegation.
    * (x/staking) [\#5600](https://github.com/cosmos/cosmos-sdk/pull/5600) Migrate the `x/staking` module to use Protocol Buffers for state
      serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
      for JSON encoding.
        * `BondStatus` is now of type `int32` instead of `byte`.
        * Types of `int16` in the `Params` type are now of type `int32`.
        * Every reference of `crypto.Pubkey` in context of a `Validator` is now of type string. `GetPubKeyFromBech32` must be used to get the `crypto.Pubkey`.
        * The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
          provided is specified by `ModuleCdc`.
    * (x/staking) [\#7979](https://github.com/cosmos/cosmos-sdk/pull/7979) keeper pubkey storage serialization migration
      from bech32 to protobuf.
    * (x/supply) [\#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) Removed the `x/supply` module by merging the existing types and APIs into the `x/bank` module.
    * (x/supply) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/supply` module to use Protocol Buffers for state
      serialization instead of Amino.
        * The `internal` sub-package has been removed in order to expose the types proto file.
        * The `x/supply` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
          requiring a concrete codec to know how to serialize `SupplyI` types.
        * The `SupplyI` interface has been modified to no longer return `SupplyI` on methods. Instead the
          concrete type's receiver should modify the type.
    * (x/upgrade) [\#5659](https://github.com/cosmos/cosmos-sdk/pull/5659) Migrate the `x/upgrade` module to use Protocol
      Buffers for state serialization instead of Amino.
        * The `internal` sub-package has been removed in order to expose the types proto file.
        * The `x/upgrade` module now accepts a `codec.Marshaler` interface.

## [v0.39.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.1) - 2020-08-11

### Client Breaking

* (x/auth) [\#6861](https://github.com/cosmos/cosmos-sdk/pull/6861) Remove public key Bech32 encoding for all account types for JSON serialization, instead relying on direct Amino encoding. In addition, JSON serialization utilizes Amino instead of the Go stdlib, so integers are treated as strings.

## [v0.39.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.39.0) - 2020-07-20

### API Breaking Changes

* (baseapp) [\#5837](https://github.com/cosmos/cosmos-sdk/issues/5837) Transaction simulation now returns a `SimulationResponse` which contains the `GasInfo` and `Result` from the execution.

### Client Breaking Changes

* (x/auth) [\#6745](https://github.com/cosmos/cosmos-sdk/issues/6745) Remove BaseAccount's custom JSON {,un}marshalling.

## [v0.38.5] - 2020-07-02

## [v0.38.4] - 2020-05-21

## [v0.38.3] - 2020-04-09

## [v0.38.2] - 2020-03-25

## [v0.38.1] - 2020-02-11

## [v0.38.0] - 2020-01-23

### State Machine Breaking

* (genesis) [\#5506](https://github.com/cosmos/cosmos-sdk/pull/5506) The `x/distribution` genesis state
  now includes `params` instead of individual parameters.
* (genesis) [\#5017](https://github.com/cosmos/cosmos-sdk/pull/5017) The `x/genaccounts` module has been
  deprecated and all components removed except the `legacy/` package. This requires changes to the
  genesis state. Namely, `accounts` now exist under `app_state.auth.accounts`. The corresponding migration
  logic has been implemented for v0.38 target version. Applications can migrate via:
  `$ {appd} migrate v0.38 genesis.json`.
* (modules) [\#5299](https://github.com/cosmos/cosmos-sdk/pull/5299) Handling of `ABCIEvidenceTypeDuplicateVote`
  during `BeginBlock` along with the corresponding parameters (`MaxEvidenceAge`) have moved from the
  `x/slashing` module to the `x/evidence` module.

### API Breaking Changes

* (modules) [\#5506](https://github.com/cosmos/cosmos-sdk/pull/5506) Remove individual setters of `x/distribution` parameters. Instead, follow the module spec in getting parameters, setting new value(s) and finally calling `SetParams`.
* (types) [\#5495](https://github.com/cosmos/cosmos-sdk/pull/5495) Remove redundant `(Must)Bech32ify*` and `(Must)Get*KeyBech32` functions in favor of `(Must)Bech32ifyPubKey` and `(Must)GetPubKeyFromBech32` respectively, both of which take a `Bech32PubKeyType` (string).
* (types) [\#5430](https://github.com/cosmos/cosmos-sdk/pull/5430) `DecCoins#Add` parameter changed from `DecCoins`
  to `...DecCoin`, `Coins#Add` parameter changed from `Coins` to `...Coin`.
* (baseapp/types) [\#5421](https://github.com/cosmos/cosmos-sdk/pull/5421) The `Error` interface (`types/errors.go`)
  has been removed in favor of the concrete type defined in `types/errors/` which implements the standard `error` interface.
    * As a result, the `Handler` and `Querier` implementations now return a standard `error`.
      Within `BaseApp`, `runTx` now returns a `(GasInfo, *Result, error)` tuple and `runMsgs` returns a
      `(*Result, error)` tuple. A reference to a `Result` is now used to indicate success whereas an error
      signals an invalid message or failed message execution. As a result, the fields `Code`, `Codespace`,
      `GasWanted`, and `GasUsed` have been removed the `Result` type. The latter two fields are now found
      in the `GasInfo` type which is always returned regardless of execution outcome.
    * Note to developers: Since all handlers and queriers must now return a standard `error`, the `types/errors/`
      package contains all the relevant and pre-registered errors that you typically work with. A typical
      error returned will look like `sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "...")`. You can retrieve
      relevant ABCI information from the error via `ABCIInfo`.
* (client) [\#5442](https://github.com/cosmos/cosmos-sdk/pull/5442) Remove client/alias.go as it's not necessary and
  components can be imported directly from the packages.
* (store) [\#4748](https://github.com/cosmos/cosmos-sdk/pull/4748) The `CommitMultiStore` interface
  now requires a `SetInterBlockCache` method. Applications that do not wish to support this can simply
  have this method perform a no-op.
* (modules) [\#4665](https://github.com/cosmos/cosmos-sdk/issues/4665) Refactored `x/gov` module structure and dev-UX:
    * Prepare for module spec integration
    * Update gov keys to use big endian encoding instead of little endian
* (modules) [\#5017](https://github.com/cosmos/cosmos-sdk/pull/5017) The `x/genaccounts` module has been deprecated and all components removed except the `legacy/` package.
* [\#4486](https://github.com/cosmos/cosmos-sdk/issues/4486) Vesting account types decoupled from the `x/auth` module and now live under `x/auth/vesting`. Applications wishing to use vesting account types must be sure to register types via `RegisterCodec` under the new vesting package.
* [\#4486](https://github.com/cosmos/cosmos-sdk/issues/4486) The `NewBaseVestingAccount` constructor returns an error
  if the provided arguments are invalid.
* (x/auth) [\#5006](https://github.com/cosmos/cosmos-sdk/pull/5006) Modular `AnteHandler` via composable decorators:
    * The `AnteHandler` interface now returns `(newCtx Context, err error)` instead of `(newCtx Context, result sdk.Result, abort bool)`
    * The `NewAnteHandler` function returns an `AnteHandler` function that returns the new `AnteHandler`
      interface and has been moved into the `auth/ante` directory.
    * `ValidateSigCount`, `ValidateMemo`, `ProcessPubKey`, `EnsureSufficientMempoolFee`, and `GetSignBytes`
      have all been removed as public functions.
    * Invalid Signatures may return `InvalidPubKey` instead of `Unauthorized` error, since the transaction
      will first hit `SetPubKeyDecorator` before the `SigVerificationDecorator` runs.
    * `StdTx#GetSignatures` will return an array of just signature byte slices `[][]byte` instead of
      returning an array of `StdSignature` structs. To replicate the old behavior, use the public field
      `StdTx.Signatures` to get back the array of StdSignatures `[]StdSignature`.
* (modules) [\#5299](https://github.com/cosmos/cosmos-sdk/pull/5299) `HandleDoubleSign` along with params `MaxEvidenceAge` and `DoubleSignJailEndTime` have moved from the `x/slashing` module to the `x/evidence` module.
* (keys) [\#4941](https://github.com/cosmos/cosmos-sdk/issues/4941) Keybase concrete types constructors such as `NewKeyBaseFromDir` and `NewInMemory` now accept optional parameters of type `KeybaseOption`. These
  optional parameters are also added on the keys sub-commands functions, which are now public, and allows
  these options to be set on the commands or ignored to default to previous behavior.
* [\#5547](https://github.com/cosmos/cosmos-sdk/pull/5547) `NewKeyBaseFromHomeFlag` constructor has been removed.
* [\#5439](https://github.com/cosmos/cosmos-sdk/pull/5439) Further modularization was done to the `keybase`
  package to make it more suitable for use with different key formats and algorithms:
    * The `WithKeygenFunc` function added as a `KeybaseOption` which allows a custom bytes to key
      implementation to be defined when keys are created.
    * The `WithDeriveFunc` function added as a `KeybaseOption` allows custom logic for deriving a key
      from a mnemonic, bip39 password, and HD Path.
    * BIP44 is no longer build into `keybase.CreateAccount()`. It is however the default when using
      the `client/keys` add command.
    * `SupportedAlgos` and `SupportedAlgosLedger` functions return a slice of `SigningAlgo`s that are
      supported by the keybase and the ledger integration respectively.
* (simapp) [\#5419](https://github.com/cosmos/cosmos-sdk/pull/5419) The `helpers.GenTx()` now accepts a gas argument.
* (baseapp) [\#5455](https://github.com/cosmos/cosmos-sdk/issues/5455) A `sdk.Context` is now passed into the `router.Route()` function.

### Client Breaking Changes

* (rest) [\#5270](https://github.com/cosmos/cosmos-sdk/issues/5270) All account types now implement custom JSON serialization.
* (rest) [\#4783](https://github.com/cosmos/cosmos-sdk/issues/4783) The balance field in the DelegationResponse type is now sdk.Coin instead of sdk.Int
* (x/auth) [\#5006](https://github.com/cosmos/cosmos-sdk/pull/5006) The gas required to pass the `AnteHandler` has
  increased significantly due to modular `AnteHandler` support. Increase GasLimit accordingly.
* (rest) [\#5336](https://github.com/cosmos/cosmos-sdk/issues/5336) `MsgEditValidator` uses `description` instead of `Description` as a JSON key.
* (keys) [\#5097](https://github.com/cosmos/cosmos-sdk/pull/5097) Due to the keybase -> keyring transition, keys need to be migrated. See `keys migrate` command for more info.
* (x/auth) [\#5424](https://github.com/cosmos/cosmos-sdk/issues/5424) Drop `decode-tx` command from x/auth/client/cli, duplicate of the `decode` command.

## [v0.37.14] - 2020-08-12

# - end -
