package handler_options

import (
	corestoretypes "cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	feemarketante "github.com/skip-mev/feemarket/x/feemarket/ante"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
)

type HandlerOptions struct {
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         ante.FeegrantKeeper
	SignModeHandler        *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error

	AccountKeeper         feemarketante.AccountKeeper
	BankKeeper            feemarketante.BankKeeper
	Codec                 codec.Codec
	IBCkeeper             *ibckeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	FeeMarketKeeper       *feemarketkeeper.Keeper
	TxFeeChecker          ante.TxFeeChecker
	TXCounterStoreService corestoretypes.KVStoreService
	WasmConfig            *wasmtypes.WasmConfig
}
