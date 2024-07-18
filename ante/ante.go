package ante

import (
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"

	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	corestoretypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiaerrors "github.com/cosmos/gaia/v19/types/errors"
)

// UseFeeMarketDecorator to make the integration testing easier: we can switch off its ante and post decorators with this flag
var UseFeeMarketDecorator = true

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ante.HandlerOptions
	Codec                 codec.BinaryCodec
	IBCkeeper             *ibckeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	FeeMarketKeeper       *feemarketkeeper.Keeper
	TxFeeChecker          ante.TxFeeChecker
	TXCounterStoreService corestoretypes.KVStoreService
	WasmConfig            *wasmtypes.WasmConfig
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
	if opts.FeeMarketKeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrLogic, "FeeMarket keeper is required for AnteHandler")
	}

	if opts.StakingKeeper == nil {
		return nil, errorsmod.Wrap(gaiaerrors.ErrNotFound, "staking param store is required for AnteHandler")
	}

	sigGasConsumer := opts.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),                                               // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(opts.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(opts.TXCounterStoreService),
		ante.NewExtensionOptionsDecorator(opts.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(opts.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		NewGovVoteDecorator(opts.Codec, opts.StakingKeeper),
		NewGovExpeditedProposalsDecorator(opts.Codec),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(opts.IBCkeeper),
	}

	if UseFeeMarketDecorator {
		anteDecorators = append(anteDecorators,
			feemarketante.NewFeeMarketCheckDecorator(
				opts.FeeMarketKeeper,
				ante.NewDeductFeeDecorator(
					opts.AccountKeeper,
					opts.BankKeeper,
					opts.FeegrantKeeper,
					opts.TxFeeChecker)))
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
