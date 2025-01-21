//go:build unsafe_start_local_validator
// +build unsafe_start_local_validator

package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"cosmossdk.io/math"
	cometdbm "github.com/cometbft/cometbft-db"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	gaia "github.com/cosmos/gaia/v23/app"

	"cosmossdk.io/log"
	"github.com/cometbft/cometbft/crypto"
	tmd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	valVotingPower int64 = 900000000000000
)

var (
	flagValidatorOperatorAddress = "validator-operator"
	flagValidatorPubKey          = "validator-pubkey"
	flagValidatorPrivKey         = "validator-privkey"
	flagAccountsToFund           = "accounts-to-fund"
)

type valArgs struct {
	validatorOperatorAddress string           // valoper address
	validatorConsPubKeyByte  []byte           // validator's consensus public key
	validatorConsPrivKey     crypto.PrivKey   // validator's consensus private key
	accountsToFund           []sdk.AccAddress // list of accounts to fund and use for testing later on
	homeDir                  string
}

func init() {
	unsafeStartValidatorFn = testnetUnsafeStartLocalValidatorCmd
}

func testnetUnsafeStartLocalValidatorCmd(ac appCreator) *cobra.Command {
	cmd := server.StartCmd(ac.newTestingApp, gaia.DefaultNodeHome)
	cmd.Use = "unsafe-start-local-validator"
	cmd.Short = "Updates chain's application and consensus state with provided validator info and starts the node"
	cmd.Long = `The unsafe-start-local-validator command modifies both application and consensus stores within a local mainnet node and starts the node,
with the aim of facilitating testing procedures. This command replaces existing validator data with updated information,
thereby removing the old validator set and introducing a new set suitable for local testing purposes. By altering the state extracted from the mainnet node,
it enables developers to configure their local environments to reflect mainnet conditions more accurately.

Example:
	simd testnet unsafe-start-local-validator --validator-operator="cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r" --validator-pukey="SLpHEfzQHuuNO9J1BB/hXyiH6c1NmpoIVQ2pMWmyctE=" --validator-privkey="AiayvI2px5CZVl/uOGmacfFjcIBoyk3Oa2JPBO6zEcdIukcR/NAe64070nUEH+FfKIfpzU2amghVDakxabJy0Q==" --accounts-to-fund="cosmos1ju6tlfclulxumtt2kglvnxduj5d93a64r5czge,cosmos1r5v5srda7xfth3hn2s26txvrcrntldjumt8mhl" [other_server_start_flags]
	`
	cmd.Flags().String(flagValidatorOperatorAddress, "", "Validator operator address e.g. cosmosvaloper17fjdcqy7g80pn0seexcch5pg0dtvs45p57t97r")
	cmd.Flags().String(flagValidatorPubKey, "", "Validator tendermint/PubKeyEd25519 consensus public key from the priv_validato_key.json file")
	cmd.Flags().String(flagValidatorPrivKey, "", "Validator tendermint/PrivKeyEd25519 consensus private key from the priv_validato_key.json file")
	cmd.Flags().String(flagAccountsToFund, "", "Comma-separated list of account addresses that will be funded for testing purposes")

	return cmd
}

// parse the input flags and returns valArgs
func getCommandArgs(appOpts servertypes.AppOptions) (valArgs, error) {
	args := valArgs{}
	// validate and set validator operator address
	valoperAddress := cast.ToString(appOpts.Get(flagValidatorOperatorAddress))
	if valoperAddress == "" {
		return args, errors.New("invalid validator operator address string")
	}
	_, err := sdk.ValAddressFromBech32(valoperAddress)
	if err != nil {
		return args, fmt.Errorf("invalid validator operator address format %w", err)
	}
	args.validatorOperatorAddress = valoperAddress

	// validate and set validator pubkey
	validatorPubKey := cast.ToString(appOpts.Get(flagValidatorPubKey))
	if validatorPubKey == "" {
		return args, errors.New("invalid validator pubkey string")
	}
	decPubKey, err := base64.StdEncoding.DecodeString(validatorPubKey)
	if err != nil {
		return args, fmt.Errorf("cannot decode validator pubkey %w", err)
	}
	args.validatorConsPubKeyByte = []byte(decPubKey)

	// validate  and set validator privkey
	validatorPrivKey := cast.ToString(appOpts.Get(flagValidatorPrivKey))
	if validatorPrivKey == "" {
		return args, fmt.Errorf("invalid validator private key %w", err)
	}
	decPrivKey, err := base64.StdEncoding.DecodeString(validatorPrivKey)
	if err != nil {
		return args, fmt.Errorf("cannot decode validator private key %w", err)
	}
	args.validatorConsPrivKey = tmd25519.PrivKey([]byte(decPrivKey))

	// validate  and set accounts to fund
	accountsString := cast.ToString(appOpts.Get(flagAccountsToFund))

	for _, account := range strings.Split(accountsString, ",") {
		if account != "" {
			addr, err := sdk.AccAddressFromBech32(account)
			if err != nil {
				return args, fmt.Errorf("invalid bech32 address format %w", err)
			}
			args.accountsToFund = append(args.accountsToFund, addr)
		}
	}

	// home dir
	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	if homeDir == "" {
		return args, errors.New("invalid home dir")
	}
	args.homeDir = homeDir

	return args, nil
}

