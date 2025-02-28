package e2e

import (
	"encoding/json"
	"fmt"
	"time"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	icagen "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govmigrv3 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v3"
	govmigrv4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govlegacytypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func modifyGenesis(path, moniker, amountStr string, addrAll []sdk.AccAddress, basefee string, denom string) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	var balances []banktypes.Balance
	var genAccounts []*authtypes.BaseAccount
	for _, addr := range addrAll {
		balance := banktypes.Balance{Address: addr.String(), Coins: coins.Sort()}
		balances = append(balances, balance)
		genAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
		genAccounts = append(genAccounts, genAccount)
	}

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)
	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	for _, addr := range addrAll {
		if accs.Contains(addr) {
			return fmt.Errorf("failed to add account to genesis state; account already exists: %s", addr)
		}
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	for _, genAcct := range genAccounts {
		accs = append(accs, genAcct)
		accs = authtypes.SanitizeGenesisAccounts(accs)
	}

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}
	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances...)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	appState[banktypes.ModuleName] = bankGenStateBz

	// add ica host allowed msg types
	var icaGenesisState icagen.GenesisState

	if appState[icatypes.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[icatypes.ModuleName], &icaGenesisState)
	}

	icaGenesisState.HostGenesisState.Params.AllowMessages = []string{
		"/cosmos.authz.v1beta1.MsgExec",
		"/cosmos.authz.v1beta1.MsgGrant",
		"/cosmos.authz.v1beta1.MsgRevoke",
		"/cosmos.bank.v1beta1.MsgSend",
		"/cosmos.bank.v1beta1.MsgMultiSend",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
		"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
		"/cosmos.distribution.v1beta1.MsgFundCommunityPool",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.feegrant.v1beta1.MsgGrantAllowance",
		"/cosmos.feegrant.v1beta1.MsgRevokeAllowance",
		"/cosmos.gov.v1beta1.MsgVoteWeighted",
		"/cosmos.gov.v1beta1.MsgSubmitProposal",
		"/cosmos.gov.v1beta1.MsgDeposit",
		"/cosmos.gov.v1beta1.MsgVote",
		"/cosmos.staking.v1beta1.MsgEditValidator",
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.staking.v1beta1.MsgCreateValidator",
		"/cosmos.vesting.v1beta1.MsgCreateVestingAccount",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/tendermint.liquidity.v1beta1.MsgCreatePool",
		"/tendermint.liquidity.v1beta1.MsgSwapWithinBatch",
		"/tendermint.liquidity.v1beta1.MsgDepositWithinBatch",
		"/tendermint.liquidity.v1beta1.MsgWithdrawWithinBatch",
	}

	icaGenesisStateBz, err := cdc.MarshalJSON(&icaGenesisState)
	if err != nil {
		return fmt.Errorf("failed to marshal interchain accounts genesis state: %w", err)
	}
	appState[icatypes.ModuleName] = icaGenesisStateBz

	feemarketState := feemarkettypes.GetGenesisStateFromAppState(cdc, appState)
	feemarketState.Params.MinBaseGasPrice = math.LegacyMustNewDecFromStr(basefee)
	feemarketState.Params.FeeDenom = denom
	feemarketState.Params.DistributeFees = true
	feemarketState.State.BaseGasPrice = math.LegacyMustNewDecFromStr(basefee)
	feemarketStateBz, err := cdc.MarshalJSON(&feemarketState)
	if err != nil {
		return fmt.Errorf("failed to marshal feemarket genesis state: %w", err)
	}
	appState[feemarkettypes.ModuleName] = feemarketStateBz

	stakingGenState := stakingtypes.GetGenesisStateFromAppState(cdc, appState)
	stakingGenState.Params.BondDenom = denom
	stakingGenStateBz, err := cdc.MarshalJSON(stakingGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal staking genesis state: %s", err)
	}
	appState[stakingtypes.ModuleName] = stakingGenStateBz

	// Refactor to separate method
	amnt := math.NewInt(10000)
	quorum, _ := math.LegacyNewDecFromStr("0.000000000000000001")
	threshold, _ := math.LegacyNewDecFromStr("0.000000000000000001")

	maxDepositPeriod := 10 * time.Minute
	votingPeriod := 15 * time.Second
	expeditedVoting := 13 * time.Second

	govStateLegacy := govlegacytypes.NewGenesisState(1,
		govlegacytypes.NewDepositParams(sdk.NewCoins(sdk.NewCoin(denom, amnt)), maxDepositPeriod),
		govlegacytypes.NewVotingParams(votingPeriod),
		govlegacytypes.NewTallyParams(quorum, threshold, govlegacytypes.DefaultVetoThreshold),
	)

	govStateV3, err := govmigrv3.MigrateJSON(govStateLegacy)
	if err != nil {
		return fmt.Errorf("failed to migrate v1beta1 gov genesis state to v3: %w", err)
	}

	govStateV4, err := govmigrv4.MigrateJSON(govStateV3)
	if err != nil {
		return fmt.Errorf("failed to migrate v1beta1 gov genesis state to v4: %w", err)
	}

	govStateV4.Params.ExpeditedVotingPeriod = &expeditedVoting
	govStateV4.Params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(denom, amnt)) // same as normal for testing

	govGenStateBz, err := cdc.MarshalJSON(govStateV4)
	if err != nil {
		return fmt.Errorf("failed to marshal gov genesis state: %w", err)
	}
	appState[govtypes.ModuleName] = govGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}
	genDoc.AppState = appStateJSON

	return genutil.ExportGenesisFile(genDoc, genFile)
}
