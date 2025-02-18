package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	cosmosante "github.com/cosmos/gaia/v23/ante/cosmos"
	evmante "github.com/cosmos/gaia/v23/ante/evm"
	"github.com/cosmos/gaia/v23/ante/handler_options"
)

func NewAnteHandler(opts handler_options.HandlerOptions) (sdk.AnteHandler, error) {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		txWithExtensions, ok := tx.(ante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeUrl := opts[0].GetTypeUrl(); typeUrl {
				case "/ethermint.evm.v1.ExtensionOptionsEthereumTx":
					anteHandler = evmante.NewAnteHandler()
				//case "/ethermint.types.v1.ExtensionOptionDynamicFeeTx": //todo: is this relevant?
				//	// cosmos-sdk tx with dynamic fee extension
				//	anteHandler = cosmosante.NewAnteHandler(opts)
				//}
				default:
					return ctx, errorsmod.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension type: %s", typeUrl,
					)
				}
				return anteHandler(ctx, tx, simulate)
			}
		}
		return cosmosante.NewAnteHandler(opts)(ctx, tx, simulate)
	}, nil
}
