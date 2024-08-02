# ADR 003: Interchain Accounts Controller Module

## Changelog

- 2024-03-08: Initial Draft

## Status

Proposed

## Abstract

The Interchain Accounts Controller IBC module allows users of one chain to create and control accounts on other chains. The Hub currently doesn't have ICA Controller module enabled, so it is not possible to create accounts on other chains from the Hub chain.

## Context

Enabling the ICA Controller module on the Hub would support various use cases. One such case could be the provider-based governance that would allow the ATOM stakers to participate in a governance on consumer chains.

## Decision

The ICA Controller module will be included in the application, so the Hub will have both ICA Host and Controller modules. The implementation will use the Controller module's built-in authentication mechanism, since we don't have a need for custom authentication logic. According to this, users will directly use `MsgRegisterInterchainAccount` and `MsgSendTx` messages defined by the Controller module. The possibility provided by the Controller module to define underlying application to have custom processing of IBC messages exchanged by the Controller module (e.g. `OnChanOpenInit`, `OnAcknowledgementPacket`, etc.) will not be used, since there is currently no need for this.

```go
// ICA Controller keeper
appKeepers.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
	appCodec,
	appKeepers.keys[icacontrollertypes.StoreKey],
	appKeepers.GetSubspace(icacontrollertypes.SubModuleName),
	appKeepers.IBCKeeper.ChannelKeeper, // ICS4Wrapper
	appKeepers.IBCKeeper.ChannelKeeper,
	&appKeepers.IBCKeeper.PortKeeper,
	appKeepers.ScopedICAControllerKeeper,
	bApp.MsgServiceRouter(),
)

// Create ICA module
appKeepers.ICAModule = ica.NewAppModule(&appKeepers.ICAControllerKeeper, &appKeepers.ICAHostKeeper)

// Create Interchain Accounts Controller Stack
var icaControllerStack porttypes.IBCModule = icacontroller.NewIBCMiddleware(nil, appKeepers.ICAControllerKeeper)

// Add Interchain Accounts Controller IBC route
ibcRouter.AddRoute(icacontrollertypes.SubModuleName, icaControllerStack)
```

## Consequences

### Positive

- Users of the Hub will have a possibility to create and utilize Interchain Accounts on other IBC connected chains.

### Negative

### Neutral

- Since we don't need to implement a custom authentication mechanism, we can rely on the one defined by the Controller module itself, implemented through the `MsgRegisterInterchainAccount` and `MsgSendTx` messages.

## References

[https://github.com/cosmos/gaia/issues/2869](https://github.com/cosmos/gaia/issues/2869)