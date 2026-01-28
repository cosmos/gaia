package gov

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/gaia/v26/ante"
)

// msgServer wraps the SDK gov MsgServer to add vote validation.
// This ensures all vote messages are validated for stake requirements,
// regardless of their origin (user tx, ICA, wasm, authz, etc.).
type msgServer struct {
	govv1.MsgServer
	stakingKeeper *stakingkeeper.Keeper
}

var _ govv1.MsgServer = &msgServer{}

// NewMsgServerImpl returns an implementation of the gov MsgServer interface
// that validates voter stake before delegating to the SDK implementation.
func NewMsgServerImpl(keeper *govkeeper.Keeper, sk *stakingkeeper.Keeper) govv1.MsgServer {
	return &msgServer{
		MsgServer:     govkeeper.NewMsgServerImpl(keeper),
		stakingKeeper: sk,
	}
}

// Vote validates that the voter has sufficient stake before processing the vote.
func (m *msgServer) Vote(ctx context.Context, msg *govv1.MsgVote) (*govv1.MsgVoteResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}

	if err := ante.ValidateVoterStake(sdkCtx, m.stakingKeeper, voter); err != nil {
		return nil, err
	}

	return m.MsgServer.Vote(ctx, msg)
}

// VoteWeighted validates that the voter has sufficient stake before processing the weighted vote.
func (m *msgServer) VoteWeighted(ctx context.Context, msg *govv1.MsgVoteWeighted) (*govv1.MsgVoteWeightedResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, err
	}

	if err := ante.ValidateVoterStake(sdkCtx, m.stakingKeeper, voter); err != nil {
		return nil, err
	}

	return m.MsgServer.VoteWeighted(ctx, msg)
}
