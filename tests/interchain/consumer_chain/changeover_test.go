package consumer_chain_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"golang.org/x/sync/errgroup"
)

type ChangeoverSuite struct {
	*chainsuite.Suite
	consumerCfg chainsuite.ConsumerConfig
	Consumer    *chainsuite.Chain
}

func (s *ChangeoverSuite) SetupSuite() {
	s.Suite.SetupSuite()
	validators := 1
	fullNodes := 0
	genesisChanges := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", chainsuite.GovVotingPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", chainsuite.GovDepositPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", chainsuite.Ucon),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", strconv.Itoa(chainsuite.GovMinDepositAmount)),
	}
	spec := &interchaintest.ChainSpec{
		Name:          "ics-consumer",
		ChainName:     "ics-consumer",
		Version:       "v6.4.0-rc0",
		NumValidators: &validators,
		NumFullNodes:  &fullNodes,
		ChainConfig: ibc.ChainConfig{
			Denom:         chainsuite.Ucon,
			GasPrices:     "0.025" + chainsuite.Ucon,
			GasAdjustment: 2.0,
			Gas:           "auto",
			ConfigFileOverrides: map[string]any{
				"config/config.toml": chainsuite.DefaultConfigToml(),
			},
			ModifyGenesisAmounts: chainsuite.DefaultGenesisAmounts(chainsuite.Ucon),
			ModifyGenesis:        cosmos.ModifyGenesis(genesisChanges),
			Bin:                  "interchain-security-sd",
			Images: []ibc.DockerImage{
				{
					Repository: chainsuite.HyphaICSRepo,
					Version:    "v6.4.0-rc0",
					UIDGID:     chainsuite.ICSUidGuid,
				},
			},
			Bech32Prefix: "consumer",
		},
	}
	consumer, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, spec)
	s.Require().NoError(err)

	s.Consumer = consumer

	s.UpgradeChain()
}

func (s *ChangeoverSuite) TestRewardsWithChangeover() {
	transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Consumer)
	s.Require().NoError(err)
	rewardDenom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", transferCh.ChannelID, s.Consumer.Config().Denom)).IBCDenom()

	s.Run("changeover", func() {
		s.changeSovereignToConsumer(s.Consumer, transferCh)
	})

	s.Run("rewards", func() {
		govAuthority, err := s.Chain.GetGovernanceAddress(s.GetContext())
		s.Require().NoError(err)
		rewardDenomsProp := providertypes.MsgChangeRewardDenoms{
			DenomsToAdd: []string{rewardDenom},
			Authority:   govAuthority,
		}
		prop, err := s.Chain.BuildProposal([]cosmos.ProtoMessage{&rewardDenomsProp},
			"add denoms to list of registered reward denoms",
			"add denoms to list of registered reward denoms",
			"", chainsuite.GovDepositAmount, "", false)
		s.Require().NoError(err)
		propResult, err := s.Chain.SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
		s.Require().NoError(err)
		s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propResult.ProposalID))

		faucetAddrBts, err := s.Consumer.GetAddress(s.GetContext(), interchaintest.FaucetAccountKeyName)
		s.Require().NoError(err)
		faucetAddr := types.MustBech32ifyAddressBytes(s.Consumer.Config().Bech32Prefix, faucetAddrBts)
		_, err = s.Consumer.Validators[0].ExecTx(
			s.GetContext(), interchaintest.FaucetAccountKeyName,
			"bank", "send", string(faucetAddr), s.Consumer.ValidatorWallets[0].Address,
			"1"+s.Consumer.Config().Denom, "--fees", "100000000"+s.Consumer.Config().Denom,
		)
		s.Require().NoError(err)

		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), chainsuite.BlocksPerDistribution+2, s.Chain, s.Consumer))
		s.Require().NoError(s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Consumer))
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain, s.Consumer))

		rewardStr, err := s.Chain.QueryJSON(
			s.GetContext(), fmt.Sprintf("total.#(%%\"*%s\")", rewardDenom),
			"distribution", "rewards", s.Chain.ValidatorWallets[0].Address,
		)
		s.Require().NoError(err)
		rewards, err := chainsuite.StrToSDKInt(rewardStr.String())
		s.Require().NoError(err)
		s.Require().True(rewards.GT(sdkmath.NewInt(0)), "rewards: %s", rewards.String())
	})
}

func TestChangeover(t *testing.T) {
	s := &ChangeoverSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.OneValidator,
			},
		}),
		consumerCfg: chainsuite.ConsumerConfig{
			ChainName:             "ics-consumer",
			ShouldCopyProviderKey: allProviderKeysCopied(),
			Denom:                 chainsuite.Ucon,
			TopN:                  0,
			AllowInactiveVals:     true,
			MinStake:              1_000_000,
		},
	}
	suite.Run(t, s)
}

