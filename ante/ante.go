package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
	ibcante "github.com/cosmos/ibc-go/v3/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
)

// TODO: Add this back to app.go

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	middleware.TxHandlerOptions

	IBCkeeper            *ibckeeper.Keeper
	BypassMinFeeMsgTypes []string
}

func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
	if opts.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if opts.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if opts.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	var sigGasConsumer = opts.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = middleware.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		middleware.NewSetUpContextDecorator(),
		middleware.NewRejectExtensionOptionsDecorator(),
		NewMempoolFeeDecorator(opts.BypassMinFeeMsgTypes),
		middleware.NewValidateBasicDecorator(),
		middleware.NewTxTimeoutHeightDecorator(),
		middleware.NewValidateMemoDecorator(opts.AccountKeeper),
		middleware.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		middleware.NewDeductFeeDecorator(opts.AccountKeeper, opts.BankKeeper, opts.FeegrantKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		middleware.NewSetPubKeyDecorator(opts.AccountKeeper),
		middleware.NewValidateSigCountDecorator(opts.AccountKeeper),
		middleware.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		middleware.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		middleware.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewAnteDecorator(opts.IBCkeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
