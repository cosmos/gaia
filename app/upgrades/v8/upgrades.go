package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("start to run module migrations...")

		// Add Bank denom metadata
		keepers.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*banktypes.DenomUnit{
				{"uatom", uint32(0), []string{"microatom"}},
				{"matom", uint32(3), []string{"milliatom"}},
				{"atom", uint32(6), nil},
			},
			Base:    "uatom",
			Display: "atom",
		})
		
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
