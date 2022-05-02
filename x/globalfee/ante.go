package globalfee

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

var _ sdk.AnteDecorator = GlobalMinimumChainFeeDecorator{}

// paramSource is a read only subset of paramtypes.Subspace
type paramSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

// GlobalMinimumChainFeeDecorator Ante decorator that enforces a minimum fee set for all transactions.
// This minimum can be 0 though.
type GlobalMinimumChainFeeDecorator struct {
	paramSource paramSource
}

// NewGlobalMinimumChainFeeDecorator constructor
func NewGlobalMinimumChainFeeDecorator(paramSpace paramtypes.Subspace) GlobalMinimumChainFeeDecorator {
	if !paramSpace.HasKeyTable() {
		panic("paramspace was not set up via module")
	}

	return GlobalMinimumChainFeeDecorator{
		paramSource: paramSpace,
	}
}

// AnteHandle method that performs custom pre- and post-processing.
func (g GlobalMinimumChainFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if g.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "tx must be a sdk FeeTx")
		}

		var minGasPrices sdk.DecCoins
		g.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &minGasPrices)
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdk.NewDec(int64(feeTx.GetGas()))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				amount := fee.Ceil().RoundInt()
				requiredFees[i] = sdk.NewCoin(gp.Denom, amount)
			}

			if !feeTx.GetFee().IsAnyGTE(requiredFees) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "got: %s required: %s", feeTx.GetFee(), requiredFees)
			}
		}
	}
	return next(ctx, tx, simulate)
}
