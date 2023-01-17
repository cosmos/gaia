package gaia

import (
	"github.com/cosmos/interchain-security/testutil/e2e"
	ibcproviderkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
)

// ProviderApp interface implementations for e2e tests

// GetProviderKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetProviderKeeper() ibcproviderkeeper.Keeper { //nolint:nolintlint
	return app.ProviderKeeper
}

// GetE2eStakingKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetE2eStakingKeeper() e2e.E2eStakingKeeper { //nolint:nolintlint
	return app.StakingKeeper
}

// GetE2eBankKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetE2eBankKeeper() e2e.E2eBankKeeper { //nolint:nolintlint
	return app.BankKeeper
}

// GetE2eSlashingKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetE2eSlashingKeeper() e2e.E2eSlashingKeeper { //nolint:nolintlint
	return app.SlashingKeeper
}

// GetE2eDistributionKeeper implements the ProviderApp interface.
func (app *GaiaApp) GetE2eDistributionKeeper() e2e.E2eDistributionKeeper { //nolint:nolintlint
	return app.DistrKeeper
}