// returns gaia app with modified application and consensus states by replacing validator related data
func (a appCreator) newTestingApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	app := a.newApp(logger, db, traceStore, appOpts)
	gaiaApp, ok := app.(*gaia.GaiaApp)
	if !ok {
		panic(errors.New("invalid gaia application"))
	}

	// Get command args
	args, err := getCommandArgs(appOpts)
	if err != nil {
		panic(err)
	}

	// Update app state
	err = updateApplicationState(gaiaApp, args)
	if err != nil {
		panic(err)
	}

	// Update consensus state
	err = updateConsensusState(logger, appOpts, gaiaApp.CommitMultiStore().LatestVersion(), args)
	if err != nil {
		panic(err)
	}

	return gaiaApp
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
		Tokens:          math.NewInt(valVotingPower),
		DelegatorShares: math.LegacyMustNewDecFromStr("10000000"),
		Description: stakingtypes.Description{
			Moniker: "Testnet Validator",
		},
		Commission: stakingtypes.Commission{
			CommissionRates: stakingtypes.CommissionRates{
				Rate:          math.LegacyMustNewDecFromStr("0.05"),
				MaxRate:       math.LegacyMustNewDecFromStr("0.1"),
				MaxChangeRate: math.LegacyMustNewDecFromStr("0.05"),
			},
		},
		MinSelfDelegation: math.OneInt(),
	}

	newValAddr, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(newVal.GetOperator())
	if err != nil {
		return err
	}

	store := appCtx.KVStore(app.GetKey(stakingtypes.ModuleName))
	validators, err := app.StakingKeeper.GetAllValidators(appCtx)
	if err != nil {
		return err
	}
	for _, v := range validators {
		valConsAddr, err := v.GetConsAddr()
		if err != nil {
			return err
		}

		// delete the old validator record
		valAddr, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(v.GetOperator())
		if err != nil {
			return err
		}
		store.Delete(stakingtypes.GetValidatorKey(valAddr))
		store.Delete(stakingtypes.GetValidatorByConsAddrKey(valConsAddr))
		store.Delete(stakingtypes.GetValidatorsByPowerIndexKey(v, app.StakingKeeper.PowerReduction(appCtx), app.StakingKeeper.ValidatorAddressCodec()))
		store.Delete(stakingtypes.GetLastValidatorPowerKey(valAddr))
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
	app.StakingKeeper.SetLastValidatorPower(appCtx, newValAddr, valVotingPower)
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(appCtx, newValAddr); err != nil {
		return err
	}

	// DISTRIBUTION
	// Initialize records for this validator across all distribution stores
	app.DistrKeeper.SetValidatorHistoricalRewards(appCtx, newValAddr, 0, distrtypes.NewValidatorHistoricalRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorCurrentRewards(appCtx, newValAddr, distrtypes.NewValidatorCurrentRewards(sdk.DecCoins{}, 1))
	app.DistrKeeper.SetValidatorAccumulatedCommission(appCtx, newValAddr, distrtypes.InitialValidatorAccumulatedCommission())
	app.DistrKeeper.SetValidatorOutstandingRewards(appCtx, newValAddr, distrtypes.ValidatorOutstandingRewards{Rewards: sdk.DecCoins{}})

	// SLASHING
	// Set validator signing info for our new validator.
	newConsAddr := sdk.ConsAddress(pubkey.Address().Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     newConsAddr.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}

	app.SlashingKeeper.SetValidatorSigningInfo(appCtx, newConsAddr, newValidatorSigningInfo)

	// PROVIDER
	app.ProviderKeeper.DeleteLastProviderConsensusValSet(appCtx)

	// GOVERNANCE
	shortVotingPeriod := time.Second * 20
	expeditedVotingPeriod := time.Second * 10
	params, err := app.GovKeeper.Params.Get(appCtx)
	if err != nil {
		return err
	}
	params.VotingPeriod = &shortVotingPeriod
	params.ExpeditedVotingPeriod = &expeditedVotingPeriod
	err = app.GovKeeper.Params.Set(appCtx, params)
	if err != nil {
		return err
	}
	appCtx.Logger().Info("Updated governance voting period", "voting_period", shortVotingPeriod, "expedited_voting_period", expeditedVotingPeriod)

	// BANK
	bondDenom, err := app.StakingKeeper.BondDenom(appCtx)
	if err != nil {
		return err
	}
	defaultCoins := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1000000000000))

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

