package ante

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v9/x/globalfee/types"

	tmstrings "github.com/tendermint/tendermint/libs/strings"

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

	// sort fee tx's coins, feeCoins' zero coins are already removed
	feeCoins := feeTx.GetFee().Sort()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Get required Global Fee and min gas price
	globalFeesAll, err := mfd.getGlobalFees(ctx, feeTx)
	if err != nil {
		return ctx, err
	}

	// get minimum-gas-prices from app.toml or cli flag
	localFees := GetMinGasPrice(ctx, int64(feeTx.GetGas()))

	combinedFeeRequirement := CombinedFeeRequirement(globalFeesAll, localFees)
	if len(combinedFeeRequirement) == 0 {
		// todo return err
		return ctx, nil
	}
	nonZeroCoinFeesReq, zeroCoinFeesDenomReq := splitFees(combinedFeeRequirement)

	// feeCoinsNoZeroDenom is feeCoins after removing the coins whose denom is zero coins' denom in globalfees
	// e.g. feeCoins=[1atom,2photon], globalfee=[0atom,1photon,1quark], then feeCoinsNoZeroDenom = [2photon]
	// feeCoinsNoZeroDenom are used to check if the fees are meet the requirement imposed by combinedNonZeroFees
	// when feeCoins does not contain zero coins'denoms in combinedFeeRequirement
	feeCoinsNoZeroDenom, feeCoinsZeroDenom := SplitCoinsByDenoms(feeCoins, zeroCoinFeesDenomReq)

	// Check that the fees are in expected denominations.
	// if len(feeCoinsNoZeroDenom) = 0, DenomsSubsetOf returns true
	// if len(feeCoinsNoZeroDenom) != 0 && len(nonZeroCoinFeesReq) = 0, return false
	if !feeCoinsNoZeroDenom.DenomsSubsetOf(nonZeroCoinFeesReq) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fee is not a subset of required fees; got %s, required: %s", feeCoins, combinedFeeRequirement)
	}

	// Accept zero fee transactions only if both of the following statements are true:
	//
	// 	- the tx contains only message types that can bypass the minimum fee,
	//	see BypassMinFeeMsgTypes;
	//	- the total gas limit per message does not exceed MaxTotalBypassMinFeeMsgGasUsage,
	//	i.e., totalGas <=  MaxTotalBypassMinFeeMsgGasUsage
	//
	// Otherwise, minimum fees and global fees are checked to prevent spam.
	doesNotExceedMaxGasUsage := gas <= mfd.MaxTotalBypassMinFeeMsgGasUsage
	allowedToBypassMinFee := mfd.ContainsOnlyBypassMinFeeMsgs(msgs) && doesNotExceedMaxGasUsage

	// Either the transaction contains at least one message of a type
	// that cannot bypass the minimum fee or the total gas limit exceeds
	// the imposed threshold. As a result, besides check the fees are in
	// expected denominations, check the amounts are greater or equal than
	// the expected amounts.

	// only check feeCoinsNoZeroDenom has coins IsAnyGTE than nonZeroCoinFeesReq
	// when feeCoins does not contain denoms of zero denoms in combinedFeeRequirement
	if !allowedToBypassMinFee && len(feeCoinsZeroDenom) == 0 {
		// This is for dealing special case when feeCoins=[]
		if len(feeCoins) == 0 && len(zeroCoinFeesDenomReq) != 0 {
			return next(ctx, tx, simulate)
		}

		// Check that the amounts of the fees are greater or equal than
		// the expected amounts, i.e., at least one feeCoin amount must
		// be greater or equal to one of the combined required fees.

		// if len(feeCoinsNoZeroDenom) = 0, return false
		// if len(nonZeroCoinFeesReq) = 0, return false (this situation should not happen
		// because when nonZeroCoinFeesReq empty, the denom check already failed before)
		if !feeCoinsNoZeroDenom.IsAnyGTE(nonZeroCoinFeesReq) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, combinedFeeRequirement)
		}
	}

	return next(ctx, tx, simulate)
}

// ParamStoreKeyMinGasPrices type require coins sorted. getGlobalFee will also return sorted coins (might return 0denom if globalMinGasPrice is 0)
func (mfd FeeDecorator) getGlobalFees(ctx sdk.Context, feeTx sdk.FeeTx) (sdk.Coins, error) {
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

// ContainsOnlyBypassMinFeeMsgs returns true if all the given msgs type are listed
// in the BypassMinFeeMsgTypes of the FeeDecorator.
func (mfd FeeDecorator) ContainsOnlyBypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}
		return false
	}

	return true
}

// GetMinGasPrice returns the validator's minimum gas prices
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
