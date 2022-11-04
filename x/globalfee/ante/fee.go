package ante

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v8/x/globalfee/types"

	"github.com/cosmos/gaia/v8/x/globalfee"
)

const maxBypassMinFeeMsgGasUsage uint64 = 200_000

// FeeWithBypassDecorator will check if the transaction's fee is at least as large
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
	BypassMinFeeMsgTypes []string
	GlobalMinFee         globalfee.ParamSource
	StakingSubspace      paramtypes.Subspace
}

func NewFeeDecorator(bypassMsgTypes []string, globalfeeSubspace, stakingSubspace paramtypes.Subspace) FeeDecorator {
	if !globalfeeSubspace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	if !stakingSubspace.HasKeyTable() {
		panic("staking paramspace was not set up via module")
	}

	return FeeDecorator{
		BypassMinFeeMsgTypes: bypassMsgTypes,
		GlobalMinFee:         globalfeeSubspace,
		StakingSubspace:      stakingSubspace,
	}
}

func (mfd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// please note: after parsing feeflag, the zero fees are removed already
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	feeCoins := feeTx.GetFee().Sort()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Only check for minimum fees and global fee if the execution mode is CheckTx and the tx does
	// not contain operator configured bypass messages. If the tx does contain
	// operator configured bypass messages only, it's total gas must be less than
	// or equal to a constant, otherwise minimum fees and global fees are checked to prevent spam.
	containsOnlyBypassMinFeeMsgs := mfd.bypassMinFeeMsgs(msgs)
	doesNotExceedMaxGasUsage := gas <= uint64(len(msgs))*maxBypassMinFeeMsgGasUsage
	allowedToBypassMinFee := containsOnlyBypassMinFeeMsgs && doesNotExceedMaxGasUsage

	var allFees sdk.Coins
	requiredFees := getMinGasPrice(ctx, feeTx)

	if ctx.IsCheckTx() && !simulate && !allowedToBypassMinFee {

		requiredGlobalFees, err := mfd.getGlobalFee(ctx, feeTx)
		if err != nil {
			panic(err)
		}
		allFees = CombinedFeeRequirement(requiredGlobalFees, requiredFees)

		// this is to ban 1stake passing if the globalfee is 1photon or 0photon
		// if feeCoins=[] and requiredGlobalFees has 0denom coins then it should pass.
		if !DenomsSubsetOfIncludingZero(feeCoins, allFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fee is not a subset of required fees; got %s, required: %s", feeCoins, allFees)
		}
		// At least one feeCoin amount must be GTE to one of the requiredGlobalFees
		if !IsAnyGTEIncludingZero(feeCoins, allFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, allFees)
		}
	}

	// when the tx is bypass msg type, still need to check the denom is not some unknown denom
	if ctx.IsCheckTx() && !simulate && allowedToBypassMinFee {
		requiredGlobalFees, err := mfd.getGlobalFee(ctx, feeTx)
		if err != nil {
			panic(err)
		}
		// bypass tx without pay fee
		if len(feeCoins) == 0 {
			return next(ctx, tx, simulate)
		}
		// bypass with fee, fee denom must in requiredGlobalFees
		if !DenomsSubsetOfIncludingZero(feeCoins, requiredGlobalFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fees denom is wrong; got: %s required: %s", feeCoins, requiredGlobalFees)
		}
	}

	return next(ctx, tx, simulate)
}

// ParamStoreKeyMinGasPrices type require coins sorted. getGlobalFee will also return sorted coins (might return 0denom if globalMinGasPrice is 0)
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
	}
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredGlobalFees.Sort(), err
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
