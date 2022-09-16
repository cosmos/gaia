package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmtypes "github.com/tendermint/tendermint/types"

	globfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"
)

func getGenDoc(path string) (*tmtypes.GenesisDoc, error) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(path)

	genFile := config.GenesisFile()
	doc := &tmtypes.GenesisDoc{}

	if _, err := os.Stat(genFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		var err error

		doc, err = tmtypes.GenesisDocFromFile(genFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read genesis doc from file: %w", err)
		}
	}

	return doc, nil
}

func modifyGenesis(path, moniker, amountStr string, addrAll []sdk.AccAddress, globfees string, denom string) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config
	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	balances := []banktypes.Balance{}
	genAccounts := []*authtypes.BaseAccount{}
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

	// setup global fee in genesis
	globfeeState := globfeetypes.GetGenesisStateFromAppState(cdc, appState)
	minGases, err := sdk.ParseDecCoins(globfees)
	if err != nil {
		return fmt.Errorf("failed to parse fee coins: %w", err)
	}
	globfeeState.Params.MinimumGasPrices = minGases
	globfeeStateBz, err := cdc.MarshalJSON(globfeeState)
	if err != nil {
		return fmt.Errorf("failed to marshal global fee genesis state: %w", err)
	}
	appState[globfeetypes.ModuleName] = globfeeStateBz

	stakingGenState := staketypes.GetGenesisStateFromAppState(cdc, appState)
	stakingGenState.Params.BondDenom = denom
	stakingGenStateBz, err := cdc.MarshalJSON(stakingGenState)
	if err != nil {
		fmt.Errorf("Failed to marshal staking genesis state: %w", err)
	}
	appState[staketypes.ModuleName] = stakingGenStateBz

	// Refactor to separate method
	amnt := math.NewInt(10000)
	quorum, _ := sdk.NewDecFromStr("0.000000000000000001")
	threshold, _ := sdk.NewDecFromStr("0.000000000000000001")

	govState := govv1beta1.NewGenesisState(1,
		govv1beta1.NewDepositParams(sdk.NewCoins(sdk.NewCoin(denom, amnt)), 10*time.Minute),
		govv1beta1.NewVotingParams(15*time.Second),
		govv1beta1.NewTallyParams(quorum, threshold, govv1beta1.DefaultVetoThreshold),
	)

	govGenStateBz, err := cdc.MarshalJSON(govState)
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
