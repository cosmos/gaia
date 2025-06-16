package liquid

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/gaia/v25/x/liquid/types"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "LiquidValidators",
					Use:       "liquid-validators",
					Short:     "Query for all liquid validators",
					Example:   fmt.Sprintf("$ %s query liquid liquid-validators", version.AppName),
				},
				{
					RpcMethod: "LiquidValidator",
					Use:       "liquid-validator [validator-address]",
					Short:     "Query individual liquid validator by validator address",
					Example: fmt.Sprintf(
						"$ %s query liquid liquid-validator %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj",
						version.AppName, sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "TokenizeShareRecordById",
					Use:       "tokenize-share-record-by-id [id]",
					Short:     "Query individual tokenize share record information by share by id",
					Example:   fmt.Sprintf("$ %s query liquid tokenize-share-record-by-id [id]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "TokenizeShareRecordByDenom",
					Use:       "tokenize-share-record-by-denom [denom]",
					Short:     "Query individual tokenize share record information by share denom",
					Example:   fmt.Sprintf("$ %s query liquid tokenize-share-record-by-denom [denom]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
				},
				{
					RpcMethod: "TokenizeShareRecordsOwned",
					Use:       "tokenize-share-records-owned [owner]",
					Short:     "Query tokenize share records by address",
					Example:   fmt.Sprintf("$ %s query liquid tokenize-share-records-owned [owner]", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
				{
					RpcMethod: "AllTokenizeShareRecords",
					Use:       "all-tokenize-share-records",
					Short:     "Query for all tokenize share records",
					Example:   fmt.Sprintf("$ %s query liquid all-tokenize-share-records", version.AppName),
				},
				{
					RpcMethod: "LastTokenizeShareRecordId",
					Use:       "last-tokenize-share-record-id",
					Short:     "Query for last tokenize share record id",
					Example:   fmt.Sprintf("$ %s query liquid last-tokenize-share-record-id", version.AppName),
				},
				{
					RpcMethod: "TotalTokenizeSharedAssets",
					Use:       "total-tokenize-share-assets",
					Short:     "Query for total tokenized staked assets",
					Example:   fmt.Sprintf("$ %s query liquid total-tokenize-share-assets", version.AppName),
				},
				{
					RpcMethod: "TotalLiquidStaked",
					Use:       "total-liquid-staked",
					Short:     "Query for total liquid staked tokens",
					Example:   fmt.Sprintf("$ %s query liquid total-liquid-staked", version.AppName),
				},
				{
					RpcMethod: "TokenizeShareLockInfo",
					Use:       "tokenize-share-lock-info [address]",
					Short:     "Query tokenize share lock information",
					Long:      "Query the status of a tokenize share lock for a given account",
					Example: fmt.Sprintf(
						"$ %s query liquid tokenize-share-lock-info %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj",
						version.AppName, sdk.GetConfig().GetBech32AccountAddrPrefix()),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "TokenizeShareRecordReward",
					Use:       "tokenize-share-record-rewards [owner]",
					Short:     "Query liquid tokenize share record rewards",
					Example: fmt.Sprintf(`$ %s query liquid tokenize-share-record-rewards %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj`,
						version.AppName, sdk.GetConfig().GetBech32AccountAddrPrefix()),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner_address"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			RpcCommandOptions:    []*autocliv1.RpcCommandOptions{},
			EnhanceCustomCommand: false, // use custom commands only until v0.51
		},
	}
}
