//go:build unsafe_set_local_validator
// +build unsafe_set_local_validator

package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	gaia "github.com/cosmos/gaia/v16/app"
	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/crypto"
	tmd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	flagValidatorOperatorAddress = "validator-operator"
	flagValidatorPubKey          = "validator-pukey"
	flagValidatorPrivKey         = "validator-privkey"
	flagAccountsToFund           = "accounts-to-fund"
)

type valArgs struct {
	validatorOperatorAddress string           // valoper address
	validatorConsPubKeyByte  []byte           // validator's consensus public key
	validatorConsPrivKey     crypto.PrivKey   // validator's consensus private key
	accountsToFund           []sdk.AccAddress // list of accounts to fund and use for testing later on
}

func init() {
	unsafeSetValidatorFn = testnetUnsafeSetLocalValidatorCmd
}

func testnetUnsafeSetLocalValidatorCmd(appCreator servertypes.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsafe-set-local-validator",
		Short: "Updates chain's application and consensus state with provided validator info",
		Long: `The unsafe-set-local-validator command modifies both application and consensus stores within a local mainnet node,
with the aim of facilitating testing procedures. This command replaces existing validator data with updated information,
thereby removing the old validator set and introducing a new set suitable for local testing purposes. By altering the state extracted from the mainnet node,
it enables developers to configure their local environments to reflect mainnet conditions more accurately.

Example:
	simd testnet unsafe-set-local-validator --validator-operator="cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r" --validator-pukey="SLpHEfzQHuuNO9J1BB/hXyiH6c1NmpoIVQ2pMWmyctE=" --validator-privkey="AiayvI2px5CZVl/uOGmacfFjcIBoyk3Oa2JPBO6zEcdIukcR/NAe64070nUEH+FfKIfpzU2amghVDakxabJy0Q==" --accounts-to-fund="cosmos1ju6tlfclulxumtt2kglvnxduj5d93a64r5czge,cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl"
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args, err := getCommandArgs(cmd)
			if err != nil {
				return err
			}

			return modifyValidatorInfo(cmd, appCreator, args)
		},
	}

	cmd.Flags().String(flagValidatorOperatorAddress, "", "Validator operator address e.g. cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r")
	cmd.Flags().String(flagValidatorPubKey, "", "Validator tendermint/PubKeyEd25519 consensus public key from the priv_validato_key.json file")
	cmd.Flags().String(flagValidatorPrivKey, "", "Validator tendermint/PrivKeyEd25519 consensus private key from the priv_validato_key.json file")
	cmd.Flags().String(flagAccountsToFund, "", "Comma-separated list of account addresses that will be funded for testing purposes")

	return cmd
}

// parse the input flags and returns valArgs
func getCommandArgs(cmd *cobra.Command) (valArgs, error) {
	args := valArgs{}
	// validate and set validator operator address
	valoperAddress, err := cmd.Flags().GetString(flagValidatorOperatorAddress)
	if err != nil || valoperAddress == "" {
		return args, fmt.Errorf("invalid validator operator address string %w", err)
	}
	_, err = sdk.ValAddressFromBech32(valoperAddress)
	if err != nil {
		return args, fmt.Errorf("invalid validator operator address format %w", err)
	}
	args.validatorOperatorAddress = valoperAddress

	// validate and set validator pubkey
	validatorPubKey, err := cmd.Flags().GetString(flagValidatorPubKey)
	if err != nil || validatorPubKey == "" {
		return args, fmt.Errorf("invalid validator pubkey string %w", err)
	}
	decPubKey, err := base64.StdEncoding.DecodeString(validatorPubKey)
	if err != nil {
		return args, fmt.Errorf("cannot decode validator pubkey %w", err)
	}
	args.validatorConsPubKeyByte = []byte(decPubKey)

	// validate  and set validator privkey
	validatorPrivKey, err := cmd.Flags().GetString(flagValidatorPrivKey)
	if err != nil || validatorPrivKey == "" {
		return args, fmt.Errorf("invalid validator private key %w", err)
	}
	decPrivKey, err := base64.StdEncoding.DecodeString(validatorPrivKey)
	if err != nil {
		return args, fmt.Errorf("cannot decode validator private key %w", err)
	}
	args.validatorConsPrivKey = tmd25519.PrivKey([]byte(decPrivKey))

	// validate  and set accounts to fund
	accountsString, err := cmd.Flags().GetString(flagAccountsToFund)
	if err != nil {
		return args, fmt.Errorf("invalid addresses to fund %w", err)
	}

	for _, account := range strings.Split(accountsString, ",") {
		if account != "" {
			addr, err := sdk.AccAddressFromBech32(account)
			if err != nil {
				return args, fmt.Errorf("invalid bech32 address format %w", err)
			}
			args.accountsToFund = append(args.accountsToFund, addr)
		}
	}

	return args, nil
}

// modifies application and consensus states by replacing validator related data
func modifyValidatorInfo(cmd *cobra.Command, appCreator servertypes.AppCreator, args valArgs) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	homeDir, err := cmd.Flags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}
	serverCtx.Config.SetRoot(homeDir)

	//UPDATE APP STATE
	db, err := openDB(serverCtx.Config.RootDir, "application", server.GetAppDBBackend(serverCtx.Viper))
	if err != nil {
		return err
	}
	defer func() {
		if derr := db.Close(); derr != nil {
			serverCtx.Logger.Error("Failed to close application db", "err", derr)
			err = derr
		}
	}()

	app := appCreator(serverCtx.Logger, db, nil, serverCtx.Viper)
	gaiaApp, ok := app.(*gaia.GaiaApp)
	if !ok {
		return errors.New("invalid gaia application")
	}

	// we need to rollback to the previous version because app.CommitMultiStore().Commit() increments the version and
	// if we do not rollback, we will have a mismatch with the core and app versions(heights)
	latestHeight := rootmulti.GetLatestVersion(db)
	app.CommitMultiStore().RollbackToVersion(latestHeight - 1)
	if err != nil {
		return err
	}

	err = updateApplicationState(gaiaApp, args)
	if err != nil {
		return err
	}

	// save changes to the app store, this will also increment the version
	app.CommitMultiStore().Commit()

	//UPDATE CONSENSUS STATE
	appHash := app.CommitMultiStore().LastCommitID().Hash
	err = updateConsensusState(serverCtx, args, appHash)
	if err != nil {
		return err
	}

	return nil
}

func updateApplicationState(app *gaia.GaiaApp, args valArgs) error {
	pubkey := &ed25519.PubKey{Key: args.validatorConsPubKeyByte}
	pubkeyAny, err := types.NewAnyWithValue(pubkey)
	if err != nil {
		return err
	}

	appCtx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

	// STAKING
	// Create Validator struct for our new validator.
	newVal := stakingtypes.Validator{
		OperatorAddress: args.validatorOperatorAddress,
		ConsensusPubkey: pubkeyAny,
		Jailed:          false,
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(900000000000000),
		DelegatorShares: sdk.MustNewDecFromStr("10000000"),
		Description: stakingtypes.Description{
			Moniker: "Testnet Validator",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          sdk.MustNewDecFromStr("0.05"),
				MaxRate:       sdk.MustNewDecFromStr("0.1"),
				MaxChangeRate: sdk.MustNewDecFromStr("0.05"),
			},
		},
		MinSelfDelegation: sdk.OneInt(),
	}

	store := appCtx.KVStore(app.GetKey(stakingtypes.ModuleName))
	for _, v := range app.StakingKeeper.GetAllValidators(appCtx) {
		valConsAddr, err := v.GetConsAddr()
		if err != nil {
			return err
		}

		// delete the old validator record
		store.Delete(stakingtypes.GetValidatorKey(v.GetOperator()))
		store.Delete(stakingtypes.GetValidatorByConsAddrKey(valConsAddr))
		store.Delete(stakingtypes.GetValidatorsByPowerIndexKey(v, app.StakingKeeper.PowerReduction(appCtx)))
		store.Delete(stakingtypes.GetLastValidatorPowerKey(v.GetOperator()))
		if v.IsUnbonding() {
			app.StakingKeeper.DeleteValidatorQueueTimeSlice(appCtx, v.UnbondingTime, v.UnbondingHeight)
		}
	}

	// Add our validator to power and last validators store
	app.StakingKeeper.SetValidator(appCtx, newVal)
	err = app.StakingKeeper.SetValidatorByConsAddr(appCtx, newVal)
	if err != nil {
		return err
	}
	app.StakingKeeper.SetValidatorByPowerIndex(appCtx, newVal)
	app.StakingKeeper.SetLastValidatorPower(appCtx, newVal.GetOperator(), 0)
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(appCtx, newVal.GetOperator()); err != nil {
		return err
	}

	// DISTRIBUTION
	// Initialize records for this validator across all distribution stores
	app.DistrKeeper.SetValidatorHistoricalRewards(appCtx, newVal.GetOperator(), 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorCurrentRewards(appCtx, newVal.GetOperator(), distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorAccumulatedCommission(appCtx, newVal.GetOperator(), distrtypes.InitialValidatorAccumulatedCommission())
	app.DistrKeeper.SetValidatorOutstandingRewards(appCtx, newVal.GetOperator(), distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})

	// SLASHING
	// Set validator signing info for our new validator.
	newConsAddr := sdk.ConsAddress(pubkey.Address().Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}

	app.SlashingKeeper.SetValidatorSigningInfo(appCtx, newConsAddr, newValidatorSigningInfo)

	// BANK
	defaultCoins := sdk.NewCoins(sdk.NewInt64Coin(app.StakingKeeper.BondDenom(appCtx), 1000000000000))

	// Fund testnet accounts
	for _, account := range args.accountsToFund {
		err := app.BankKeeper.MintCoins(appCtx, minttypes.ModuleName, defaultCoins)
		if err != nil {
			return err
		}
		err = app.BankKeeper.SendCoinsFromModuleToAccount(appCtx, minttypes.ModuleName, account, defaultCoins)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateConsensusState(serverCtx *server.Context, args valArgs, appHash []byte) error {
	// create validator set from the local validator
	newTmVal := tmtypes.NewValidator(tmd25519.PubKey(args.validatorConsPubKeyByte), 900000000000000)
	vals := []*tmtypes.Validator{newTmVal}
	validatorSet := tmtypes.NewValidatorSet(vals)

	// CHANGE STATE CONSENSUS STORE
	stateDB, err := openDB(serverCtx.Config.RootDir, "state", server.GetAppDBBackend(serverCtx.Viper))
	if err != nil {
		return err
	}

	stateStore := sm.NewBootstrapStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})

	// load state in order to change validator set info
	state, err := stateStore.Load()
	if err != nil {
		return err
	}
	defer func() {
		if derr := stateStore.Close(); derr != nil {
			serverCtx.Logger.Error("Failed to close consensus state db", "err", derr)
			err = derr
		}
	}()

	state.LastValidators = validatorSet
	state.Validators = validatorSet
	state.NextValidators = validatorSet
	state.AppHash = appHash // needs to be updated since it is changed earlier in updateApplicationState
	// save state store
	if err = stateStore.Save(state); err != nil {
		return err
	}

	// last voting data must be updated because the distribution module will allocate tokens based on the last saved votes,
	// and the voting validator address has to be present in the staking module, which is not the case for old validator
	valInfo, err := loadValidatorsInfo(stateDB, state.LastBlockHeight)
	if err != nil {
		return err
	}

	protoValSet, err := validatorSet.ToProto()
	if err != nil {
		return err
	}
	valInfo.ValidatorSet = protoValSet
	valInfo.LastHeightChanged = state.LastBlockHeight

	// when the storeState is saved in consensus it is done for the nextBlock+1,
	// that is why we need to update 2 future blocks
	saveValidatorsInfo(stateDB, state.LastBlockHeight, valInfo)
	saveValidatorsInfo(stateDB, state.LastBlockHeight+1, valInfo)
	saveValidatorsInfo(stateDB, state.LastBlockHeight+2, valInfo)

	// CHANGE BLOCK CONSENSUS STORE
	// we need to change the last commit data by updating the signature's info. Consensus will match the validator's set length
	// and size of the lastCommit signatures when building the last commit info and they have to match
	blockStoreDB, err := openDB(serverCtx.Config.RootDir, "blockstore", server.GetAppDBBackend(serverCtx.Viper))
	if err != nil {
		return err
	}
	defer func() {
		if derr := blockStoreDB.Close(); derr != nil {
			serverCtx.Logger.Error("Failed to close consensus blockstore db", "err", derr)
			err = derr
		}
	}()

	blockStore := store.NewBlockStore(blockStoreDB)
	lastCommit := blockStore.LoadSeenCommit(state.LastBlockHeight)

	vote := lastCommit.GetVote(0)
	if vote == nil {
		return errors.New("cannot get the vote from the last commit")
	}

	voteSignBytes := tmtypes.VoteSignBytes(state.ChainID, vote.ToProto())
	signatureBytes, err := args.validatorConsPrivKey.Sign(voteSignBytes)
	if err != nil {
		return err
	}

	lastCommit.Signatures = []tmtypes.CommitSig{{
		BlockIDFlag:      lastCommit.Signatures[0].BlockIDFlag,
		ValidatorAddress: newTmVal.Address,
		Timestamp:        lastCommit.Signatures[0].Timestamp,
		Signature:        []byte(signatureBytes),
	}}

	return blockStore.SaveSeenCommit(state.LastBlockHeight, lastCommit)
}

func loadValidatorsInfo(db dbm.DB, height int64) (*cmtstate.ValidatorsInfo, error) {
	buf, err := db.Get(calcValidatorsKey(height))
	if err != nil {
		return nil, err
	}

	if len(buf) == 0 {
		return nil, errors.New("validator info retrieved from db is empty")
	}

	v := new(cmtstate.ValidatorsInfo)
	err = v.Unmarshal(buf)

	return v, err
}

func saveValidatorsInfo(db dbm.DB, height int64, valInfo *cmtstate.ValidatorsInfo) error {
	bz, err := valInfo.Marshal()
	if err != nil {
		return err
	}

	err = db.Set(calcValidatorsKey(height), bz)
	if err != nil {
		return err
	}

	return nil
}

func calcValidatorsKey(height int64) []byte {
	return []byte(fmt.Sprintf("validatorsKey:%v", height))
}

func openDB(rootDir, dbName string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB(dbName, backendType, dataDir)
}
