package gaia

import (
	feemarketpost "github.com/skip-mev/feemarket/x/feemarket/post"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/gaia/v23/ante"
)

// PostHandlerOptions are the options required for constructing a FeeMarket PostHandler.
type PostHandlerOptions struct {
	AccountKeeper   feemarketpost.AccountKeeper
	BankKeeper      feemarketpost.BankKeeper
	FeeMarketKeeper feemarketpost.FeeMarketKeeper
}

// NewPostHandler returns a PostHandler chain with the fee deduct decorator.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	if !ante.UseFeeMarketDecorator {
		return nil, nil
	}

	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for post builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for post builder")
	}

	if options.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for post builder")
	}

	postDecorators := []sdk.PostDecorator{
		feemarketpost.NewFeeMarketDeductDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeeMarketKeeper,
		),
	}

	return sdk.ChainPostDecorators(postDecorators...), nil
}
