package ante

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v9/x/globalfee/types"

	"github.com/cosmos/gaia/v9/x/globalfee"
)

// FeeWithBypassDecorator checks if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee, and the fee denom should be in the global fees' denoms.
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
// If the tx msg type is one of the bypass msg types, the tx is valid even if the min fee is lower than normally required.
// If the bypass tx still carries fees, the fee denom should be the same as global fee required.

var _ sdk.AnteDecorator = FeeDecorator{}

type FeeDecorator struct {
	BypassMinFeeMsgTypes            []string
	GlobalMinFee                    globalfee.ParamSource
	StakingSubspace                 paramtypes.Subspace
	MaxTotalBypassMinFeeMsgGasUsage uint64
}

func NewFeeDecorator(bypassMsgTypes []string, globalfeeSubspace, stakingSubspace paramtypes.Subspace, maxTotalBypassMinFeeMsgGasUsage uint64) FeeDecorator {
	if !globalfeeSubspace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	if !stakingSubspace.HasKeyTable() {
		panic("staking paramspace was not set up via module")
	}

	return FeeDecorator{
		BypassMinFeeMsgTypes:            bypassMsgTypes,
		GlobalMinFee:                    globalfeeSubspace,
		StakingSubspace:                 stakingSubspace,
		MaxTotalBypassMinFeeMsgGasUsage: maxTotalBypassMinFeeMsgGasUsage,
	}
}

// AnteHandle implements the AnteDecorator interface
func (mfd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// please note: after parsing feeflag, the zero fees are removed already
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Only check for minimum fees and global fee if the execution mode is CheckTx
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	// sort fee tx's coins
	feeCoins := feeTx.GetFee().Sort()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Get required Global Fee and min gas price
	requiredGlobalFees, err := mfd.getGlobalFee(ctx, feeTx)
	if err != nil {
		panic(err)
	}
	requiredFees := getMinGasPrice(ctx, feeTx)

	// Accept zero fee transactions only if both of the following statements are true:
	//
	// 	- the tx contains only message types that can bypass the minimum fee,
	//	see BypassMinFeeMsgTypes;
	//	- the total gas limit per message does not exceed MaxTotalBypassMinFeeMsgGasUsage,
	//	i.e., totalGas <= MaxTotalBypassMinFeeMsgGasUsage
	// Otherwise, minimum fees and global fees are checked to prevent spam.
	doesNotExceedMaxGasUsage := gas <= mfd.MaxTotalBypassMinFeeMsgGasUsage
	allowedToBypassMinFee := mfd.containsOnlyBypassMinFeeMsgs(msgs) && doesNotExceedMaxGasUsage

	if allowedToBypassMinFee {
		// Transactions with zero fees are accepted
		if len(feeCoins) == 0 {
			return next(ctx, tx, simulate)
		}
		// If the transaction fee is non-zero, then check that the fees are in
		// expected denominations.
		if !DenomsSubsetOfIncludingZero(feeCoins, requiredGlobalFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fees denom is wrong; got: %s required: %s", feeCoins, requiredGlobalFees)
		}
	}

	// Either the transaction contains at least on message of a type
	// that cannot bypass the minimum fee or the total gas limit exceeds
	// the imposed threshold. As a result, check that the fees are in
	// expected denominations and the amounts are greater or equal than
	// the expected amounts.

	allFees := CombinedFeeRequirement(requiredGlobalFees, requiredFees)

	// Check that the fees are in expected denominations. Note that a zero fee
	// is accepted if the global fee has an entry with a zero amount, e.g., 0uatoms.
	if !DenomsSubsetOfIncludingZero(feeCoins, allFees) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fee is not a subset of required fees; got %s, required: %s", feeCoins, allFees)
	}
	// Check that the amounts of the fees are greater or equal than
	// the expected amounts, i.e., at least one feeCoin amount must
	// be greater or equal to one of the combined required fees.
	if !IsAnyGTEIncludingZero(feeCoins, allFees) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, allFees)
	}

	return next(ctx, tx, simulate)
}

// getGlobalFee returns the global fees for a given fee tx's gas (might also returns 0denom if globalMinGasPrice is 0)
// sorted in ascending order.
// Note that ParamStoreKeyMinGasPrices type requires coins sorted.
func (mfd FeeDecorator) getGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) (sdk.Coins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	if mfd.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		mfd.GlobalMinFee.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
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

func (mfd FeeDecorator) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := mfd.getBondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (mfd FeeDecorator) getBondDenom(ctx sdk.Context) string {
	var bondDenom string
	if mfd.StakingSubspace.Has(ctx, stakingtypes.KeyBondDenom) {
		mfd.StakingSubspace.Get(ctx, stakingtypes.KeyBondDenom, &bondDenom)
	}

	return bondDenom
}
