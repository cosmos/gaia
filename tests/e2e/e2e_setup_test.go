package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	// "github.com/cosmos/cosmos-sdk/crypto/hd"
	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmjson "github.com/cometbft/cometbft/libs/json"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"cosmossdk.io/math"
	evidencetypes "cosmossdk.io/x/evidence/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/tx"
)

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.commonHelper.Resources.ChainA, err = common.NewChain()
	s.Require().NoError(err)

	s.commonHelper.Resources.ChainB, err = common.NewChain()
	s.Require().NoError(err)

	s.commonHelper.Resources.DkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.commonHelper.Resources.DkrNet, err = s.commonHelper.Resources.DkrPool.CreateNetwork(fmt.Sprintf("%s-%s-testnet", s.commonHelper.Resources.ChainA.ID, s.commonHelper.Resources.ChainB.ID))
	s.Require().NoError(err)

	s.commonHelper.Resources.ValResources = make(map[string][]*dockertest.Resource)

	vestingMnemonic, err := common.CreateMnemonic()
	s.Require().NoError(err)

	jailedValMnemonic, err := common.CreateMnemonic()
	s.Require().NoError(err)

	// The bootstrapping phase is as follows:
	//
	// 1. Initialize Gaia validator nodes.
	// 2. Create and initialize Gaia validator genesis files (both chains)
	// 3. Start both networks.
	// 4. Create and run IBC relayer (Hermes) containers.

	s.T().Logf("starting e2e infrastructure for chain A; chain-id: %s; datadir: %s", s.commonHelper.Resources.ChainA.ID, s.commonHelper.Resources.ChainA.DataDir)
	s.initNodes(s.commonHelper.Resources.ChainA)
	s.initGenesis(s.commonHelper.Resources.ChainA, vestingMnemonic, jailedValMnemonic)
	s.initValidatorConfigs(s.commonHelper.Resources.ChainA)
	s.runValidators(s.commonHelper.Resources.ChainA, 0)

	s.T().Logf("starting e2e infrastructure for chain B; chain-id: %s; datadir: %s", s.commonHelper.Resources.ChainB.ID, s.commonHelper.Resources.ChainB.DataDir)
	s.initNodes(s.commonHelper.Resources.ChainB)
	s.initGenesis(s.commonHelper.Resources.ChainB, vestingMnemonic, jailedValMnemonic)
	s.initValidatorConfigs(s.commonHelper.Resources.ChainB)
	s.runValidators(s.commonHelper.Resources.ChainB, 10)

	s.commonHelper.TestCounters = common.TestCounters{
		ProposalCounter:           0,
		ContractsCounter:          0,
		IBCV2PacketSequence:       1,
		ContractsCounterPerSender: map[string]uint64{},
	}

	s.commonHelper.Suite = &s.Suite

	s.tx = tx.Helper{
		Suite:        &s.Suite,
		CommonHelper: &s.commonHelper,
	}

	time.Sleep(10 * time.Second)
	s.runIBCRelayer()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if str := os.Getenv("GAIA_E2E_SKIP_CLEANUP"); len(str) > 0 {
		skipCleanup, err := strconv.ParseBool(str)
		s.Require().NoError(err)

		if skipCleanup {
			return
		}
	}

	s.T().Log("tearing down e2e integration test suite...")

	s.Require().NoError(s.commonHelper.Resources.DkrPool.Purge(s.commonHelper.Resources.HermesResource))

	for _, vr := range s.commonHelper.Resources.ValResources {
		for _, r := range vr {
			s.Require().NoError(s.commonHelper.Resources.DkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.commonHelper.Resources.DkrPool.RemoveNetwork(s.commonHelper.Resources.DkrNet))

	os.RemoveAll(s.commonHelper.Resources.ChainA.DataDir)
	os.RemoveAll(s.commonHelper.Resources.ChainB.DataDir)

	for _, td := range s.commonHelper.Resources.TmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) initNodes(c *common.Chain) {
	s.Require().NoError(c.CreateAndInitValidators(2))
	/* Adding 4 accounts to val0 local directory
	c.GenesisAccounts[0]: Relayer Account
	c.GenesisAccounts[1]: ICA Owner
	c.GenesisAccounts[2]: Test Account 1
	c.GenesisAccounts[3]: Test Account 2
	*/
	s.Require().NoError(c.AddAccountFromMnemonic(5))
	// Initialize a genesis file for the first validator
	val0ConfigDir := c.Validators[0].ConfigDir()
	var addrAll []sdk.AccAddress
	for _, val := range c.Validators {
		addr, err := val.KeyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, addr)
	}

	for _, addr := range c.GenesisAccounts {
		acctAddr, err := addr.KeyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, acctAddr)
	}

	s.Require().NoError(
		modifyGenesis(val0ConfigDir, "", common.InitBalanceStr, addrAll, common.InitialBaseFeeAmt, common.UAtomDenom),
	)
	// copy the genesis file to the remaining validators
	for _, val := range c.Validators[1:] {
		_, err := common.CopyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.ConfigDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

// TODO find a better way to manipulate accounts to add genesis accounts
func (s *IntegrationTestSuite) addGenesisVestingAndJailedAccounts(
	c *common.Chain,
	valConfigDir,
	vestingMnemonic,
	jailedValMnemonic string,
	appGenState map[string]json.RawMessage,
) map[string]json.RawMessage {
	var (
		authGenState    = authtypes.GetGenesisStateFromAppState(common.Cdc, appGenState)
		bankGenState    = banktypes.GetGenesisStateFromAppState(common.Cdc, appGenState)
		stakingGenState = stakingtypes.GetGenesisStateFromAppState(common.Cdc, appGenState)
	)

	// create genesis vesting accounts keys
	kb, err := keyring.New(common.KeyringAppName, keyring.BackendTest, valConfigDir, nil, common.Cdc)
	s.Require().NoError(err)

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	s.Require().NoError(err)

	// create jailed validator account keys
	jailedValKey, err := kb.NewAccount(jailedValidatorKey, jailedValMnemonic, "", sdk.FullFundraiserPath, algo)
	s.Require().NoError(err)

	// create genesis vesting accounts keys
	c.GenesisVestingAccounts = make(map[string]sdk.AccAddress)
	for i, key := range genesisVestingKeys {
		// Use the first wallet from the same mnemonic by HD path
		acc, err := kb.NewAccount(key, vestingMnemonic, "", common.HDPath(i), algo)
		s.Require().NoError(err)
		c.GenesisVestingAccounts[key], err = acc.GetAddress()
		s.Require().NoError(err)
		s.T().Logf("created %s genesis account %s\n", key, c.GenesisVestingAccounts[key].String())
	}
	var (
		continuousVestingAcc = c.GenesisVestingAccounts[continuousVestingKey]
		delayedVestingAcc    = c.GenesisVestingAccounts[delayedVestingKey]
	)

	// add jailed validator to staking store
	pubKey, err := jailedValKey.GetPubKey()
	s.Require().NoError(err)

	jailedValAcc, err := jailedValKey.GetAddress()
	s.Require().NoError(err)

	jailedValAddr := sdk.ValAddress(jailedValAcc)
	val, err := stakingtypes.NewValidator(
		jailedValAddr.String(),
		pubKey,
		stakingtypes.NewDescription("jailed", "", "", "", ""),
	)
	s.Require().NoError(err)
	val.Jailed = true
	val.Tokens = math.NewInt(common.SlashingShares)
	val.DelegatorShares = math.LegacyNewDec(common.SlashingShares)
	stakingGenState.Validators = append(stakingGenState.Validators, val)

	// add jailed validator delegations
	stakingGenState.Delegations = append(stakingGenState.Delegations, stakingtypes.Delegation{
		DelegatorAddress: jailedValAcc.String(),
		ValidatorAddress: jailedValAddr.String(),
		Shares:           math.LegacyNewDec(common.SlashingShares),
	})

	appGenState[stakingtypes.ModuleName], err = common.Cdc.MarshalJSON(stakingGenState)
	s.Require().NoError(err)

	// add jailed account to the genesis
	baseJailedAccount := authtypes.NewBaseAccount(jailedValAcc, pubKey, 0, 0)
	s.Require().NoError(baseJailedAccount.Validate())

	// add continuous vesting account to the genesis
	baseVestingContinuousAccount := authtypes.NewBaseAccount(
		continuousVestingAcc, nil, 0, 0)
	baseVestingAcc, err := authvesting.NewBaseVestingAccount(
		baseVestingContinuousAccount,
		sdk.NewCoins(vestingAmountVested),
		time.Now().Add(time.Duration(rand.Intn(80)+150)*time.Second).Unix(),
	)
	s.Require().NoError(err)
	vestingContinuousGenAccount := authvesting.NewContinuousVestingAccountRaw(
		baseVestingAcc,
		time.Now().Add(time.Duration(rand.Intn(40)+90)*time.Second).Unix(),
	)
	s.Require().NoError(vestingContinuousGenAccount.Validate())

	// add delayed vesting account to the genesis
	baseVestingDelayedAccount := authtypes.NewBaseAccount(
		delayedVestingAcc, nil, 0, 0)
	baseVestingAcc, err = authvesting.NewBaseVestingAccount(
		baseVestingDelayedAccount,
		sdk.NewCoins(vestingAmountVested),
		time.Now().Add(time.Duration(rand.Intn(40)+90)*time.Second).Unix(),
	)
	s.Require().NoError(err)
	vestingDelayedGenAccount := authvesting.NewDelayedVestingAccountRaw(baseVestingAcc)
	s.Require().NoError(vestingDelayedGenAccount.Validate())

	// unpack and append accounts
	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	s.Require().NoError(err)
	accs = append(accs, vestingContinuousGenAccount, vestingDelayedGenAccount, baseJailedAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)
	genAccs, err := authtypes.PackAccounts(accs)
	s.Require().NoError(err)
	authGenState.Accounts = genAccs

	// update auth module state
	appGenState[authtypes.ModuleName], err = common.Cdc.MarshalJSON(&authGenState)
	s.Require().NoError(err)

	// update balances
	vestingContinuousBalances := banktypes.Balance{
		Address: continuousVestingAcc.String(),
		Coins:   vestingBalance,
	}
	vestingDelayedBalances := banktypes.Balance{
		Address: delayedVestingAcc.String(),
		Coins:   vestingBalance,
	}
	jailedValidatorBalances := banktypes.Balance{
		Address: jailedValAcc.String(),
		Coins:   sdk.NewCoins(common.TokenAmount),
	}
	stakingModuleBalances := banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(common.UAtomDenom, math.NewInt(common.SlashingShares))),
	}
	bankGenState.Balances = append(
		bankGenState.Balances,
		vestingContinuousBalances,
		vestingDelayedBalances,
		jailedValidatorBalances,
		stakingModuleBalances,
	)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	// update the denom metadata for the bank module
	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     common.UAtomDenom,
		Base:        common.UAtomDenom,
		Symbol:      common.UAtomDenom,
		Name:        common.UAtomDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    common.UAtomDenom,
				Exponent: 0,
			},
		},
	})

	// update bank module state
	appGenState[banktypes.ModuleName], err = common.Cdc.MarshalJSON(bankGenState)
	s.Require().NoError(err)

	return appGenState
}

