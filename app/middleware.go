package gaia

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	gaiaante "github.com/cosmos/gaia/v7/ante"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibcmiddleware "github.com/cosmos/ibc-go/v3/modules/core/middleware"
	wasmkeeper "github.com/cosmos/wasmd/x/wasm/keeper"
	wasmTypes "github.com/cosmos/wasmd/x/wasm/types"
)

// ComposeMiddlewares compose multiple middlewares on top of a tx.Handler. The
// middleware order in the variadic arguments is from outer to inner.
//
// Example: Given a base tx.Handler H, and two middlewares A and B, the
// middleware stack:
// ```
// A.pre
//   B.pre
//     H
//   B.post
// A.post
// ```
// is created by calling `ComposeMiddlewares(H, A, B)`.
func ComposeMiddlewares(txHandler tx.Handler, middlewares ...tx.Middleware) tx.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		txHandler = middlewares[i](txHandler)
	}

	return txHandler
}

type TxHandlerOptions struct {
	authmiddleware.TxHandlerOptions

	TXCounterStoreKey    storetypes.StoreKey
	WasmConfig           *wasmTypes.WasmConfig
	IBCKeeper            *ibckeeper.Keeper
	BypassMinFeeMsgTypes []string
}

// NewDefaultTxHandler defines a TxHandler middleware stacks that should work
// for most applications.
func NewDefaultTxHandler(options TxHandlerOptions) (tx.Handler, error) {
	if options.TxDecoder == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "txDecoder is required for middlewares")
	}

	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for middlewares")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for middlewares")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for middlewares")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = middleware.DefaultSigVerificationGasConsumer
	}

	// bez version vvv

	// 	anteDecorators := []sdk.AnteDecorator{
	// 		middleware.NewSetUpContextDecorator(),
	// 		middleware.NewRejectExtensionOptionsDecorator(),
	// 		NewMempoolFeeDecorator(opts.BypassMinFeeMsgTypes),
	// 		middleware.NewValidateBasicDecorator(),
	// 		middleware.NewTxTimeoutHeightDecorator(),
	// 		middleware.NewValidateMemoDecorator(opts.AccountKeeper),
	// 		middleware.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
	// 		middleware.NewDeductFeeDecorator(opts.AccountKeeper, opts.BankKeeper, opts.FeegrantKeeper),
	// 		// SetPubKeyDecorator must be called before all signature verification decorators
	// 		middleware.NewSetPubKeyDecorator(opts.AccountKeeper),
	// 		middleware.NewValidateSigCountDecorator(opts.AccountKeeper),
	// 		middleware.SigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
	// 		middleware.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
	// 		middleware.NewIncrementSequenceDecorator(opts.AccountKeeper),
	// 		ibcante.NewAnteDecorator(opts.IBCkeeper),
	// 	}

	// end bez version ^^^

	return ComposeMiddlewares(
		middleware.NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		// min fee Middleware
		gaiaante.NewMempoolFeeDecorator(options.BypassMinFeeMsgTypes),

		middleware.NewTxDecoderMiddleware(options.TxDecoder),
		// Wasm Middleware
		wasmkeeper.CountTxMiddleware(options.TXCounterStoreKey),
		wasmkeeper.LimitSimulationGasMiddleware(options.WasmConfig.SimulationGasLimit),
		// Set a new GasMeter on sdk.Context.
		//
		// Make sure the Gas middleware is outside of all other middlewares
		// that reads the GasMeter. In our case, the Recovery middleware reads
		// the GasMeter to populate GasInfo.
		middleware.GasTxMiddleware,
		// Recover from panics. Panics outside of this middleware won't be
		// caught, be careful!
		middleware.RecoveryTxMiddleware,
		// Choose which events to index in Tendermint. Make sure no events are
		// emitted outside of this middleware.
		middleware.NewIndexEventsTxMiddleware(options.IndexEvents),
		// Reject all extension options which can optionally be included in the
		// tx.
		middleware.RejectExtensionOptionsMiddleware,
		middleware.MempoolFeeMiddleware,
		middleware.ValidateBasicMiddleware,
		middleware.TxTimeoutHeightMiddleware,
		middleware.ValidateMemoMiddleware(options.AccountKeeper),
		middleware.ConsumeTxSizeGasMiddleware(options.AccountKeeper),
		// Wasm Middleware
		wasmkeeper.CountTxMiddleware(options.TXCounterStoreKey),
		wasmkeeper.LimitSimulationGasMiddleware(options.WasmConfig.SimulationGasLimit),
		// No gas should be consumed in any middleware above in a "post" handler part. See
		// ComposeMiddlewares godoc for details.
		// `DeductFeeMiddleware` and `IncrementSequenceMiddleware` should be put outside of `WithBranchedStore` middleware,
		// so their storage writes are not discarded when tx fails.
		middleware.DeductFeeMiddleware(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		middleware.TxPriorityMiddleware,
		middleware.SetPubKeyMiddleware(options.AccountKeeper),
		middleware.ValidateSigCountMiddleware(options.AccountKeeper),
		middleware.SigGasConsumeMiddleware(options.AccountKeeper, sigGasConsumer),
		middleware.SigVerificationMiddleware(options.AccountKeeper, options.SignModeHandler),
		middleware.IncrementSequenceMiddleware(options.AccountKeeper),
		// Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
		// These kinds of middlewares should be put under this:
		// - Could return error after messages executed successfully.
		// - Storage writes should be discarded together when tx failed.
		middleware.WithBranchedStore,
		// Consume block gas. All middlewares whose gas consumption after their `next` handler
		// should be accounted for, should go below this middleware.
		middleware.ConsumeBlockGasMiddleware,
		middleware.NewTipMiddleware(options.BankKeeper),
		// Ibc v3 middleware
		ibcmiddleware.IBCTxMiddleware(options.IBCKeeper),
	), nil
}
