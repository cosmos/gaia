package evm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v23/ante/evm/decorators"
	"github.com/cosmos/gaia/v23/ante/handler_options"
)

func NewAnteHandler(opts handler_options.HandlerOptions) sdk.AnteHandler {
	anteDecorators := []sdk.AnteDecorator{
		decorators.NewMsgTypeDecorator(),
		decorators.NewGovVoteDecorator(opts.Codec, opts.StakingKeeper),
		decorators.NewGovExpeditedProposalsDecorator(opts.Codec),
	}
	return sdk.ChainAnteDecorators(anteDecorators...)
}