func (s *IntegrationTestSuite) initGenesis(c *common.Chain, vestingMnemonic, jailedValMnemonic string) {
	var (
		serverCtx = server.NewDefaultContext()
		config    = serverCtx.Config
		validator = c.Validators[0]
	)

	config.SetRoot(validator.ConfigDir())
	config.Moniker = validator.Moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	appGenState = s.addGenesisVestingAndJailedAccounts(
		c,
		validator.ConfigDir(),
		vestingMnemonic,
		jailedValMnemonic,
		appGenState,
	)

	var evidenceGenState evidencetypes.GenesisState
	s.Require().NoError(common.Cdc.UnmarshalJSON(appGenState[evidencetypes.ModuleName], &evidenceGenState))

	evidenceGenState.Evidence = make([]*codectypes.Any, common.NumberOfEvidences)
	for i := range evidenceGenState.Evidence {
		pk := ed25519.GenPrivKey()
		evidence := &evidencetypes.Equivocation{
			Height:           1,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()).String(),
		}
		evidenceGenState.Evidence[i], err = codectypes.NewAnyWithValue(evidence)
		s.Require().NoError(err)
	}

	appGenState[evidencetypes.ModuleName], err = common.Cdc.MarshalJSON(&evidenceGenState)
	s.Require().NoError(err)

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(common.Cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.Validators))
	for i, val := range c.Validators {
		createValmsg, err := val.BuildCreateValidatorMsg(common.StakingAmountCoin)
		s.Require().NoError(err)
		signedTx, err := val.SignMsg(createValmsg)

		s.Require().NoError(err)

		txRaw, err := common.Cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	appGenState[genutiltypes.ModuleName], err = common.Cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)

	genDoc.AppState, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	bz, err := tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	vestingPeriod, err := generateVestingPeriod()
	s.Require().NoError(err)

	rawTx, _, err := buildRawTx()
	s.Require().NoError(err)

	// write the updated genesis file to each validator.
	for _, val := range c.Validators {
		err = common.WriteFile(filepath.Join(val.ConfigDir(), "config", "genesis.json"), bz)
		s.Require().NoError(err)

		err = common.WriteFile(filepath.Join(val.ConfigDir(), vestingPeriodFile), vestingPeriod)
		s.Require().NoError(err)

		err = common.WriteFile(filepath.Join(val.ConfigDir(), rawTxFile), rawTx)
		s.Require().NoError(err)
	}
}

