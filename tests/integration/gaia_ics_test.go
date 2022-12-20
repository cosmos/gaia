package integration

//
//import (
//	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
//	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
//	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
//	ibctesting "github.com/cosmos/ibc-go/v3/testing"
//	"github.com/cosmos/interchain-security/testutil/e2e"
//	ibcproviderkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
//	"testing"
//
//	gaiaApp "github.com/cosmos/gaia/v8/app"
//)
//
//var _ ibctesting.TestingApp = (*GaiaTestingApp)(nil)
//
//type GaiaTestingApp struct {
//	*gaiaApp.GaiaApp
//}
//
//// GetProviderKeeper implements the ProviderApp interface.
//func (app *GaiaTestingApp) GetProviderKeeper() ibcproviderkeeper.Keeper { //nolint:unused
//	return app.ProviderKeeper
//}
//
//// GetE2eStakingKeeper implements the ProviderApp interface.
//func (app *GaiaTestingApp) GetE2eStakingKeeper() e2e.E2eStakingKeeper { //nolint:unused
//	return app.StakingKeeper
//}
//
//// GetE2eBankKeeper implements the ProviderApp interface.
//func (app *GaiaTestingApp) GetE2eBankKeeper() e2e.E2eBankKeeper { //nolint:unused
//	return app.BankKeeper
//}
//
//// GetE2eSlashingKeeper implements the ProviderApp interface.
//func (app *GaiaTestingApp) GetE2eSlashingKeeper() e2e.E2eSlashingKeeper { //nolint:unused
//	return app.SlashingKeeper
//}
//
//// GetE2eDistributionKeeper implements the ProviderApp interface.
//func (app *GaiaTestingApp) GetE2eDistributionKeeper() e2e.E2eDistributionKeeper { //nolint:unused
//	return app.DistrKeeper
//}
//
//// GetStakingKeeper implements the TestingApp interface. Needed for ICS.
//func (appKeepers *GaiaTestingApp) GetStakingKeeper() stakingkeeper.Keeper {
//	return appKeepers.StakingKeeper
//}
//
//// GetIBCKeeper implements the TestingApp interface.
//func (appKeepers *GaiaTestingApp) GetIBCKeeper() *ibckeeper.Keeper {
//	return appKeepers.IBCKeeper
//}
//
//// GetScopedIBCKeeper implements the TestingApp interface.
//func (appKeepers *GaiaTestingApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
//	return appKeepers.ScopedIBCKeeper
//}
//
//func TestGaiaCompile(t *testing.T) {
//	gaia := gaiaApp.DefaultNodeHome
//	t.Logf(gaia)
//}
