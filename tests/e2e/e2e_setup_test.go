package e2e

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gaia/v8/app/params"
	ibcclienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	gaiadBinary    = "gaiad"
	gaiaHomePath   = "/home/nonroot/.gaia"
	photonDenom    = "photon"
	uatomDenom     = "uatom"
	initBalanceStr = "110000000000stake,100000000000000000photon,100000000000000000uatom"
	minGasPrice    = "0.00001"
	// the test globalfee in genesis is the same as minGasPrice
	// global fee lower/higher than min_gas_price
	initialGlobalFeeAmt        = "0.00001"
	lowGlobalFeesAmt           = "0.000001"
	highGlobalFeeAmt           = "0.0001"
	gas                        = 200000
	govSendMsgRecipientAddress = "cosmos1pkueemdeps77dwrqma03pwqk93nw39nuhccz02"
	govProposalBlockBuffer     = 35
	relayerAccountIndex        = 0
	icaOwnerAccountIndex       = 1
)

var (
	gaiaConfigPath    = filepath.Join(gaiaHomePath, "config")
	stakingAmount     = math.NewInt(100000000000)
	stakingAmountCoin = sdk.NewCoin(uatomDenom, stakingAmount)
	tokenAmount       = sdk.NewCoin(uatomDenom, math.NewInt(3300000000)) // 3,300uatom
	fees              = sdk.NewCoin(uatomDenom, math.NewInt(330000))     // 0.33uatom
	depositAmount     = sdk.NewCoin(uatomDenom, math.NewInt(10000000))   // 10uatom
	distModuleAddress = authtypes.NewModuleAddress(distrtypes.ModuleName).String()
	govModuleAddress  = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	proposalCounter   = 0
	sendGovAmount     = sdk.NewInt64Coin(uatomDenom, 10)
)

type UpgradePlan struct {
	Name   string `json:"name"`
	Height int    `json:"height"`
	Info   string `json:"info"`
}

type SoftwareUpgrade struct {
	Type      string      `json:"@type"`
	Authority string      `json:"authority"`
	Plan      UpgradePlan `json:"plan"`
}

type CancelSoftwareUpgrade struct {
	Type      string `json:"@type"`
	Authority string `json:"authority"`
}

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	chainA         *chain
	chainB         *chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource
	valResources   map[string][]*dockertest.Resource
}

type AddressResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

type GroupMember struct {
	Address  string `json:"address"`
	Weight   string `json:"weight"`
	Metadata string `json:"metadata"`
}

type MsgSend struct {
	Type   string     `json:"@type"`
	From   string     `json:"from_address"`
	To     string     `json:"to_address"`
	Amount []sdk.Coin `json:"amount"`
}

type ThresholdPolicy struct {
	Type      string               `json:"@type"`
	Threshold string               `json:"threshold"`
	Windows   DecisionPolicyWindow `json:"windows"`
}

type PercentagePolicy struct {
	Type       string               `json:"@type"`
	Percentage string               `json:"percentage"`
	Windows    DecisionPolicyWindow `json:"windows"`
}