// initValidatorConfigs initializes the validator configs for the given Chain.
func (s *IntegrationTestSuite) initValidatorConfigs(c *common.Chain) {
	for i, val := range c.Validators {
		tmCfgPath := filepath.Join(val.ConfigDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()

		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.InstanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.Validators); j++ {
			if i == j {
				continue
			}

			peer := c.Validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.NodeKey.ID(), peer.Moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.ConfigDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.API.Address = "tcp://0.0.0.0:1317"
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", common.MinGasPrice, common.UAtomDenom)
		appConfig.GRPC.Address = "0.0.0.0:9090"

		srvconfig.SetConfigTemplate(srvconfig.DefaultConfigTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
}

// runValidators runs the validators in the Chain
func (s *IntegrationTestSuite) runValidators(c *common.Chain, portOffset int) {
	s.T().Logf("starting Gaia %s validator containers...", c.ID)

	s.commonHelper.Resources.ValResources[c.ID] = make([]*dockertest.Resource, len(c.Validators))
	for i, val := range c.Validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.InstanceName(),
			NetworkID: s.commonHelper.Resources.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:%s", val.ConfigDir(), common.GaiaHomePath),
			},
			Repository: "cosmos/gaiad-e2e",
		}

		s.Require().NoError(exec.Command("chmod", "-R", "0777", val.ConfigDir()).Run()) //nolint:gosec // this is a test

		// expose the first validator for debugging and communication
		if val.Index == 0 {
			runOpts.PortBindings = map[docker.Port][]docker.PortBinding{
				"1317/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 1317+portOffset)}},
				"6060/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6060+portOffset)}},
				"6061/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6061+portOffset)}},
				"6062/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6062+portOffset)}},
				"6063/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6063+portOffset)}},
				"6064/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6064+portOffset)}},
				"6065/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 6065+portOffset)}},
				"9090/tcp":  {{HostIP: "", HostPort: fmt.Sprintf("%d", 9090+portOffset)}},
				"26656/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26656+portOffset)}},
				"26657/tcp": {{HostIP: "", HostPort: fmt.Sprintf("%d", 26657+portOffset)}},
			}
		}

		resource, err := s.commonHelper.Resources.DkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.commonHelper.Resources.ValResources[c.ID][i] = resource
		s.T().Logf("started Gaia %s validator container: %s", c.ID, resource.Container.ID)
	}

	rpcClient, err := rpchttp.New("tcp://localhost:26657", "/websocket")
	s.Require().NoError(err)

	s.Require().Eventually(
		func() bool {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			status, err := rpcClient.Status(ctx)
			if err != nil {
				return false
			}

			// let the node produce a few blocks
			if status.SyncInfo.CatchingUp || status.SyncInfo.LatestBlockHeight < 3 {
				return false
			}
			return true
		},
		5*time.Minute,
		time.Second,
		"Gaia node failed to produce blocks",
	)
}

