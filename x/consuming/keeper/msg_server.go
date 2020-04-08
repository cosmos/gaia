package keeper

import (
	"context"

	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) RequestData(goCtx context.Context, msg *types.MsgRequestData) (*types.MsgRequestDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sourceChannelEnd, found := k.ChannelKeeper.GetChannel(ctx, "consuming", msg.SourceChannel)
	if !found {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnknownRequest,
			"unknown channel %s port consuming",
			msg.SourceChannel,
		)
	}
	destinationPort := sourceChannelEnd.Counterparty.PortId
	destinationChannel := sourceChannelEnd.Counterparty.ChannelId
	sequence, found := k.ChannelKeeper.GetNextSequenceSend(
		ctx, "consuming", msg.SourceChannel,
	)
	if !found {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrUnknownRequest,
			"unknown sequence number for channel %s port oracle",
			msg.SourceChannel,
		)
	}
	sourcePort := "band-consumer"
	packet := bandoracle.NewOracleRequestPacketData(
		"band-consumer",
		bandoracle.OracleScriptID(msg.OracleScriptID),
		msg.Calldata,
		uint64(msg.AskCount),
		uint64(msg.MinCount),
	)
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, msg.SourceChannel))
	if !ok {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}
	err := k.ChannelKeeper.SendPacket(ctx, channelCap, channeltypes.NewPacket(
		packet.GetBytes(),
		sequence,
		sourcePort,
		msg.SourceChannel,
		destinationPort,
		destinationChannel,
		clienttypes.NewHeight(100, 100),
		0, // Arbitrarily high timeout for now
	))
	if err != nil {
		return nil, err
	}
	return &types.MsgRequestDataResponse{}, nil
}
