package types

// NewGenesisState returns new GenesisState.
func NewGenesisState(params Params, liquidityPoolRecords []PoolRecord) *GenesisState {
	return &GenesisState{
		Params:      params,
		PoolRecords: liquidityPoolRecords,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []PoolRecord{})
}

// ValidateGenesis validates GenesisState.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	for _, record := range data.PoolRecords {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates PoolRecord.
func (record PoolRecord) Validate() error {
	if record.PoolBatch.DepositMsgIndex == 0 ||
		(len(record.DepositMsgStates) > 0 && record.PoolBatch.DepositMsgIndex != record.DepositMsgStates[len(record.DepositMsgStates)-1].MsgIndex+1) {
		return ErrBadBatchMsgIndex
	}
	if record.PoolBatch.WithdrawMsgIndex == 0 ||
		(len(record.WithdrawMsgStates) != 0 && record.PoolBatch.WithdrawMsgIndex != record.WithdrawMsgStates[len(record.WithdrawMsgStates)-1].MsgIndex+1) {
		return ErrBadBatchMsgIndex
	}
	if record.PoolBatch.SwapMsgIndex == 0 ||
		(len(record.SwapMsgStates) != 0 && record.PoolBatch.SwapMsgIndex != record.SwapMsgStates[len(record.SwapMsgStates)-1].MsgIndex+1) {
		return ErrBadBatchMsgIndex
	}
	return nil
}
