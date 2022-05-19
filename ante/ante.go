package ante

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"math"
	// "github.com/cosmos/cosmos-sdk/x/auth/ante"
	// ibcante "github.com/cosmos/ibc-go/v3/modules/core/ante"
	// ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
)

// // HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// // channel keeper.
// type HandlerOptions struct {
// 	ante.HandlerOptions

// 	IBCkeeper            *ibckeeper.Keeper
// 	BypassMinFeeMsgTypes []string
// }

// func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
// 	if opts.AccountKeeper == nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
// 	}
// 	if opts.BankKeeper == nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
// 	}
// 	if opts.SignModeHandler == nil {
// 		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
// 	}

// 	var sigGasConsumer = opts.SigGasConsumer
// 	if sigGasConsumer == nil {
// 		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
// 	}

// 	anteDecorators := []sdk.AnteDecorator{
// 		ante.NewSetUpContextDecorator(),
// 		ante.NewRejectExtensionOptionsDecorator(),
// 		NewMempoolFeeDecorator(opts.BypassMinFeeMsgTypes),
// 		ante.NewValidateBasicDecorator(),
// 		ante.NewTxTimeoutHeightDecorator(),
// 		ante.NewValidateMemoDecorator(opts.AccountKeeper),
// 		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
// 		ante.NewDeductFeeDecorator(opts.AccountKeeper, opts.BankKeeper, opts.FeegrantKeeper),
// 		// SetPubKeyDecorator must be called before all signature verification decorators
// 		ante.NewSetPubKeyDecorator(opts.AccountKeeper),
// 		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
// 		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
// 		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
// 		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
// 		ibcante.NewAnteDecorator(opts.IBCkeeper),
// 	}

// TxHandlerOptions extend the SDK's TxHandlerOptions options by requiring the IBC
// channel keeper and bypass-min-fee types
type TxHandlerOptions struct {
	// ante.HandlerOptions
	middleware.TxHandlerOptions
	// IBCkeeper            *ibckeeper.Keeper
	BypassMinFeeMsgTypes []string
}


func NewTxHandler(options TxHandlerOptions) (tx.Handler, error) {
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

	var extensionOptionChecker = options.ExtensionOptionChecker
	if extensionOptionChecker == nil {
		extensionOptionChecker = rejectExtensionOption

	}

	var txFeeChecker = options.TxFeeChecker
	if txFeeChecker == nil {
		txFeeChecker = checkTxFeeWithValidatorMinGasPrices
	}

	return middleware.ComposeMiddlewares(
		middleware.NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
		middleware.NewTxDecoderMiddleware(options.TxDecoder),
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
		// Reject all extension options other than the ones needed by the feemarket.
		middleware.NewExtensionOptionsMiddleware(extensionOptionChecker),
		middleware.ValidateBasicMiddleware,
		middleware.TxTimeoutHeightMiddleware,
		middleware.ValidateMemoMiddleware(options.AccountKeeper),
		middleware.ConsumeTxSizeGasMiddleware(options.AccountKeeper),
		// No gas should be consumed in any middleware above in a "post" handler part. See
		// ComposeMiddlewares godoc for details.
		// `DeductFeeMiddleware` and `IncrementSequenceMiddleware` should be put outside of `WithBranchedStore` middleware,
		// so their storage writes are not discarded when tx fails.
		middleware.DeductFeeMiddleware(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, txFeeChecker),
		middleware.SetPubKeyMiddleware(options.AccountKeeper),
		middleware.ValidateSigCountMiddleware(options.AccountKeeper),
		middleware.SigGasConsumeMiddleware(options.AccountKeeper, sigGasConsumer),
		middleware.SigVerificationMiddleware(options.AccountKeeper, options.SignModeHandler),
		middleware.IncrementSequenceMiddleware(options.AccountKeeper),
		// Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
		// These kinds of middlewares should be put under this:
		// - Could return error after messages executed succesfully.
		// - Storage writes should be discarded together when tx failed.
		middleware.WithBranchedStore,
		// Consume block gas. All middlewares whose gas consumption after their `next` handler
		// should be accounted for, should go below this middleware.
		middleware.ConsumeBlockGasMiddleware,
		middleware.NewTipMiddleware(options.BankKeeper),
		// ibcante.NewAnteDecorator(opts.IBCkeeper),
	), nil
}


// helper functions
// rejectExtensionOption is the default extension check that reject all tx
// extensions.
func rejectExtensionOption(*codectypes.Any) bool {
	return false
}


// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, can the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// This is only for local mempool purposes, if this is a DeliverTx, the `MinGasPrices` should be zero.
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	priority := getTxPriority(feeCoins)
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func getTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
