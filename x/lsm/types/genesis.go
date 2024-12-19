package types

import "cosmossdk.io/math"

func NewGenesisState(
	params Params,
	tsr []TokenizeShareRecord,
	recordId uint64,
	liquidStakeTokens math.Int,
	locks []TokenizeShareLock,
) *GenesisState {
	return &GenesisState{
		Params:                    params,
		TokenizeShareRecords:      tsr,
		LastTokenizeShareRecordId: recordId,
		TotalLiquidStakedTokens:   liquidStakeTokens,
		TokenizeShareLocks:        locks,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}
