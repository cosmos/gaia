package v7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"

	"github.com/cosmos/gaia/v9/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm[icatypes.ModuleName] = mm.Modules[icatypes.ModuleName].ConsensusVersion()
		// create ICS27 Controller submodule params
		controllerParams := icacontrollertypes.Params{}
		// create ICS27 Host submodule params
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				authzMsgExec,
				authzMsgGrant,
				authzMsgRevoke,
				bankMsgSend,
				bankMsgMultiSend,
				distrMsgSetWithdrawAddr,
				distrMsgWithdrawValidatorCommission,
				distrMsgFundCommunityPool,
				distrMsgWithdrawDelegatorReward,
				feegrantMsgGrantAllowance,
				feegrantMsgRevokeAllowance,
				govMsgVoteWeighted,
				govMsgSubmitProposal,
				govMsgDeposit,
				govMsgVote,
				stakingMsgEditValidator,
				stakingMsgDelegate,
				stakingMsgUndelegate,
				stakingMsgBeginRedelegate,
				stakingMsgCreateValidator,
				vestingMsgCreateVestingAccount,
				ibcMsgTransfer,
				liquidityMsgCreatePool,
				liquidityMsgSwapWithinBatch,
				liquidityMsgDepositWithinBatch,
				liquidityMsgWithdrawWithinBatch,
			},
		}

		ctx.Logger().Info("start to init interchainaccount module...")

		// initialize ICS27 module
		icaModule, correctTypecast := mm.Modules[icatypes.ModuleName].(ica.AppModule)
		if !correctTypecast {
			panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
		}
		icaModule.InitModule(ctx, controllerParams, hostParams)

		ctx.Logger().Info("start to run module migrations...")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
