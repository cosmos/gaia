package integrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
)

const (
	v26ImageVersion         = "v26.0.0"
	consumerRewardsPoolAddr = "cosmos1ap0mh6xzfn8943urr84q6ae7zfnar48am2erhd"
)

type ICSOffboardSuite struct {
	*chainsuite.Suite
	consumer              *chainsuite.Chain
	rewardsPoolPreUpgrade map[string]string // denom -> integer amount captured before v28 upgrade
}

func TestICSOffboard(t *testing.T) {
	s := &ICSOffboardSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			Scope:          chainsuite.ChainScopeSuite,
			CreateRelayer:  true,
			UpgradeOnSetup: false,
			ChainSpec: &interchaintest.ChainSpec{
				Version:       v26ImageVersion,
				NumValidators: &chainsuite.SixValidators,
			},
		}),
	}
	suite.Run(t, s)
}

func (s *ICSOffboardSuite) TestICSOffboardFlow() {
	ctx := s.GetContext()

	// Phase 1: Create consumer chain on v26
	s.Run("CreateConsumer_v26", func() {
		cfg := chainsuite.ConsumerConfig{
			ChainName:             "ics-consumer",
			Version:               "v4.5.0",
			ShouldCopyProviderKey: allProviderKeysCopied(),
			Denom:                 chainsuite.Ucon,
			TopN:                  94,
			Spec: &interchaintest.ChainSpec{
				ChainConfig: ibc.ChainConfig{
					Images: []ibc.DockerImage{
						{
							Repository: chainsuite.HyphaICSRepo,
							Version:    "v4.5.0",
							UIDGID:     chainsuite.ICSUidGuid,
						},
					},
				},
			},
		}
		consumer, err := s.Chain.AddConsumerChain(ctx, s.Relayer, cfg)
		s.Require().NoError(err)
		s.consumer = consumer
	})

	// Phase 2: Verify CCV works on v26
	s.Run("VerifyCCV_v26", func() {
		err := s.Chain.CheckCCV(ctx, s.consumer, s.Relayer, 1_000_000, 0, 1)
		s.Require().NoError(err)
	})

	// Phase 3: Verify consumer is listed on v26
	s.Run("ConsumerListed_v26", func() {
		s.requireConsumerListed(ctx)
	})

	// Phase 4: Add ICS messages to the chain state
	s.Run("UpdateProviderParams_v26", func() {
		// Query current provider module params (the response body IS the params object)
		rawParams, _, err := s.Chain.Validators[0].ExecQuery(ctx, "provider", "params")
		s.Require().NoError(err)

		// Unmarshal into a map so we can modify a single field
		var paramsMap map[string]interface{}
		s.Require().NoError(json.Unmarshal(rawParams, &paramsMap))
		paramsMap["number_of_epochs_to_start_receiving_rewards"] = "5"

		updatedParamsBytes, err := json.Marshal(paramsMap)
		s.Require().NoError(err)

		// Get the governance authority address
		authority, err := s.Chain.GetGovernanceAddress(ctx)
		s.Require().NoError(err)

		// Build the MsgUpdateParams message
		msgUpdateParams := fmt.Sprintf(`{
			"@type": "/interchain_security.ccv.provider.v1.MsgUpdateParams",
			"authority": "%s",
			"params": %s
		}`, authority, string(updatedParamsBytes))

		// Build and submit (but do not pass) the proposal
		prop, err := s.Chain.BuildProposal(
			nil,
			"Update Provider Params",
			"Proposal to change number_of_epochs_to_start_receiving_rewards to 5",
			"ipfs://CID",
			chainsuite.GovDepositAmount,
			s.Chain.ValidatorWallets[0].Moniker,
			false,
		)
		s.Require().NoError(err)
		prop.Messages = []json.RawMessage{json.RawMessage(msgUpdateParams)}

		_, err = s.Chain.SubmitProposal(ctx, s.Chain.ValidatorWallets[0].Moniker, prop)
		s.Require().NoError(err)
	})

	// Phase 5: Upgrade to v27
	s.Run("UpgradeToV27", func() {
		err := s.Chain.Upgrade(ctx, s.Env.OldGaiaImageVersion, s.Env.OldGaiaImageVersion)
		s.Require().NoError(err)
		s.restartRelayer(ctx)
	})

	// Phase 6: Consumer still listed on v27
	s.Run("ConsumerListed_v27", func() {
		s.requireConsumerListed(ctx)
	})

	// Phase 7: CCV still works on v27
	s.Run("VerifyCCV_v27", func() {
		err := s.Chain.CheckCCV(ctx, s.consumer, s.Relayer, 1_000_000, 0, 1)
		s.Require().NoError(err)
	})

	// Phase 8: Creating new consumer should be disabled on v27
	s.Run("CreateConsumerDisabled_v27", func() {
		s.requireCreateConsumerBlocked(ctx)
	})

	// Phase 9: IBC transfer still works on v27
	s.Run("IBCTransfer_v27", func() {
		err := chainsuite.SendSimpleIBCTx(ctx, s.Chain, s.consumer, s.Relayer)
		s.Require().NoError(err)
	})

	// Phase 10: Upgrade to v28
	s.Run("UpgradeToV28", func() {
		err := s.Chain.Upgrade(ctx, s.Env.UpgradeName, s.Env.NewGaiaImageVersion)
		s.Require().NoError(err)
		s.restartRelayer(ctx)

		// Query the applied upgrade plan to get the actual upgrade height.
		planOut, _, err := s.Chain.Validators[0].ExecQuery(ctx, "upgrade", "applied", s.Env.UpgradeName)
		s.Require().NoError(err)
		upgradeHeight := gjson.GetBytes(planOut, "height").Int()
		s.Require().Positive(upgradeHeight, "upgrade plan height must be positive")
		s.T().Logf("Upgrade %q applied at height %d", s.Env.UpgradeName, upgradeHeight)

		// Query consumer rewards pool at upgradeHeight-1: this is the exact committed
		// state that the upgrade handler's BeginBlock read, with no race from late arrivals.
		out, _, err := s.Chain.Validators[0].ExecQuery(ctx, "bank", "balances", consumerRewardsPoolAddr,
			"--height", fmt.Sprintf("%d", upgradeHeight-1))
		s.Require().NoError(err)
		s.rewardsPoolPreUpgrade = make(map[string]string)
		gjson.GetBytes(out, "balances").ForEach(func(_, coin gjson.Result) bool {
			s.rewardsPoolPreUpgrade[coin.Get("denom").String()] = coin.Get("amount").String()
			return true
		})
		s.T().Logf("Consumer rewards pool balance at height %d (pre-upgrade): %v", upgradeHeight-1, s.rewardsPoolPreUpgrade)
	})

	// Phase 11: Chain liveness after v28
	s.Run("Liveness_v28", func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		s.Require().NoError(testutil.WaitForBlocks(timeoutCtx, 5, s.Chain))
	})

	// Phase 12: Consumer offboarded — provider module removed
	s.Run("ConsumerOffboarded_v28", func() {
		_, _, err := s.Chain.Validators[0].ExecQuery(ctx, "provider", "list-consumer-chains")
		if err != nil {
			// Provider query command doesn't exist — module removed. This is expected.
			return
		}
		// If the query somehow succeeds, the consumer should not be listed
		s.requireConsumerNotListed(ctx)
	})

	// Phase 13: IBC transfers still work after offboarding
	s.Run("IBCTransfer_v28", func() {
		err := chainsuite.SendSimpleIBCTx(ctx, s.Chain, s.consumer, s.Relayer)
		s.Require().NoError(err)
	})

	// Phase 14: Provider port channels must be closed after offboarding
	s.Run("ProviderPortChannelsClosed_v28", func() {
		out, _, err := s.Chain.Validators[0].ExecQuery(ctx, "ibc", "channel", "channels")
		s.Require().NoError(err)
		providerChannels := gjson.GetBytes(out, `channels.#(port_id=="provider")#`)
		s.Require().True(providerChannels.Exists() && len(providerChannels.Array()) > 0,
			"expected at least one channel on the provider port")
		for _, ch := range providerChannels.Array() {
			state := ch.Get("state").String()
			channelID := ch.Get("channel_id").String()
			s.Require().Equal("STATE_CLOSED", state,
				"expected channel %s on provider port to be STATE_CLOSED after upgrade", channelID)
		}
	})

	// Phase 15: Verify consumer rewards balances are transferred to community pool after offboarding
	s.Run("ConsumerRewardsTransferred_v28", func() {
		if len(s.rewardsPoolPreUpgrade) == 0 {
			s.T().Skip("consumer rewards pool was empty before upgrade; skipping transfer check")
		}

		// The community pool must contain every denom that was held in the rewards pool
		// before the upgrade handler ran, with at least the same integer amount.
		// uatom is skipped because it accrues continuously from staking rewards.
		//
		// Use the REST API rather than ExecQuery: the CLI returns coins as concatenated
		// "amount+denom" strings, while the API returns structured {denom, amount} objects.
		apiURL := s.Chain.GetHostAPIAddress() + "/cosmos/distribution/v1beta1/community_pool"
		resp, err := http.Get(apiURL) //nolint:gosec
		s.Require().NoError(err)
		defer resp.Body.Close()
		cpBody, err := io.ReadAll(resp.Body)
		s.Require().NoError(err)

		cpAmounts := make(map[string]sdkmath.Int)
		gjson.GetBytes(cpBody, "pool").ForEach(func(_, entry gjson.Result) bool {
			denom := entry.Get("denom").String()
			amt, err := chainsuite.StrToSDKInt(entry.Get("amount").String())
			if err == nil {
				cpAmounts[denom] = amt
			}
			return true
		})
		for denom, preUpgradeAmtStr := range s.rewardsPoolPreUpgrade {
			if denom == "uatom" {
				continue
			}
			preUpgradeAmt, err := chainsuite.StrToSDKInt(preUpgradeAmtStr)
			s.Require().NoError(err)
			cpAmt, ok := cpAmounts[denom]
			s.Require().True(ok,
				"community pool should contain denom %s after consumer rewards transfer", denom)
			s.Require().True(cpAmt.GTE(preUpgradeAmt),
				"community pool amount for %s (%s) should be >= pre-upgrade rewards pool amount (%s)",
				denom, cpAmt, preUpgradeAmt)
			s.T().Logf("Verified transfer of %s %s from consumer rewards pool to community pool", preUpgradeAmt, denom)
		}
	})

	// Phase 16: Verify proposals query remains operational
	s.Run("ProposalsQuery_v28", func() {
		out, _, err := s.Chain.Validators[0].ExecQuery(ctx, "gov", "proposals")
		s.Require().NoError(err)
		s.Require().Contains(string(out), s.Env.UpgradeName,
			"proposals query should return the upgrade proposal after v28 upgrade")
	})
}