type DecisionPolicyWindow struct {
	VotingPeriod       string `json:"voting_period"`
	MinExecutionPeriod string `json:"min_execution_period"`
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")

	var err error
	s.chainA, err = newChain()
	s.Require().NoError(err)

	s.chainB, err = newChain()
	s.Require().NoError(err)

	s.dkrPool, err = dockertest.NewPool("")
	s.Require().NoError(err)

	s.dkrNet, err = s.dkrPool.CreateNetwork(fmt.Sprintf("%s-%s-testnet", s.chainA.id, s.chainB.id))
	s.Require().NoError(err)

	s.valResources = make(map[string][]*dockertest.Resource)

	// The boostrapping phase is as follows:
	//
	// 1. Initialize Gaia validator nodes.
	// 2. Create and initialize Gaia validator genesis files (both chains)
	// 3. Start both networks.
	// 4. Create and run IBC relayer (Hermes) containers.

	s.T().Logf("starting e2e infrastructure for chain A; chain-id: %s; datadir: %s", s.chainA.id, s.chainA.dataDir)
	s.initNodes(s.chainA)
	s.initGenesis(s.chainA)
	s.initValidatorConfigs(s.chainA)
	s.runValidators(s.chainA, 0)

	s.T().Logf("starting e2e infrastructure for chain B; chain-id: %s; datadir: %s", s.chainB.id, s.chainB.dataDir)
	s.initNodes(s.chainB)
	s.initGenesis(s.chainB)
	s.initValidatorConfigs(s.chainB)
	s.runValidators(s.chainB, 10)

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

	s.Require().NoError(s.dkrPool.Purge(s.hermesResource))

	for _, vr := range s.valResources {
		for _, r := range vr {
			s.Require().NoError(s.dkrPool.Purge(r))
		}
	}

	s.Require().NoError(s.dkrPool.RemoveNetwork(s.dkrNet))

	os.RemoveAll(s.chainA.dataDir)
	os.RemoveAll(s.chainB.dataDir)

	for _, td := range s.tmpDirs {
		os.RemoveAll(td)
	}
}

func (s *IntegrationTestSuite) initNodes(c *chain) {
	s.Require().NoError(c.createAndInitValidators(2))
	/* Adding 4 accounts to val0 local directory
	c.genesisAccounts[0]: Relayer Wallet
	c.genesisAccounts[1]: ICA Owner
	c.genesisAccounts[2]: Test Account 1
	c.genesisAccounts[3]: Test Account 2
	*/
	s.Require().NoError(c.addAccountFromMnemonic(4))
	// Initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	var addrAll []sdk.AccAddress
	for _, val := range c.validators {
		address, err := val.keyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, address)
	}

	for _, addr := range c.genesisAccounts {
		acctAddr, err := addr.keyInfo.GetAddress()
		s.Require().NoError(err)
		addrAll = append(addrAll, acctAddr)
	}

	s.Require().NoError(
		modifyGenesis(val0ConfigDir, "", initBalanceStr, addrAll, initialGlobalFeeAmt+uatomDenom, uatomDenom),
	)
	// copy the genesis file to the remaining validators
	for _, val := range c.validators[1:] {
		_, err := copyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) initGenesis(c *chain) {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(c.validators[0].configDir())
	config.Moniker = c.validators[0].moniker

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	s.Require().NoError(err)

	var bankGenState banktypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState))

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     uatomDenom,
		Base:        uatomDenom,
		Symbol:      uatomDenom,
		Name:        uatomDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    uatomDenom,
				Exponent: 0,
			},
		},
	})

	bz, err := cdc.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	appGenState[banktypes.ModuleName] = bz

	var genUtilGenState genutiltypes.GenesisState
	s.Require().NoError(cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState))

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.validators))
	for i, val := range c.validators {
		createValmsg, err := val.buildCreateValidatorMsg(stakingAmountCoin)
		s.Require().NoError(err)
		signedTx, err := val.signMsg(createValmsg)

		s.Require().NoError(err)

		txRaw, err := cdc.MarshalJSON(signedTx)
		s.Require().NoError(err)

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err = cdc.MarshalJSON(&genUtilGenState)
	s.Require().NoError(err)
	appGenState[genutiltypes.ModuleName] = bz

	bz, err = json.MarshalIndent(appGenState, "", "  ")
	s.Require().NoError(err)

	genDoc.AppState = bz

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	s.Require().NoError(err)

	// write the updated genesis file to each validator.
	for _, val := range c.validators {
		err = writeFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz)
		s.Require().NoError(err)
	}
}

