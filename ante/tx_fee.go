package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NoOpTxFeeChecker is a TxFeeChecker, see x/auth/ante/fee.go, that performs a no-op
// by not checking the tx fee and always retruns a zero priority for all fee
func NoOpTxFeeChecker(_ sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	return feeTx.GetFee(), 0, nil
}
