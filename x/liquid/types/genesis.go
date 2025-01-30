package types

import "cosmossdk.io/math"

func NewGenesisState(
	params Params,
	tsr []TokenizeShareRecord,
	recordID uint64,
	liquidStakeTokens math.Int,
	locks []TokenizeShareLock,
) *GenesisState {
	return &GenesisState{
		Params:                    params,
		TokenizeShareRecords:      tsr,
		LastTokenizeShareRecordId: recordID,
		TotalLiquidStakedTokens:   liquidStakeTokens,
		TokenizeShareLocks:        locks,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func ValidateGenesis(gs *GenesisState) error {
	return gs.Params.Validate()
}