// initValidatorConfigs initializes the validator configs for the given chain.
func (s *IntegrationTestSuite) initValidatorConfigs(c *chain) {
	for i, val := range c.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		s.Require().NoError(vpr.ReadInConfig())

		valConfig := tmconfig.DefaultConfig()

		s.Require().NoError(vpr.Unmarshal(valConfig))

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		var peers []string

		for j := 0; j < len(c.validators); j++ {
			if i == j {
				continue
			}

			peer := c.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.nodeKey.ID(), peer.moniker, j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")

		appConfig := srvconfig.DefaultConfig()
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", minGasPrice, uatomDenom)

		//	 srvconfig.WriteConfigFile(appCfgPath, appConfig)
		appCustomConfig := params.CustomAppConfig{
			Config: *appConfig,
			BypassMinFeeMsgTypes: []string{
				// todo: use ibc as exmaple ?
				sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
				sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
				sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
			},
		}

		customAppTemplate := `
###############################################################################
###                        Custom Gaia Configuration                        ###
###############################################################################
# bypass-min-fee-msg-types defines custom message types the operator may set that
# will bypass minimum fee checks during CheckTx.
#
# Example:
# ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement", ...]
bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward","/ibc.applications.transfer.v1.MsgTransfer"]
` + srvconfig.DefaultConfigTemplate
		srvconfig.SetConfigTemplate(customAppTemplate)
		srvconfig.WriteConfigFile(appCfgPath, appCustomConfig)
	}
}

