package ante

import (
	"errors"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"

	tmstrings "github.com/cometbft/cometbft/libs/strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	gaiaerrors "github.com/cosmos/gaia/v18/types/errors"
	"github.com/cosmos/gaia/v18/x/globalfee"
	"github.com/cosmos/gaia/v18/x/globalfee/types"
)

// FeeWithBypassDecorator checks if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee, and the fee denom should be in the global fees' denoms.
//
//
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
// If the tx msg type is one of the bypass msg types, the tx is valid even if the min fee is lower than normally required.
// If the bypass tx still carries fees, the fee denom should be the same as global fee required.

var _ sdk.AnteDecorator = FeeDecorator{}

type FeeDecorator struct {
	GlobalMinFeeParamSource globalfee.ParamSource
	StakingKeeper           *stakingkeeper.Keeper
}

func NewFeeDecorator(globalfeeSubspace paramtypes.Subspace, sk *stakingkeeper.Keeper) FeeDecorator {
	if !globalfeeSubspace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	return FeeDecorator{
		GlobalMinFeeParamSource: globalfeeSubspace,
		StakingKeeper:           sk,
	}
}

// AnteHandle implements the AnteDecorator interface
func (mfd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(gaiaerrors.ErrTxDecode, "Tx must implement the sdk.FeeTx interface")
	}

	// Do not check minimum-gas-prices and global fees during simulations
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Get the required fees according to the CheckTx or DeliverTx modes
	minGasPrices, err := mfd.GetTxGasPrices(ctx)
	if err != nil {
		return ctx, err
	}

	// Sort fee tx's coins, zero coins in feeCoins are already removed
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()
	// If the feeCoins pass the denoms check, check they are bypass-msg types.
	//
	// Bypass min fee requires:
	// 	- the tx contains only message types that can bypass the minimum fee,
	//	see BypassMinFeeMsgTypes;
	//	- the total gas limit per message does not exceed MaxTotalBypassMinFeeMsgGasUsage,
	//	i.e., totalGas <=  MaxTotalBypassMinFeeMsgGasUsage
	// Otherwise, minimum fees and global fees are checked to prevent spam.
	maxTotalBypassMinFeeMsgGasUsage := mfd.GetMaxTotalBypassMinFeeMsgGasUsage(ctx)
	doesNotExceedMaxGasUsage := gas <= maxTotalBypassMinFeeMsgGasUsage
	allBypassMsgs := mfd.ContainsOnlyBypassMinFeeMsgs(ctx, msgs)
	allowedToBypassMinFee := allBypassMsgs && doesNotExceedMaxGasUsage

	// If bypass msg, add value to context so feeabs ante can check it
	if allowedToBypassMinFee {
		return next(ctx.WithValue(feeabstypes.ByPassMsgKey{}, true).WithValue(feeabstypes.GlobalFeeKey{}, true), tx, simulate)
	}

	if allBypassMsgs && !doesNotExceedMaxGasUsage {
		return next(ctx.WithMinGasPrices(minGasPrices).WithValue(feeabstypes.ByPassExceedMaxGasUsageKey{}, true).WithValue(feeabstypes.GlobalFeeKey{}, true), tx, simulate)
	}

	return next(ctx.WithMinGasPrices(minGasPrices).WithValue(feeabstypes.GlobalFeeKey{}, true), tx, simulate)
}

// GetTxFeeRequired returns the required fees for the given FeeTx.
// In case the FeeTx's mode is CheckTx, it returns the combined requirements
// of local min gas prices and global fees. Otherwise, in DeliverTx, it returns the global fee.
func (mfd FeeDecorator) GetTxFeeRequired(ctx sdk.Context, tx sdk.FeeTx) (sdk.Coins, error) {
	// Get required global fee min gas prices
	// Note that it should never be empty since its default value is set to coin={"StakingBondDenom", 0}
	globalFees, err := mfd.GetGlobalFee(ctx, tx)
	if err != nil {
		return sdk.Coins{}, err
	}

	// In DeliverTx, the global fee min gas prices are the only tx fee requirements.
	if !ctx.IsCheckTx() {
		return globalFees, nil
	}

	// In CheckTx mode, the local and global fee min gas prices are combined
	// to form the tx fee requirements

	// Get local minimum-gas-prices
	localFees := GetMinGasPrice(ctx, int64(tx.GetGas()))
	c, err := CombinedFeeRequirement(globalFees, localFees)

	// Return combined fee requirements
	return c, err
}

// GetGlobalFee returns the global fees for a given fee tx's gas
// (might also return 0denom if globalMinGasPrice is 0)
// sorted in ascending order.
// Note that ParamStoreKeyMinGasPrices type requires coins sorted.
func (mfd FeeDecorator) GetGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) (sdk.Coins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	if mfd.GlobalMinFeeParamSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		mfd.GlobalMinFeeParamSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	}
	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = mfd.DefaultZeroGlobalFee(ctx)
		if err != nil {
			return sdk.Coins{}, err
		}
	}
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredGlobalFees.Sort(), nil
}

