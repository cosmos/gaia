# IBC Rate Limiting

## Overview

This `ratelimit` module is a native golang implementation, inspired by Osmosis's CosmWasm [`ibc-rate-limit`](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-rate-limit) module. The module is meant as a safety control in the event of a bug, attack, or economic failure of an external zone. It prevents massive inflows or outflows of IBC tokens in a short time frame. See [here](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-rate-limit#motivation) for an excellent summary by the Osmosis team on the motivation for rate limiting.

Each rate limit is applied at a ChannelOrClientID + Denom granularity and is evaluated in evenly spaced fixed windows. For instance, a rate limit might be specified on `uosmo` (denominated as `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` on Stride), on the Stride <-> Osmosis transfer channel (`channel-5`), with a 24 hour window.

Each rate limit will also have a configurable threshold that dictates the max inflow/outflow along the channel. The threshold is represented as a percentage of the total value along the channel. The channel value is calculated by querying the total supply of the denom at the start of the time window, and it remains constant until the window expires. For instance, the rate limit described above might have a threshold of 10% for both inflow and outflow. If the total supply of `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` was 100, then any transfer that would cause a net inflow or outflow greater than 10 (i.e. greater than 10% the channel value) would be rejected. Once the time window expires, the net inflow and outflow are reset to 0 and the channel value is re-calculated.

The _net_ inflow and outflow is used (rather than the total inflow/outflow) to prevent DOS attacks where someone repeatedly sends the same token back and forth across the same channel, causing the rate limit to be reached.

The module is implemented as IBC Middleware around the transfer module. An "hour epoch" abstraction is leveraged to determine when each rate limit window has expired (each window is denominated in hours). This means all rate limit windows with the same window duration will start and end at the same time. In the case of a 24 hour rate limit window, the rate limit will reset at the end of the day in UTC (i.e. 00:00 UTC).

Note: channels are removed in IBC v2, thus client IDs are used instead of channel IDs for IBC v2. IBC v1 rate limits should use channel IDs, while IBC v2 rate limits should use client IDs.

## Integration
To add the rate limit module, wire it up in `app.go` in line with the following example. The module must be included in a middleware stack alongside the transfer module.

```go
// app.go

// Import the rate limit module
import (
  "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/ratelimit"
  ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/ratelimit/keeper"
  ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/ratelimit/types"
)

...

// Register the AppModule
ModuleBasics = module.NewBasicManager(
    ...
    ratelimit.AppModuleBasic{},
    ...
)

...

// Add the RatelimitKeeper to the App
type App struct {
  ...
  RatelimitKeeper ratelimitkeeper.Keeper
  ...
}

// Add the store key
keys := sdk.NewKVStoreKeys(
  ...
  ratelimittypes.StoreKey,
  ...
)

// Create the rate limit keeper
app.RatelimitKeeper = *ratelimitkeeper.NewKeeper(
  appCodec,
  keys[ratelimittypes.StoreKey],
  app.GetSubspace(ratelimittypes.ModuleName),
  authtypes.NewModuleAddress(govtypes.ModuleName).String(),
  app.BankKeeper,
  app.IBCKeeper.ChannelKeeper,
  app.IBCKeeper.ChannelKeeper, // ICS4Wrapper
)

// Add the rate limit module to a middleware stack with the transfer module
//
// Note: If the integrating chain already has middleware wired, you'll just
// have to add the rate limit module to the existing stack
//
// The following will create the following stack
// - Core IBC
// - ratelimit
// - transfer
// - base app
var transferStack ibcporttypes.IBCModule = transfer.NewIBCModule(app.TransferKeeper)
transferStack = ratelimit.NewIBCMiddleware(app.RatelimitKeeper, transferStack)

// Add IBC Router
ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)

// Add the rate limit module to the module manager
app.mm = module.NewManager(
  ...
  ratelimit.NewAppModule(appCodec, app.RatelimitKeeper),
)

// Add the rate limit module to begin and end blockers, and init genesis
app.mm.SetOrderBeginBlockers(
  ...
  ratelimittypes.ModuleName,
)

app.mm.SetOrderEndBlockers(
  ...
  ratelimittypes.ModuleName,
)

genesisModuleOrder := []string{
  ...
  ratelimittypes.ModuleName,
}

// Add the rate limit module to the params keeper
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
  ...
  paramsKeeper.Subspace(ratelimittypes.ModuleName)
  ...
}
```

## Implementation

Each rate limit is defined by the following three components:

