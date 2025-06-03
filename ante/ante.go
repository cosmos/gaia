package ante

import (
	evmcosmosante "github.com/cosmos/evm/ante/cosmos"
	evmante "github.com/cosmos/evm/ante/evm"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	ibcante "github.com/cosmos/ibc-go/v10/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	corestoretypes "cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiaerrors "github.com/cosmos/gaia/v25/types/errors"
)

// HandlerOptions extend the SDK's AnteHandler options by requiring the IBC
// channel keeper.
type HandlerOptions struct {
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         ante.FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error

	AccountKeeper         anteinterfaces.AccountKeeper
	BankKeeper            anteinterfaces.BankKeeper
	Codec                 codec.BinaryCodec
	IBCkeeper             *ibckeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	TxFeeChecker          ante.TxFeeChecker
	TXCounterStoreService corestoretypes.KVStoreService
	WasmConfig            *wasmtypes.NodeConfig

	// TODO -- isn't this in module params in skip-mev/feemarket?
	MaxTxGasWanted  uint64
	FeeMarketKeeper anteinterfaces.FeeMarketKeeper
	EvmKeeper       anteinterfaces.EVMKeeper
}

func (opts *HandlerOptions) Validate() error {
	if opts.AccountKeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if opts.BankKeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if opts.SignModeHandler == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "sign mode handler is required for AnteHandler")
	}
	if opts.IBCkeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "IBC keeper is required for AnteHandler")
	}
	if opts.FeeMarketKeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "FeeMarket keeper is required for AnteHandler")
	}
	if opts.EvmKeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrLogic, "EvmKeeper keeper is required for AnteHandler")
	}
	if opts.StakingKeeper == nil {
		return errorsmod.Wrap(gaiaerrors.ErrNotFound, "staking param store is required for AnteHandler")
	}
	return nil
}

func NewEvmAnteHandler(opts HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		evmante.NewEVMMonoDecorator(
			opts.AccountKeeper,
			opts.FeeMarketKeeper,
			opts.EvmKeeper,
			opts.MaxTxGasWanted,
		),
	)
}

func NewCosmosAnteHandler(opts HandlerOptions) sdk.AnteHandler {
	sigGasConsumer := opts.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		evmcosmosante.NewRejectMessagesDecorator(), // reject MsgEthereumTxs
		evmcosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			// TODO -- do we care about vesting accounts here?
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),
		ante.NewSetUpContextDecorator(),                                               // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(opts.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(opts.TXCounterStoreService),
		ante.NewExtensionOptionsDecorator(opts.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(opts.AccountKeeper),
		// TODO -- can I disable feemarket anymore?
		evmcosmosante.NewMinGasPriceDecorator(opts.FeeMarketKeeper, opts.EvmKeeper),
		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		NewGovVoteDecorator(opts.Codec, opts.StakingKeeper),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(opts.IBCkeeper),
		evmante.NewGasWantedDecorator(opts.EvmKeeper, opts.FeeMarketKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}
	return func(ctx sdk.Context, tx sdk.Tx, sim bool) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		txWithExtensions, ok := tx.(ante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/cosmos.evm.vm.v1.ExtensionOptionsEthereumTx":
					// handle as *evmtypes.MsgEthereumTx
					anteHandler = NewEvmAnteHandler(options)
				case "/cosmos.evm.types.v1.ExtensionOptionDynamicFeeTx":
					// cosmos-sdk tx with dynamic fee extension
					anteHandler = NewCosmosAnteHandler(options)
				default:
					return ctx, errorsmod.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// handle as totally normal Cosmos SDK tx
		switch tx.(type) {
		case sdk.Tx:
			anteHandler = NewCosmosAnteHandler(options)
		default:
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}, nil
}