func (s *ChangeoverSuite) changeSovereignToConsumer(consumer *chainsuite.Chain, transferCh *ibc.ChannelOutput) {
	cfg := s.consumerCfg
	currentHeight, err := consumer.Height(s.GetContext())
	s.Require().NoError(err)
	initialHeight := uint64(currentHeight) + 60
	cfg.InitialHeight = initialHeight
	spawnTime := time.Now().Add(60 * time.Second)
	cfg.DistributionTransmissionChannel = transferCh.ChannelID

	err = s.Chain.CreateConsumerPermissionless(s.GetContext(), consumer.Config().ChainID, cfg, spawnTime)
	s.Require().NoError(err)

	consumerChains, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	consumerChain := gjson.GetBytes(consumerChains, fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID))
	consumerID := consumerChain.Get("consumer_id").String()

	eg := errgroup.Group{}
	for i := range consumer.Validators {
		i := i
		eg.Go(func() error {
			key, _, err := consumer.Validators[i].ExecBin(s.GetContext(), "tendermint", "show-validator")
			if err != nil {
				return err
			}
			keyStr := strings.TrimSpace(string(key))
			_, err = s.Chain.Validators[i].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[i].Moniker, "provider", "opt-in", consumerID, keyStr)
			return err
		})
	}
	s.Require().NoError(eg.Wait())

	s.Require().NoError(err)
	time.Sleep(time.Until(spawnTime))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     "5000000" + chainsuite.Ucon,
		Title:       "Changeover",
		Name:        "sovereign-changeover",
		Description: "Changeover",
		Height:      int64(initialHeight) - 3,
	}
	upgradeTx, err := consumer.UpgradeProposal(s.GetContext(), interchaintest.FaucetAccountKeyName, proposal)
	s.Require().NoError(err)
	err = consumer.PassProposal(s.GetContext(), upgradeTx.ProposalID)
	s.Require().NoError(err)

	currentHeight, err = consumer.Height(s.GetContext())
	s.Require().NoError(err)

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(s.GetContext(), (time.Duration(int64(initialHeight)-currentHeight)+10)*chainsuite.CommitTimeout)
	defer timeoutCtxCancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(int64(initialHeight)-currentHeight)+3, consumer)
	s.Require().Error(err)

	s.Require().NoError(consumer.StopAllNodes(s.GetContext()))

	genesis, err := consumer.GetNode().GenesisFileContent(s.GetContext())
	s.Require().NoError(err)

	ccvState, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "provider", "consumer-genesis", consumerID)
	s.Require().NoError(err)
	genesis, err = sjson.SetRawBytes(genesis, "app_state.ccvconsumer", ccvState)
	s.Require().NoError(err)

	genesis, err = sjson.SetBytes(genesis, "app_state.slashing.params.signed_blocks_window", strconv.Itoa(chainsuite.SlashingWindowConsumer))
	s.Require().NoError(err)
	genesis, err = sjson.SetBytes(genesis, "app_state.ccvconsumer.params.reward_denoms", []string{chainsuite.Ucon})
	s.Require().NoError(err)
	genesis, err = sjson.SetBytes(genesis, "app_state.ccvconsumer.params.provider_reward_denoms", []string{s.Chain.Config().Denom})
	s.Require().NoError(err)
	genesis, err = sjson.SetBytes(genesis, "app_state.ccvconsumer.params.blocks_per_distribution_transmission", chainsuite.BlocksPerDistribution)
	s.Require().NoError(err)

	for _, val := range consumer.Validators {
		val := val
		eg.Go(func() error {
			if err := val.OverwriteGenesisFile(s.GetContext(), []byte(genesis)); err != nil {
				return err
			}
			return val.WriteFile(s.GetContext(), []byte(genesis), ".sovereign/config/genesis.json")
		})
	}
	s.Require().NoError(eg.Wait())

	consumer.ChangeBinary(s.GetContext(), "interchain-security-cdd")
	s.Require().NoError(consumer.StartAllNodes(s.GetContext()))
	s.Require().NoError(s.Relayer.ConnectProviderConsumer(s.GetContext(), s.Chain, consumer))
	s.Require().NoError(s.Relayer.StopRelayer(s.GetContext(), chainsuite.GetRelayerExecReporter(s.GetContext())))
	s.Require().NoError(s.Relayer.StartRelayer(s.GetContext(), chainsuite.GetRelayerExecReporter(s.GetContext())))
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))
}