func updateConsensusState(logger log.Logger, appOpts servertypes.AppOptions, appHeight int64, args valArgs) error {
	// create validator set from the local validator
	newTmVal := tmtypes.NewValidator(tmd25519.PubKey(args.validatorConsPubKeyByte), valVotingPower)
	vals := []*tmtypes.Validator{newTmVal}
	validatorSet := tmtypes.NewValidatorSet(vals)

	// CHANGE STATE CONSENSUS STORE
	stateDB, err := openDB(args.homeDir, "state", cometdbm.BackendType(server.GetAppDBBackend(appOpts)))
	if err != nil {
		return err
	}

	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: false,
	})

	// load state in order to change validator set info
	state, err := stateStore.Load()
	if err != nil {
		return err
	}
	defer func() {
		if derr := stateStore.Close(); derr != nil {
			logger.Error("Failed to close consensus state db", "err", derr)
			err = derr
		}
	}()

	state.LastValidators = validatorSet
	state.Validators = validatorSet
	state.NextValidators = validatorSet
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
	blockStoreDB, err := openDB(args.homeDir, "blockstore", cometdbm.BackendType(server.GetAppDBBackend(appOpts)))
	if err != nil {
		return err
	}
	defer func() {
		if derr := blockStoreDB.Close(); derr != nil {
			logger.Error("Failed to close consensus blockstore db", "err", derr)
			err = derr
		}
	}()

	blockStore := store.NewBlockStore(blockStoreDB)
	lastCommit := blockStore.LoadSeenCommit(state.LastBlockHeight)

	var vote *tmtypes.Vote
	for idx, commitSig := range lastCommit.Signatures {
		if commitSig.BlockIDFlag == tmtypes.BlockIDFlagAbsent {
			continue
		}
		vote = lastCommit.GetVote(int32(idx))
		break
	}
	if vote == nil {
		return errors.New("cannot get the vote from the last commit")
	}

	voteSignBytes := tmtypes.VoteSignBytes(state.ChainID, vote.ToProto())
	signatureBytes, err := args.validatorConsPrivKey.Sign(voteSignBytes)
	if err != nil {
		return err
	}

	lastCommit.Signatures = []tmtypes.CommitSig{{
		BlockIDFlag:      tmtypes.BlockIDFlagCommit,
		ValidatorAddress: newTmVal.Address,
		Timestamp:        vote.Timestamp,
		Signature:        []byte(signatureBytes),
	}}

	// if store height is greater than app height and state height, we will remove the last block from the store to avoid
	// replaying this block to the app. If only the state height is lower, we do not delete the block from the store because
	// block would not be replayed to the app (the mock app will be used by consensus instead) and only consensus state
	// will be updated when the node is run. If all three versions are equal, everything is ok. The only scenario from
	// which we cannot recover is the one when app height is lower that the version written in the app's iavl stores, because
	// the hash of the next block would not be the same as the written hash and we do not have an exported function to delete
	// that greater version from the app iavl store
	blockStoreState := store.LoadBlockStoreState(blockStoreDB)
	if blockStoreState.Height > state.LastBlockHeight && blockStoreState.Height > appHeight {
		blockStore.DeleteLatestBlock()
	}

	return blockStore.SaveSeenCommit(state.LastBlockHeight, lastCommit)
}

func loadValidatorsInfo(db cometdbm.DB, height int64) (*cmtstate.ValidatorsInfo, error) {
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

func saveValidatorsInfo(db cometdbm.DB, height int64, valInfo *cmtstate.ValidatorsInfo) error {
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

func openDB(rootDir, dbName string, backendType cometdbm.BackendType) (cometdbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return cometdbm.NewDB(dbName, backendType, dataDir)
}
