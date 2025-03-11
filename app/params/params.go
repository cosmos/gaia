package params

// Simulation parameter constants
const (
	StakePerAccount           = "stake_per_account"
	InitiallyBondedValidators = "initially_bonded_validators"
)

// Keeper constants
const (
	// MaxIBCCallbackGas should roughly be a couple orders of magnitude larger than needed.
	MaxIBCCallbackGas = uint64(10_000_000)
)
