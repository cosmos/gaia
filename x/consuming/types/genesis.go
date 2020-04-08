package types

// NewGenesisState creates a new genesis state.
func NewGenesisState() GenesisState {
	return GenesisState{}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error { return nil }
