package gaia

import (
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/gaia/v23/ante/cosmos"
	feemarketpost "github.com/skip-mev/feemarket/x/feemarket/post"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// PostHandlerOptions are the options required for constructing a FeeMarket PostHandler.
type PostHandlerOptions struct {
	AccountKeeper   feemarketpost.AccountKeeper
	BankKeeper      feemarketpost.BankKeeper
	FeeMarketKeeper feemarketpost.FeeMarketKeeper
}

// NewPostHandler returns a PostHandler chain with the fee deduct decorator.
func NewPostHandler(options PostHandlerOptions) (sdk.PostHandler, error) {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool) (newCtx sdk.Context, err error) {
		var postHandler sdk.PostHandler
		txWithExtensions, ok := tx.(ante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeUrl := opts[0].GetTypeUrl(); typeUrl {
				case "/ethermint.evm.v1.ExtensionOptionsEthereumTx":
					postHandler = sdk.ChainPostDecorators()
				default:
					return ctx, errorsmod.Wrapf(
						sdkerrors.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension type: %s", typeUrl,
					)
				}
				if postHandler == nil {
					return ctx, nil
				}
				return postHandler(ctx, tx, simulate, success)
			}
		}
		if !cosmos.UseFeeMarketDecorator {
			return ctx, nil
		}

		if options.AccountKeeper == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for post builder")
		}

		if options.BankKeeper == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for post builder")
		}

		if options.FeeMarketKeeper == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrLogic, "feemarket keeper is required for post builder")
		}

		postDecorators := []sdk.PostDecorator{
			feemarketpost.NewFeeMarketDeductDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeeMarketKeeper,
			),
		}

		return sdk.ChainPostDecorators(postDecorators...)(ctx, tx, simulate, success)
	}, nil

}
