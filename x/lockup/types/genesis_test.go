package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidation(t *testing.T) {
	defaultGenesis := DefaultGenesisState()
	err := defaultGenesis.ValidateBasic()
	assert.Nil(t, err, "error produced from default genesis ValidateBasic %v", err)
	err = ValidateLocked(defaultGenesis.Params.Locked)
	assert.Nil(t, err, "error produced from default locked validation")
	err = ValidateLockExempt(defaultGenesis.Params.LockExempt)
	assert.Nil(t, err, "error produced from default lockExempt validation")
	err = ValidateLockedMessageTypes(defaultGenesis.Params.LockedMessageTypes)
	assert.Nil(t, err, "error produced from default lockedMessageTypes validation")

	badGenesis := GenesisState{Params: &Params{Locked: false}}
	err = badGenesis.ValidateBasic()
	assert.NotNil(t, err, "badGenesis did not produce an error after ValidateBasic")
	err = ValidateLocked(nil)
	assert.NotNil(t, err, "nil locked did not produce an error in validation fn")
	err = ValidateLockExempt(badGenesis.Params.LockExempt)
	assert.NotNil(t, err, "badGenesis lockExempt did not produce an error in validation fn")
	err = ValidateLockedMessageTypes(nil)
	assert.NotNil(t, err, "badGenesis lockedMessageTypes did not produce an error in validation fn")
}