// runValidators runs the validators in the chain
func (s *IntegrationTestSuite) runValidators(c *chain, portOffset int) {
	s.T().Logf("starting Gaia %s validator containers...", c.id)

	s.valResources[c.id] = make([]*dockertest.Resource, len(c.validators))
	for i, val := range c.validators {
		runOpts := &dockertest.RunOptions{
			Name:      val.instanceName(),
			NetworkID: s.dkrNet.Network.ID,
			Mounts: []string{
				fmt.Sprintf("%s/:%s", val.configDir(), gaiaHomePath),
			},
			Repository: "cosmos/gaiad-e2e",
		}

		s.Require().NoError(exec.Command("chmod", "-R", "0777", val.configDir()).Run())

		// expose the first validator for debugging and communication
		if val.index == 0 {
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

		resource, err := s.dkrPool.RunWithOptions(runOpts, noRestart)
		s.Require().NoError(err)

		s.valResources[c.id][i] = resource
		s.T().Logf("started Gaia %s validator container: %s", c.id, resource.Container.ID)
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

func (s *IntegrationTestSuite) writeGovProposals(c *chain) {
	type GovMessageSend struct {
		Type   string     `json:"@type"`
		From   string     `json:"from_address"`
		To     string     `json:"to_address"`
		Amount []sdk.Coin `json:"amount"`
	}

	msgSendMessages := []GovMessageSend{
		{
			Type:   "/cosmos.bank.v1beta1.MsgSend",
			From:   govModuleAddress,
			To:     govSendMsgRecipientAddress,
			Amount: []sdk.Coin{sendGovAmount},
		},
	}

	msgSendBody, err := json.MarshalIndent(struct {
		Messages []GovMessageSend `json:"messages"`
		Metadata string           `json:"metadata"`
		Deposit  string           `json:"deposit"`
	}{
		Messages: msgSendMessages,
		Metadata: b64.StdEncoding.EncodeToString([]byte("Testing 1, 2, 3!")),
		Deposit:  "5000uatom",
	}, "", " ")

	s.Require().NoError(err)

	legacyCommunitySpendBody, err := json.MarshalIndent(struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Recipient   string `json:"recipient"`
		Amount      string `json:"amount"`
		Deposit     string `json:"deposit"`
	}{
		Title:       "Community Pool Spend",
		Description: "Fund Gov !",
		Recipient:   govModuleAddress,
		Amount:      "1000uatom",
		Deposit:     "5000uatom",
	}, "", " ")

	s.Require().NoError(err)

	for _, val := range c.validators {
		err = writeFile(filepath.Join(val.configDir(), "config", "proposal.json"), legacyCommunitySpendBody)
		s.Require().NoError(err)

		err = writeFile(filepath.Join(val.configDir(), "config", "proposal_2.json"), msgSendBody)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) writeGovUpgradeSoftwareProposal(c *chain, height int) {
	softwareUpgradeMessages := []SoftwareUpgrade{
		{
			Type:      "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
			Authority: govModuleAddress,
			Plan: UpgradePlan{
				Name:   "upgrade-1",
				Height: height,
				Info:   "binary-1",
			},
		},
	}
	cancelSoftwareUpgradeMessages := []CancelSoftwareUpgrade{
		{
			Type:      "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
			Authority: govModuleAddress,
		},
	}

	upgradeProposalBody, err := json.MarshalIndent(struct {
		Messages []SoftwareUpgrade `json:"messages"`
		Metadata string            `json:"metadata"`
		Deposit  string            `json:"deposit"`
	}{
		Messages: softwareUpgradeMessages,
		Metadata: b64.StdEncoding.EncodeToString([]byte("Testing 1, 2, 3!")),
		Deposit:  "5000uatom",
	}, "", " ")

	cancelUpgradeProposalBody, err := json.MarshalIndent(struct {
		Messages []CancelSoftwareUpgrade `json:"messages"`
		Metadata string                  `json:"metadata"`
		Deposit  string                  `json:"deposit"`
	}{
		Messages: cancelSoftwareUpgradeMessages,
		Metadata: "VGVzdGluZyAxLCAyLCAzIQ==",
		Deposit:  "5000uatom",
	}, "", " ")

	err = writeFile(filepath.Join(c.validators[0].configDir(), "config", "proposal_3.json"), upgradeProposalBody)
	s.Require().NoError(err)

	err = writeFile(filepath.Join(c.validators[0].configDir(), "config", "proposal_4.json"), cancelUpgradeProposalBody)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeGroupMembers(c *chain, groupMembers []GroupMember, filename string) {
	groupMembersBody, err := json.MarshalIndent(struct {
		Members []GroupMember `json:"members"`
	}{
		Members: groupMembers,
	}, "", " ")
	s.Require().NoError(err)

	s.writeFile(c, filename, groupMembersBody)
}

func (s *IntegrationTestSuite) writeFile(c *chain, filename string, body []byte) {
	for _, val := range c.validators {
		err := writeFile(filepath.Join(val.configDir(), "config", filename), body)
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) writeGovParamChangeProposalGlobalFees(c *chain, coins sdk.DecCoins) {
	type ParamInfo struct {
		Subspace string       `json:"subspace"`
		Key      string       `json:"key"`
		Value    sdk.DecCoins `json:"value"`
	}

	type ParamChangeMessage struct {
		Title       string      `json:"title"`
		Description string      `json:"description"`
		Changes     []ParamInfo `json:"changes"`
		Deposit     string      `json:"deposit"`
	}

	paramChangeProposalBody, err := json.MarshalIndent(ParamChangeMessage{
		Title:       "global fee test",
		Description: "global fee change",
		Changes: []ParamInfo{
			{
				Subspace: "globalfee",
				Key:      "MinimumGasPricesParam",
				Value:    coins,
			},
		},
		Deposit: "",
	}, "", " ")
	s.Require().NoError(err)

	err = writeFile(filepath.Join(c.validators[0].configDir(), "config", "proposal_globalfee.json"), paramChangeProposalBody)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeICAtx(cmd []string, path string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd = append(cmd, fmt.Sprintf("--%s=%s", flags.FlagGenerateOnly, "true"))
	s.T().Logf("dry run: ica tx %s", strings.Join(cmd, " "))

	type txResponse struct {
		Body struct {
			Messages []map[string]interface{}
		}
	}

	s.executeGaiaTxCommand(ctx, s.chainA, cmd, 0, func(stdOut []byte, stdErr []byte) bool {
		var txResp txResponse
		s.Require().NoError(json.Unmarshal(stdOut, &txResp))
		b, err := json.MarshalIndent(txResp.Body.Messages[0], "", " ")
		s.Require().NoError(err)

		err = writeFile(path, b)
		s.Require().NoError(err)
		return true
	})

	s.T().Logf("write ica transaction json to %s", path)
}

func configFile(filename string) string {
	return filepath.Join(gaiaConfigPath, filename)
}
