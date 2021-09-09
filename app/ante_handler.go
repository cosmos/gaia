package gaia

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    "github.com/cosmos/cosmos-sdk/types/tx/signing"
    authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
    authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
    channelkeeper "github.com/cosmos/ibc-go/modules/core/04-channel/keeper"
    ibcante "github.com/cosmos/ibc-go/modules/core/ante"
)

// redefine handlerOptions to add ibc-go channelKeeper
type HandlerOptions struct {
    AccountKeeper  authante.AccountKeeper
    BankKeeper     authtypes.BankKeeper
    FeegrantKeeper authante.FeegrantKeeper
    // add ibc-go channelKeeper
    Channelkeeper   channelkeeper.Keeper
    SignModeHandler authsigning.SignModeHandler
    SigGasConsumer  func(meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params) error
}

// redefine NewANteHandler to add ibc anteDecorator
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
    if options.AccountKeeper == nil {
        return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
    }

    if options.BankKeeper == nil {
        return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
    }

    //todo Check each field in channelKeeper is not nil ?

    if options.SignModeHandler == nil {
        return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
    }

    var sigGasConsumer = options.SigGasConsumer
    if sigGasConsumer == nil {
        sigGasConsumer = authante.DefaultSigVerificationGasConsumer
    }

    anteDecorators := []sdk.AnteDecorator{
        authante.NewRejectExtensionOptionsDecorator(),
        authante.NewMempoolFeeDecorator(),
        authante.NewValidateBasicDecorator(),
        authante.NewTxTimeoutHeightDecorator(),
        authante.NewValidateMemoDecorator(options.AccountKeeper),
        authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
        authante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
        authante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
        authante.NewValidateSigCountDecorator(options.AccountKeeper),
        authante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
        authante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
        authante.NewIncrementSequenceDecorator(options.AccountKeeper),
        // todo check ibcante is at the right order
        ibcante.NewAnteDecorator(options.Channelkeeper),
    }

    return sdk.ChainAnteDecorators(anteDecorators...), nil
}
