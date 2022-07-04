package ante

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"

	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

const maxBypassMinFeeMsgGasUsage = uint64(200_000)

// FeeWithBypassDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
//
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true. If fee is high enough or not
// CheckTx, then call next AnteHandler.
//
// CONTRACT: Tx must implement FeeTx to use FeeWithBypassDecorator

//todo check if this is needed
var _ sdk.AnteDecorator = BypassMinFeeDecorator{}

// paramSource is a read only subset of paramtypes.Subspace
type paramSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

type BypassMinFeeDecorator struct {
	BypassMinFeeMsgTypes []string
	GlobalMinFee         paramSource
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
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()
	msgs := feeTx.GetMsgs()

	// Only check for minimum fees if the execution mode is CheckTx and the tx does
	// not contain operator configured bypass messages. If the tx does contain
	// operator configured bypass messages only, it's total gas must be less than
	// or equal to a constant, otherwise minimum fees and global fees are checked to prevent spam.
	if ctx.IsCheckTx() && !simulate && !(mfd.bypassMinFeeMsgs(msgs) && gas <= uint64(len(msgs))*maxBypassMinFeeMsgGasUsage) {
		// check global fees
		if mfd.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
			var globalMinGasPrices sdk.DecCoins
			mfd.GlobalMinFee.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
			if !globalMinGasPrices.IsZero() {
				requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
				// Determine the required fees by multiplying each required minimum gas
				// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
				glDec := sdk.NewDec(int64(feeTx.GetGas()))
				for i, gp := range globalMinGasPrices {
					fee := gp.Amount.Mul(glDec)
					amount := fee.Ceil().RoundInt()
					requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, amount)
				}

				if !feeCoins.IsAnyGTE(requiredGlobalFees) {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredGlobalFees)
				}
			}
		}

		// check min gas price
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			// if passed global fee checks, but the denom is not in min_gas_price, skip the min gas price check
			if !feeCoins.DenomsSubsetOf(requiredFees.Sort()) {
				return next(ctx, tx, simulate)
			}

			if !feeCoins.IsAnyGTE(requiredFees) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	return next(ctx, tx, simulate)
}

func (mfd BypassMinFeeDecorator) bypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}

		return false
	}

	return true
}

//utils function: GetTxPriority
// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func GetTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
