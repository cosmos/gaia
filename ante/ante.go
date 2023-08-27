package ante

import (
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	gaiaerrors "github.com/cosmos/gaia/v13/types/errors"
	gaiafeeante "github.com/cosmos/gaia/v13/x/globalfee/ante"

	feeabsante "github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/ante"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v4/x/feeabs/keeper"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	Codec             codec.BinaryCodec
	GovKeeper         *govkeeper.Keeper
	IBCkeeper         *ibckeeper.Keeper
	FeeAbskeeper      feeabskeeper.Keeper
	GlobalFeeSubspace paramtypes.Subspace
	StakingSubspace   paramtypes.Subspace
}

func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
	if opts.AccountKeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if opts.BankKeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if opts.SignModeHandler == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "sign mode handler is required for AnteHandler")
	}
	if opts.IBCkeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "IBC keeper is required for AnteHandler")
	}
	if opts.GlobalFeeSubspace.Name() == "" {
		return nil, errorsmod.Wrap(gaiaerrors.ErrNotFound, "globalfee param store is required for AnteHandler")
	}
	if opts.StakingSubspace.Name() == "" {
		return nil, errorsmod.Wrap(gaiaerrors.ErrNotFound, "staking param store is required for AnteHandler")
	}
	if opts.GovKeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "gov keeper is required for AnteHandler")
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
		NewGovPreventSpamDecorator(opts.Codec, opts.GovKeeper),
		gaiafeeante.NewFeeDecorator(opts.GlobalFeeSubspace, opts.StakingSubspace),
		feeabsante.NewFeeAbstrationMempoolFeeDecorator(opts.FeeAbskeeper),
		feeabsante.NewFeeAbstractionDeductFeeDecorate(opts.AccountKeeper, opts.BankKeeper, opts.FeeAbskeeper, opts.FeegrantKeeper),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewAnteDecorator(opts.IBCkeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
