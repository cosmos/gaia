package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcante "github.com/cosmos/ibc-go/v3/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	gaiafeeante "github.com/cosmos/gaia/v9/x/globalfee/ante"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	IBCkeeper            *ibckeeper.Keeper
	BypassMinFeeMsgTypes []string
	GlobalFeeSubspace    paramtypes.Subspace
	StakingSubspace      paramtypes.Subspace
}

func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
	if opts.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if opts.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if opts.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for AnteHandler")
	}
	if opts.IBCkeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "IBC keeper is required for AnteHandler")
	}
	if opts.GlobalFeeSubspace.Name() == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "globalfee param store is required for AnteHandler")
	}
	if opts.StakingSubspace.Name() == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, "staking param store is required for AnteHandler")
	}

	sigGasConsumer := opts.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewRejectExtensionOptionsDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(opts.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		gaiafeeante.NewFeeDecorator(opts.BypassMinFeeMsgTypes, opts.GlobalFeeSubspace, opts.StakingSubspace, 200_000),
		ante.NewDeductFeeDecorator(opts.AccountKeeper, opts.BankKeeper, opts.FeegrantKeeper),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewAnteDecorator(opts.IBCkeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
