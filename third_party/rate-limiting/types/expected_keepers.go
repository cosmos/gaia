package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v11/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v11/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v11/modules/core/exported"
)

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/ratelimit keeper.
type BankKeeper interface {
	GetSupply(ctx context.Context, denom string) sdk.Coin
}

// ChannelKeeper defines the channel contract that must be fulfilled when
// creating a x/ratelimit keeper.
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, portID string, channelID string) (channeltypes.Channel, bool)
	GetChannelClientState(ctx sdk.Context, portID string, channelID string) (string, ibcexported.ClientState, error)
}

type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (ibcexported.ClientState, bool)
	GetClientStatus(ctx sdk.Context, clientID string) ibcexported.Status
}

// ICS4Wrapper defines the expected ICS4Wrapper for middleware
type ICS4Wrapper interface {
	WriteAcknowledgement(ctx sdk.Context, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error
	SendPacket(
		ctx sdk.Context,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (sequence uint64, err error)
	GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool)
}
