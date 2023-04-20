package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAuxFunctions(t *testing.T) {
	type badType struct{}
	for _, v := range []func(interface{}) error{
		validatePoolTypes,
		validateMinInitDepositAmount,
		validateInitPoolCoinMintAmount,
		validateMaxReserveCoinAmount,
		validatePoolCreationFee,
		validateSwapFeeRate,
		validateWithdrawFeeRate,
		validateMaxOrderAmountRatio,
		validateUnitBatchHeight,
		validateCircuitBreakerEnabled,
	} {
		err := v(badType{})
		require.EqualError(t, err, "invalid parameter type: types.badType")
	}
}
