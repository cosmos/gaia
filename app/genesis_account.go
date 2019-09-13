package app

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

var _ authexported.GenesisAccount = (*GenesisAccount)(nil)

func init() {
	authtypes.RegisterAccountTypeCodec(GenesisAccount{}, "gaia/GenesisAccount")
}

// SimGenesisAccount defines a type that implements the GenesisAccount interface
// to be used for simulation accounts in the genesis state.
type GenesisAccount struct {
	*authtypes.BaseAccount

	// vesting account fields
	OriginalVesting  sdk.Coins `json:"original_vesting" yaml:"original_vesting"`   // total vesting coins upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free" yaml:"delegated_free"`       // delegated vested coins at time of delegation
	DelegatedVesting sdk.Coins `json:"delegated_vesting" yaml:"delegated_vesting"` // delegated vesting coins at time of delegation
	StartTime        int64     `json:"start_time" yaml:"start_time"`               // vesting start time (UNIX Epoch time)
	EndTime          int64     `json:"end_time" yaml:"end_time"`                   // vesting end time (UNIX Epoch time)

	// module account fields
	ModuleName        string   `json:"module_name" yaml:"module_name"`               // name of the module account
	ModulePermissions []string `json:"module_permissions" yaml:"module_permissions"` // permissions of module account
}

// NewBaseGenesisAccount creates a GenesisAccount instance from only a BaseAccount.
func NewBaseGenesisAccount(acc *authtypes.BaseAccount) GenesisAccount {
	return GenesisAccount{
		BaseAccount: acc,
	}
}

func NewGenesisAccount(
	address sdk.AccAddress, coins, vestingAmount sdk.Coins,
	vestingStartTime, vestingEndTime int64, module string, permissions ...string,
) GenesisAccount {

	return GenesisAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       address,
			Coins:         coins,
			Sequence:      0,
			AccountNumber: 0, // ignored; set by the account keeper during InitGenesis
		},
		OriginalVesting:   vestingAmount,
		DelegatedFree:     sdk.Coins{},
		DelegatedVesting:  sdk.Coins{},
		StartTime:         vestingStartTime,
		EndTime:           vestingEndTime,
		ModuleName:        module,
		ModulePermissions: permissions,
	}
}

// Validate validates the GenesisAccount and returns an error if the account is
// invalid.
func (g GenesisAccount) Validate() error {
	if !g.OriginalVesting.IsZero() {
		if g.OriginalVesting.IsAnyGT(g.Coins) {
			return errors.New("vesting amount cannot be greater than total amount")
		}
		if g.StartTime >= g.EndTime {
			return errors.New("vesting start-time cannot be before end-time")
		}
	}

	if g.ModuleName != "" {
		ma := supply.ModuleAccount{
			BaseAccount: g.BaseAccount, Name: g.ModuleName, Permissions: g.ModulePermissions,
		}
		if err := ma.Validate(); err != nil {
			return err
		}
	}

	return g.BaseAccount.Validate()
}
