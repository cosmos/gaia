package ante

import (
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

// govVoteTypeURLs contains the type URLs for governance vote messages
// that should be validated when dispatched from wasm contracts
var govVoteTypeURLs = map[string]struct{}{
	"/cosmos.gov.v1.MsgVote":              {},
	"/cosmos.gov.v1.MsgVoteWeighted":      {},
	"/cosmos.gov.v1beta1.MsgVote":         {},
	"/cosmos.gov.v1beta1.MsgVoteWeighted": {},
}

// GovVoteMessageHandler wraps the default wasm message handler
// to validate governance votes from contracts.
// This prevents contracts from bypassing the stake requirement for voting.
type GovVoteMessageHandler struct {
	wrapped       wasmkeeper.Messenger
	stakingKeeper *stakingkeeper.Keeper
}

// NewGovVoteMessageDecorator returns a decorator function for WithMessageHandlerDecorator.
// It wraps the default message handler to validate that the contract (as voter)
// has sufficient stake before allowing governance votes.
func NewGovVoteMessageDecorator(stakingKeeper *stakingkeeper.Keeper) func(wasmkeeper.Messenger) wasmkeeper.Messenger {
	return func(wrapped wasmkeeper.Messenger) wasmkeeper.Messenger {
		return &GovVoteMessageHandler{
			wrapped:       wrapped,
			stakingKeeper: stakingKeeper,
		}
	}
}

// DispatchMsg implements the Messenger interface.
// It intercepts governance vote messages and validates that the contract
// has sufficient stake before delegating to the wrapped handler.
func (h *GovVoteMessageHandler) DispatchMsg(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	contractIBCPortID string,
	msg wasmvmtypes.CosmosMsg,
) (events []sdk.Event, data [][]byte, msgResponses [][]*codectypes.Any, err error) {
	// Check if this is a governance vote message
	if err := h.validateGovVote(ctx, contractAddr, msg); err != nil {
		return nil, nil, nil, err
	}

	// Delegate to the wrapped handler
	return h.wrapped.DispatchMsg(ctx, contractAddr, contractIBCPortID, msg)
}

// validateGovVote checks if the message is a governance vote and validates
// that the contract has sufficient stake to vote.
func (h *GovVoteMessageHandler) validateGovVote(ctx sdk.Context, contractAddr sdk.AccAddress, msg wasmvmtypes.CosmosMsg) error {
	isVoteMsg := msg.Gov != nil && msg.Gov.Vote != nil

	// Check for standard Gov.Vote message

	// Check for Gov.VoteWeighted message
	if msg.Gov != nil && msg.Gov.VoteWeighted != nil {
		isVoteMsg = true
	}

	// Check for Any message with governance vote type URLs
	if msg.Any != nil {
		if _, ok := govVoteTypeURLs[msg.Any.TypeURL]; ok {
			isVoteMsg = true
		}
	}

	if !isVoteMsg {
		return nil
	}

	// For governance votes, the contract itself is the voter
	// Validate that the contract has sufficient stake
	return ValidateVoterStake(ctx, h.stakingKeeper, contractAddr)
}
