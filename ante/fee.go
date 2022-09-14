package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v8/x/globalfee"
)

const maxBypassMinFeeMsgGasUsage = uint64(200_000)

// FeeWithBypassDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config) and global fee, and the fee denom should be in the global fees' denoms.
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use BypassMinFeeDecorator
// If the tx msg type is one of the bypass msg types, the tx is valid even if the min fee is lower than normally required.
// If the bypass tx still carries fees, the fee denom should be the same as global fee required.

var _ sdk.AnteDecorator = BypassMinFeeDecorator{}

type BypassMinFeeDecorator struct {
	BypassMinFeeMsgTypes []string
	GlobalMinFee         globalfee.ParamSource
}

const defaultZeroGlobalFeeDenom = "uatom"

func DefaultZeroGlobalFee() []sdk.DecCoin {
	return []sdk.DecCoin{sdk.NewDecCoinFromDec(defaultZeroGlobalFeeDenom, sdk.NewDec(0))}
}

func NewBypassMinFeeDecorator(bypassMsgTypes []string, paramSpace paramtypes.Subspace) BypassMinFeeDecorator {
	if !paramSpace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	return BypassMinFeeDecorator{
		BypassMinFeeMsgTypes: bypassMsgTypes,
		GlobalMinFee:         paramSpace,
	}
}

func (mfd BypassMinFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
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

		requiredGlobalFees := mfd.getGlobalFee(ctx, feeTx)
		allFees = CombinedFeeRequirement(requiredGlobalFees, requiredFees)

		// this is to ban 1stake passing if the globalfee is 1photon or 0photon
		// if feeCoins=[] and requiredGlobalFees has 0denom coins then it should pass.
		if !DenomsSubsetOfIncludingZero(feeCoins, allFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "fees %s is not a subset of required fees %s", feeCoins, allFees)
		}
		// At least one feeCoin amount must be GTE to one of the requiredGlobalFees
		if !IsAnyGTEIncludingZero(feeCoins, allFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees for global fee; got: %s required: %s", feeCoins, allFees)
		}
	}

	// when the tx is bypass msg type, still need to check the denom is not some unknown denom
	if ctx.IsCheckTx() && !simulate && allowedToBypassMinFee {
		requiredGlobalFees := mfd.getGlobalFee(ctx, feeTx)
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