func noRestart(config *docker.HostConfig) {
	// in this case we don't want the nodes to restart on failure
	config.RestartPolicy = docker.RestartPolicy{
		Name: "no",
	}
}

// runIBCRelayer bootstraps an IBC Hermes relayer by creating an IBC connection and
// a transfer channel between ChainA and ChainB.
func (s *IntegrationTestSuite) runIBCRelayer() {
	s.T().Log("starting Hermes relayer container")

	tmpDir, err := os.MkdirTemp("", "gaia-e2e-testnet-hermes-")
	s.Require().NoError(err)
	s.commonHelper.Resources.TmpDirs = append(s.commonHelper.Resources.TmpDirs, tmpDir)

	gaiaAVal := s.commonHelper.Resources.ChainA.Validators[0]
	gaiaBVal := s.commonHelper.Resources.ChainB.Validators[0]

	gaiaARly := s.commonHelper.Resources.ChainA.GenesisAccounts[common.RelayerAccountIndexHermes]
	gaiaBRly := s.commonHelper.Resources.ChainB.GenesisAccounts[common.RelayerAccountIndexHermes]

	hermesCfgPath := path.Join(tmpDir, "hermes")

	s.Require().NoError(os.MkdirAll(hermesCfgPath, 0o755))
	_, err = common.CopyFile(
		filepath.Join("./scripts/", "hermes_bootstrap.sh"),
		filepath.Join(hermesCfgPath, "hermes_bootstrap.sh"),
	)
	s.Require().NoError(err)

	s.commonHelper.Resources.HermesResource, err = s.commonHelper.Resources.DkrPool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       fmt.Sprintf("%s-%s-relayer", s.commonHelper.Resources.ChainA.ID, s.commonHelper.Resources.ChainB.ID),
			Repository: "ghcr.io/cosmos/hermes-e2e",
			Tag:        "1.0.0",
			NetworkID:  s.commonHelper.Resources.DkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:/root/hermes", hermesCfgPath),
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"3031/tcp": {{HostIP: "", HostPort: "3031"}},
			},
			Env: []string{
				fmt.Sprintf("GAIA_A_E2E_CHAIN_ID=%s", s.commonHelper.Resources.ChainA.ID),
				fmt.Sprintf("GAIA_B_E2E_CHAIN_ID=%s", s.commonHelper.Resources.ChainB.ID),
				fmt.Sprintf("GAIA_A_E2E_VAL_MNEMONIC=%s", gaiaAVal.Mnemonic),
				fmt.Sprintf("GAIA_B_E2E_VAL_MNEMONIC=%s", gaiaBVal.Mnemonic),
				fmt.Sprintf("GAIA_A_E2E_RLY_MNEMONIC=%s", gaiaARly.Mnemonic),
				fmt.Sprintf("GAIA_B_E2E_RLY_MNEMONIC=%s", gaiaBRly.Mnemonic),
				fmt.Sprintf("GAIA_A_E2E_VAL_HOST=%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainA.ID][0].Container.Name[1:]),
				fmt.Sprintf("GAIA_B_E2E_VAL_HOST=%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainB.ID][0].Container.Name[1:]),
			},
			User: "root",
			Entrypoint: []string{
				"sh",
				"-c",
				"chmod +x /root/hermes/hermes_bootstrap.sh && /root/hermes/hermes_bootstrap.sh && tail -f /dev/null",
			},
		},
		noRestart,
	)
	s.Require().NoError(err)

	s.T().Logf("started Hermes relayer container: %s", s.commonHelper.Resources.HermesResource.Container.ID)

	// XXX: Give time to both networks to start, otherwise we might see gRPC
	// transport errors.
	time.Sleep(10 * time.Second)

	// create the client, connection and channel between the two Gaia chains
	s.commonHelper.CreateConnection()
	s.commonHelper.CreateChannel()
}

func configFile(filename string) string {
	filepath := filepath.Join(common.GaiaConfigPath, filename)
	return filepath
}