1. **Path**: Defines the `ChannelOrClientId` and `Denom`
2. **Quota**: Defines the rate limit time window (`DurationHours`) and the max threshold for inflows/outflows (`MaxPercentRecv` and `MaxPercentSend` respectively)
3. **Flow**: Stores the current `Inflow`, `Outflow` and `ChannelValue`. Each time a quota expires, the inflow and outflow get reset to 0 and the channel value gets recalculated. Throughout the window, the inflow and outflow each increase monotonically. The net flow is used when determining if a transfer would exceed the quota.
   - For `Send` packets:
     $$\text{Exceeds Quota if:} \left(\frac{\text{Outflow} - \text{Inflow} + \text{Packet Amount}}{\text{ChannelValue}}\right) > \text{MaxPercentSend}$$
   - For `Receive` packets:
     $$\text{Exceeds Quota if:} \left(\frac{\text{Inflow} - \text{Outflow} + \text{Packet Amount}}{\text{ChannelValue}}\right) > \text{MaxPercentRecv}$$

## Example Walk-Through

Using the example above, let's say we created a 24 hour rate limit on `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` ("`ibc/uosmo`"), `channel-5`, on Stride, with a 10% send and receive threshold.

1. At the start of the window, the supply will be queried, to determine the channel value. Let's say the total supply was 100
2. If someone transferred `8uosmo` from `Osmosis -> Stride`, the `Inflow` would increment by 8
3. If someone tried to transfer another `8uosmo` from `Osmosis -> Stride`, it would exceed the quota since `(8+8)/100 = 16%` (which is greater than 10%) and thus, the transfer would be rejected.
4. If someone tried to transfer `12ibc/uosmo` from Stride -> Osmosis, the `Outflow` would increment by 12. Notice, even though 12 is greater than 10% the total channel value, the _net_ outflow is only `4uatom` (since it's offset by the `8uatom` `Inflow`). As a result, this transaction would succeed.
5. Now if the person in (3) attempted to retry their transfer of`8uosmo` from `Osmosis -> Stride`, the `Inflow` would increment by 8 and the transaction would succeed (leaving a net inflow of 4).
6. Finally, at the end of the 24 hours, the `Inflow` and `Outflow` would get reset to 0 and the `ChannelValue` would be re-calculated. In this example, the new channel value would be 104 (since more `uosmo` was sent to Stride, and thus more `ibc/uosmo` was minted)

| Step |           Description            | Transfer Status | Inflow | Outflow | Net Inflow | Net Outflow | Channel Value |
| :--: | :------------------------------: | :-------------: | :----: | :-----: | :--------: | :---------: | :-----------: |
|  1   |        Rate limit created        |                 |   0    |    0    |            |             |      100      |
|  2   |     8usomo Osmosis → Stride      |   Successful    |   8    |    0    |     8%     |             |      100      |
|  3   |     8usomo Osmosis → Stride      |    Rejected     |   16   |    0    | 16% (>10%) |             |      100      |
|  3   | State reverted after rejected Tx |                 |   8    |    0    |     8%     |             |      100      |
|  4   |   12ibc/uosmo Stride → Osmosis   |   Successful    |   8    |   12    |            |     4%      |      100      |
|  5   |     8usomo Osmosis → Stride      |   Successful    |   16   |   12    |     4%     |             |      100      |
|  6   |           Quota Reset            |                 |   0    |    0    |            |             |      104      |

## Denom Blacklist

The module also contains a blacklist to completely halt all IBC transfers for a given denom. There are keeper functions to add or remove denoms from the blacklist; however, these functions are not exposed externally through transactions or governance, and they should only be leveraged internally from the protocol in extreme scenarios.

## Address Whitelist

There is also a whitelist, mainly used to exclude protocol-owned accounts. For instance, Stride periodically bundles liquid staking deposits and transfers in a single transaction at the top of the epoch. Without a whitelist, this transfer would make the rate limit more likely to trigger a false positive.

## Denoms

We always want to refer to the channel or client ID and denom as they appear on the rate limited chain. For instance, in the example above where rate limiting was added to Stride, we would store the rate limit with denom `ibc/D24B4564BCD51D3D02D9987D92571EAC5915676A9BD6D9B0C1D0254CB8A5EA34` and `channel-5` (the ChannelID on Stride), instead of `uosmo` and `channel-326` (the ChannelID on Osmosis).

However, since the ratelimit module acts as middleware to the transfer module, the respective denoms need to be interpreted using the denom trace associated with each packet. There are a few scenarios at play here...

### Send Packets

The denom that the rate limiter will use for a send packet depends on whether it was a native token (i.e. tokens minted on the rate limited chain) or non-native token (e.g. ibc/...)...

#### Native vs Non-Native

- We can identify if the token is native or not by parsing the denom trace from the packet
  - If the token is **native**, it **will not** have a prefix (e.g. `ustrd`)
  - If the token is **non-native**, it **will** have a prefix (e.g. `transfer/channel-X/uosmo`)

#### Determining the denom in the rate limit

- For **native** tokens, return as is (e.g. `ustrd`)
- For **non-native** tokens, take the ibc hash (e.g. hash `transfer/channel-X/uosmo` into `ibc/...`)

### Receive Packets

The denom that the rate limiter will use for a receive packet depends on whether it was a source or sink.

#### Source vs Sink

As a token travels across IBC chains, its path is recorded in the denom trace.

- **Sink**: If the token moves **forward**, to a chain different than its previous hop, the destination chain acts as a **sink zone**, and the new port and channel are **appended** to the denom trace.
  - Ex1: `uatom` is sent from Cosmoshub to Stride
    - Stride is the first destination for `uatom`, and acts as a sink zone
    - The IBC denom becomes the hash of: `/{stride-port)/{stride-channel}/uatom`
  - Ex2: `uatom` is sent from Cosmoshub to Osmosis then to Stride
    - Here the receiving chain (Stride) is not the same as the previous hop (Cosmoshub), so Stride, once again, is acting as a sink zone
    - The IBC denom becomes the hash of: `/{stride-port)/{stride-channel}/{osmosis-port}/{osmosis-channel}/uatom`
- **Source**: If the token moves **backwards** (i.e. revisits the last chain it was sent from), the destination chain is acting as a **source zone**, and the port and channel are **removed** from the denom trace - undoing the last hop. Should a token reverse its course completely and head back along the same path to its native chain, the denom trace will unwind and reduce back down to the original base denom.
  - Ex1: `ustrd` is sent from Stride to Osmosis, and then back to Stride
    - Here the trace reduces from `/{osmosis-port}/{osmosis-channel}/ustrd` to simply `ustrd`
  - Ex2: `ujuno` is sent to Stride, then to Osmosis, then back to Stride
    - Here the trace reduces from `/{osmosis-port}/{osmosis-channel}/{stride-port}/{stride-channel}/ujuno` to just `/{stride-port}/{stride-channel}/ujuno` (the Osmosis hop is removed)
  - Stride is the source in the examples above because the token went back and forth from Stride -> Osmosis -> Stride

For a more detailed explanation, see the[ ICS-20 ADR](https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-001-coin-source-tracing.md#example) and [spec](https://github.com/cosmos/ibc/tree/main/spec/app/ics-020-fungible-token-transfer).

#### Determining the denom in the rate limit

- If the chain is acting as a **Sink**: Add on the port and channel from the rate limited chain and hash it

  - Ex1: `uosmo` sent from Osmosis to Stride

    - Packet Denom Trace: `uosmo`
    - (1) Add Stride Channel as Prefix: `transfer/channel-X/uosmo`
    - (2) Hash: `ibc/...`

  - Ex2: `ujuno` sent from Osmosis to Stride
    - Packet Denom Trace: `transfer/channel-Y/ujuno` (where channel-Y is the Juno <> Osmosis channel)
    - (1) Add Stride Channel as Prefix: `transfer/channel-X/transfer/channel-Y/ujuno`
    - (2) Hash: `ibc/...`

- If the chain is acting as a **Source**: First, remove the prefix. Then if there is still a trace prefix, hash it
  - Ex1: `ustrd` sent back to Stride from Osmosis
    - Packet Denom: `transfer/channel-X/ustrd`
    - (1) Remove Prefix: `ustrd`
    - (2) No trace remaining, leave as is: `ustrd`
  - Ex2: juno was sent to Stride, then to Osmosis, then back to Stride
    - Packet Denom: `transfer/channel-X/transfer/channel-Z/ujuno`
    - (1) Remove Prefix: `transfer/channel-Z/ujuno`
    - (2) Hash: `ibc/...`

## Packet Failures and Timeouts
When a transfer is sent, the `Outflow` for the corresponding rate limit is incremented. Consequently, if the transfer fails on the host or times out, the change in `Outflow` must be reverted. However, the decrement is only necessary if the acknowledgement or timeout is returned in the same quota window that the packet was originally sent from.

To keep track of whether the packet was sent in the same quota, the sequence number of all pending packets are stored. This is implemented by recording the sequence number of a SendPacket as it is sent, and then removing that list of sequence numbers each time the rate limit is reset at the end of the quota. Additionally, the sequence numbers are also removed when after an acknowledgement or timeout (a step that is not entirely necessary, but does reduce the size of the state).

## State

```go
RateLimit
    Path
        Denom string
        ChannelOrClientId string
    Quota
        MaxPercentSend sdkmath.Int
        MaxPercentRecv sdkmath.Int
        DurationHours uint64
    Flow
        Inflow sdkmath.Int
        Outflow sdkmath.Int
        ChannelValue sdkmath.Int
```

## Keeper functions
### RateLimit
```go
// Stores a RateLimit object in the store
SetRateLimit(rateLimit types.RateLimit)

// Removes a RateLimit object from the store
RemoveRateLimit(denom string, channelOrClientId string)

// Reads a RateLimit object from the store
GetRateLimit(denom string, channelOrClientId string) (RateLimit, found)

// Gets a list of all RateLimit objects
GetAllRateLimits() []RateLimit

// Resets the Inflow and Outflow of a RateLimit and re-calculates the ChannelValue
ResetRateLimit(denom string, channelOrClientId string)
```

### PendingSendPacket
```go
// Sets the sequence number of a packet that was just sent
SetPendingSendPacket(channelOrClientId string, sequence uint64)

// Remove a pending packet sequence number from the store
// This is used after the ack or timeout for a packet has been received
RemovePendingSendPacket(channelOrClientId string, sequence uint64)

// Checks whether the packet sequence number is in the store - indicating that it was
// sent during the current quota
CheckPacketSentDuringCurrentQuota(channelOrClientId string, sequence uint64) bool

// Removes all pending sequence numbers from the store
// This is executed when the quota resets
RemoveAllChannelPendingSendPackets(channelOrClientId string)
```

### DenomBlacklist
```go
// Adds a denom to a blacklist to prevent all IBC transfers with this denom
AddDenomToBlacklist(denom string)

// Removes a denom from a blacklist to re-enable IBC transfers for that denom
RemoveDenomFromBlacklist(denom string)

// Check if a denom is currently blacklisted
IsDenomBlacklisted(denom string) bool

// Get all the blacklisted denoms
GetAllBlacklistedDenoms() []string
```

### AddressWhitelist
```go
// Adds an pair of sender and receiver addresses to the whitelist to allow all
// IBC transfers between those addresses to skip all flow calculations
SetWhitelistedAddressPair(whitelist types.WhitelistedAddressPair)

// Removes a whitelisted address pair so that it's transfers are counted in the quota
RemoveWhitelistedAddressPair(sender, receiver string)

// Check if a sender/receiver address pair is currently whitelisted
IsAddressPairWhitelisted(sender, receiver string) bool

// Get all the whitelisted addresses
GetAllWhitelistedAddressPairs() []types.WhitelistedAddressPair
```


### Business Logic
```go
// Checks whether a packet will exceed a rate limit quota
// If it does not exceed the quota, it updates the `Inflow` or `Outflow`
// If it exceeds the quota, it returns an error
CheckRateLimitAndUpdateFlow(direction types.PacketDirection, packetInfo RateLimitedPacketInfo) (updated bool)

// Reverts the change in outflow from a SendPacket if it fails or times out
UndoSendPacket(channelOrClientId string, sequence uint64, denom string, amount sdkmath.Int)
```

## Middleware Functions

```go
SendRateLimitedPacket (ICS4Wrapper SendPacket)
ReceiveRateLimitedPacket (IBCModule OnRecvPacket)
```

## Transactions (via Governance)

```go
// Adds a new rate limit
// Errors if:
//   - `ChannelValue` is 0 (meaning supply of the denom is 0)
//   - Rate limit already exists (as identified by the `channel_or_client_id` and `denom`)
//   - Channel does not exist
AddRateLimit()
{"denom": string, "channel_or_client_id": string, "duration_hours": string, "max_percent_send": string, "max_percent_recv": string}

// Updates a rate limit quota, and resets the rate limit
// Errors if:
//   - Rate limit does not exist (as identified by the `channel_or_client_id` and `denom`)
UpdateRateLimit()
{"denom": string, "channel_or_client_id": string, "duration_hours": string, "max_percent_send": string, "max_percent_recv": string}

// Resets the `Inflow` and `Outflow` of a rate limit to 0, and re-calculates the `ChannelValue`
// Errors if:
//   - Rate limit does not exist (as identified by the `channel_or_client_id` and `denom`)
ResetRateLimit()
{"denom": string, "channel_or_client_id": string}

// Removes the rate limit from the store
// Errors if:
//   - Rate limit does not exist (as identified by the `channel_or_client_id` and `denom`)
RemoveRateLimit()
{"denom": string, "channel_or_client_id": string}
```

## Queries

```go
// Queries all rate limits
//   CLI:
//      binaryd q ratelimit list-rate-limits
//   API:
//      /Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits
QueryRateLimits()

// Queries a specific rate limit given a ChannelID and Denom
//   CLI:
//      binaryd q ratelimit rate-limit [denom] [channel-id]
//   API:
//      /Stride-Labs/ibc-rate-limiting/ratelimit/ratelimit/{denom}/{channel_or_client_id}
QueryRateLimit(denom string, channelOrClientId string)

// Queries all rate limits associated with a given host chain
//   CLI:
//      binaryd q ratelimit rate-limits-by-chain [chain-id]
//   API:
//      /Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits/{chain_id}
QueryRateLimitsByChainId(chainId string)
```