// requireConsumerListed asserts that at least one consumer chain is listed.
func (s *ICSOffboardSuite) requireConsumerListed(ctx context.Context) {
	out, _, err := s.Chain.Validators[0].ExecQuery(ctx, "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().Contains(string(out), "chain_id", "expected at least one consumer chain to be listed")
}

// requireConsumerNotListed asserts that no consumer chains are listed.
func (s *ICSOffboardSuite) requireConsumerNotListed(ctx context.Context) {
	out, _, err := s.Chain.Validators[0].ExecQuery(ctx, "provider", "list-consumer-chains")
	if err != nil {
		// Command doesn't exist — module removed, so no consumers
		return
	}
	s.Require().NotContains(string(out), s.consumer.Config().ChainID,
		"consumer chain should not be listed after offboarding")
}

// requireCreateConsumerBlocked tries to create a consumer and expects it to fail.
func (s *ICSOffboardSuite) requireCreateConsumerBlocked(ctx context.Context) {
	chainID := "blocked-1"
	spawnTime := time.Now().Add(1 * time.Minute)
	params := providertypes.MsgCreateConsumer{
		ChainId: chainID,
		Metadata: providertypes.ConsumerMetadata{
			Name:        chainID,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
		InitializationParameters: &providertypes.ConsumerInitializationParameters{
			InitialHeight:                     clienttypes.Height{RevisionNumber: 1, RevisionHeight: 1},
			SpawnTime:                         spawnTime,
			BlocksPerDistributionTransmission: 1,
			CcvTimeoutPeriod:                  2419200000000000,
			TransferTimeoutPeriod:             3600000000000,
			ConsumerRedistributionFraction:    "0.75",
			HistoricalEntries:                 10000,
			UnbondingPeriod:                   1728000000000000,
			GenesisHash:                       []byte("Z2VuX2hhc2g="),
			BinaryHash:                        []byte("YmluX2hhc2g="),
		},
	}
	paramsBz, err := json.Marshal(params)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(ctx, paramsBz, "consumer-addition.json")
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName,
		"provider", "create-consumer",
		path.Join(s.Chain.GetNode().HomeDir(), "consumer-addition.json"))
	s.Require().Error(err, "MsgCreateConsumer should be blocked on v27")
}

// restartRelayer stops and starts the relayer.
func (s *ICSOffboardSuite) restartRelayer(ctx context.Context) {
	rep := chainsuite.GetRelayerExecReporter(ctx)
	s.Require().NoError(s.Relayer.StopRelayer(ctx, rep))
	s.Require().NoError(s.Relayer.StartRelayer(ctx, rep))
}

func allProviderKeysCopied() []bool {
	return []bool{true, true, true, true, true, true}
}
