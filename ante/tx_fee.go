package ante

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NoOpTxFeeChecker performs a no-op since the Global Fee module already
// check tx fees
func NoOpTxFeeChecker(_ sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	feeCoins := feeTx.GetFee()

	return feeTx.GetFee(), getTxPriority(feeCoins, int64(feeTx.GetGas())), nil
}

// taken from github.com/cosmos/cosmos-sdk@v0.47.3/x/auth/ante/validator_tx_fee.go
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