// DefaultZeroGlobalFee returns a zero coin with the staking module bond denom
func (mfd FeeDecorator) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := mfd.StakingKeeper.BondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (mfd FeeDecorator) ContainsOnlyBypassMinFeeMsgs(ctx sdk.Context, msgs []sdk.Msg) bool {
	bypassMsgTypes := mfd.GetBypassMsgTypes(ctx)

	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), bypassMsgTypes) {
			continue
		}
		return false
	}

	return true
}

func (mfd FeeDecorator) GetBypassMsgTypes(ctx sdk.Context) (res []string) {
	if mfd.GlobalMinFeeParamSource.Has(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes) {
		mfd.GlobalMinFeeParamSource.Get(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes, &res)
	}

	return
}

func (mfd FeeDecorator) GetMaxTotalBypassMinFeeMsgGasUsage(ctx sdk.Context) (res uint64) {
	if mfd.GlobalMinFeeParamSource.Has(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
		mfd.GlobalMinFeeParamSource.Get(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &res)
	}

	return
}

// GetMinGasPrice returns a nodes's local minimum gas prices
// fees given a gas limit
func GetMinGasPrice(ctx sdk.Context, gasLimit int64) sdk.Coins {
	minGasPrices := ctx.MinGasPrices()
	// special case: if minGasPrices=[], requiredFees=[]
	if minGasPrices.IsZero() {
		return sdk.Coins{}
	}

	requiredFees := make(sdk.Coins, len(minGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(gasLimit)
	for i, gp := range minGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredFees.Sort()
}

// GetTxGasPrices returns the min-gas-prices for the given FeeTx.
// In case the FeeTx's mode is CheckTx, it returns the combined requirements
// of local min gas prices and global fees. Otherwise, in DeliverTx, it returns the global fee.
func (mfd FeeDecorator) GetTxGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	// Get required global fee min gas prices
	// Note that it should never be empty since its default value is set to coin={"StakingBondDenom", 0}
	globalGasPrices, err := mfd.GetGlobalGasPrices(ctx)
	if err != nil {
		return sdk.DecCoins{}, err
	}

	// In DeliverTx, the global fee min gas prices are the only tx fee requirements.
	if !ctx.IsCheckTx() {
		return globalGasPrices, nil
	}

	// In CheckTx mode, the local and global fee min gas prices are combined
	// to form the tx fee requirements

	// Get local minimum-gas-prices
	localGasPrices := ctx.MinGasPrices().Sort()

	// Return combined min-gas-prices
	return CombinedMinGasPrices(globalGasPrices, localGasPrices)
}

// GetGlobalGasPrices returns the global min-gas-prices
// sorted in ascending order.
// Note that ParamStoreKeyMinGasPrices type requires coins sorted.
func (mfd FeeDecorator) GetGlobalGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	if mfd.GlobalMinFeeParamSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		mfd.GlobalMinFeeParamSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	}
	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = mfd.DefaultZeroGlobalFee(ctx)
		if err != nil {
			return sdk.DecCoins{}, err
		}
	}
	return globalMinGasPrices.Sort(), nil
}

// CombinedMinGasPrices returns the globalfee's gas-prices and local min_gas_price combined and sorted.
// Both globalfee's gas-prices and local min_gas_price must be valid, but CombinedMinGasPrices
// does not validate them, so it may return 0denom.
// if globalfee is empty, CombinedMinGasPrices return sdk.DecCoins{}
func CombinedMinGasPrices(globalGasPrices, minGasPrices sdk.DecCoins) (sdk.DecCoins, error) {
	// global fees should never be empty
	// since it has a default value using the staking module's bond denom
	if len(globalGasPrices) == 0 {
		return sdk.DecCoins{}, errorsmod.Wrapf(gaiaerrors.ErrNotFound, "global fee cannot be empty")
	}

	// empty min_gas_price
	if len(minGasPrices) == 0 {
		return globalGasPrices, nil
	}

	// if min_gas_price denom is in globalfee, and the amount is higher than globalfee, add min_gas_price to allFees
	var allFees sdk.DecCoins
	for _, fee := range globalGasPrices {
		// min_gas_price denom in global fee
		ok, c := FindDecCoins(minGasPrices, fee.Denom)
		if ok && c.Amount.GT(fee.Amount) {
			allFees = append(allFees, c)
		} else {
			allFees = append(allFees, fee)
		}
	}

	return allFees.Sort(), nil
}

// FindDecCoins Clone from Find() func above for DecCoins
func FindDecCoins(coins sdk.DecCoins, denom string) (bool, sdk.DecCoin) {
	switch len(coins) {
	case 0:
		return false, sdk.DecCoin{}

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return true, coin
		}
		return false, sdk.DecCoin{}

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]
		switch {
		case denom < coin.Denom:
			return FindDecCoins(coins[:midIdx], denom)
		case denom == coin.Denom:
			return true, coin
		default:
			return FindDecCoins(coins[midIdx+1:], denom)
		}
	}
}
