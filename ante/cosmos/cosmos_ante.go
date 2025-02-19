package cosmos

import (
	"github.com/cosmos/gaia/v23/ante/gov"
	"github.com/cosmos/gaia/v23/ante/handler_options"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"

	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

// UseFeeMarketDecorator to make the integration testing easier: we can switch off its ante and post decorators with this flag
var UseFeeMarketDecorator = true

func NewAnteHandler(opts handler_options.HandlerOptions) sdk.AnteHandler {
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
		gov.NewGovVoteDecorator(opts.Codec, opts.StakingKeeper),
		gov.NewGovExpeditedProposalsDecorator(opts.Codec),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),         //todo: add ethsekp256k1
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler), //todo: verify ethereum signatures
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(opts.IBCkeeper),
	}

	if UseFeeMarketDecorator {
		anteDecorators = append(anteDecorators,
			feemarketante.NewFeeMarketCheckDecorator(
				opts.AccountKeeper,
				opts.BankKeeper,
				opts.FeegrantKeeper,
				opts.FeeMarketKeeper,
				ante.NewDeductFeeDecorator(
					opts.AccountKeeper,
					opts.BankKeeper,
					opts.FeegrantKeeper,
					opts.TxFeeChecker)))
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
