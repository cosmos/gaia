package gaia

// TODO: Enable with ICS
import (
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibcstakinginterface "github.com/cosmos/interchain-security/legacy_ibc_testing/core"
	ics "github.com/cosmos/interchain-security/testutil/integration"
	ibcproviderkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
)

// GetProviderKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetProviderKeeper() ibcproviderkeeper.Keeper { //nolint:nolintlint
	return app.ProviderKeeper
}

// GetStakingKeeper implements the TestingApp interface. Needed for ICS.
func (app *GaiaApp) GetStakingKeeper() ibcstakinginterface.StakingKeeper { //nolint:nolintlint
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *GaiaApp) GetIBCKeeper() *ibckeeper.Keeper { //nolint:nolintlint
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *GaiaApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper { //nolint:nolintlint
	return app.ScopedIBCKeeper
}

// GetE2eStakingKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetTestStakingKeeper() ics.TestStakingKeeper { //nolint:nolintlint
	return app.StakingKeeper
}

// GetE2eBankKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetTestBankKeeper() ics.TestBankKeeper { //nolint:nolintlint
	return app.BankKeeper
}

// GetE2eSlashingKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetTestSlashingKeeper() ics.TestSlashingKeeper { //nolint:nolintlint
	return app.SlashingKeeper
}

// GetE2eDistributionKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetTestDistributionKeeper() ics.TestDistributionKeeper { //nolint:nolintlint
	return app.DistrKeeper
}
