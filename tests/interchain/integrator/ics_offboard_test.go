package integrator

import (
	"context"
	"encoding/json"
	"path"
	"testing"
	"time"

	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"
)

const v26ImageVersion = "v26.0.0"

type ICSOffboardSuite struct {
	*chainsuite.Suite
	consumer *chainsuite.Chain
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

	// Phase 4: Upgrade to v27
	s.Run("UpgradeToV27", func() {
		err := s.Chain.Upgrade(ctx, s.Env.OldGaiaImageVersion, s.Env.OldGaiaImageVersion)
		s.Require().NoError(err)
		s.restartRelayer(ctx)
	})

	// Phase 5: Consumer still listed on v27
	s.Run("ConsumerListed_v27", func() {
		s.requireConsumerListed(ctx)
	})

	// Phase 6: CCV still works on v27
	s.Run("VerifyCCV_v27", func() {
		err := s.Chain.CheckCCV(ctx, s.consumer, s.Relayer, 1_000_000, 0, 1)
		s.Require().NoError(err)
	})

	// Phase 7: Creating new consumer should be disabled on v27
	s.Run("CreateConsumerDisabled_v27", func() {
		s.requireCreateConsumerBlocked(ctx)
	})

	// Phase 8: IBC transfer still works on v27
	s.Run("IBCTransfer_v27", func() {
		err := chainsuite.SendSimpleIBCTx(ctx, s.Chain, s.consumer, s.Relayer)
		s.Require().NoError(err)
	})

	// Phase 9: Upgrade to v28
	s.Run("UpgradeToV28", func() {
		err := s.Chain.Upgrade(ctx, s.Env.UpgradeName, s.Env.NewGaiaImageVersion)
		s.Require().NoError(err)
		s.restartRelayer(ctx)
	})

	// Phase 10: Chain liveness after v28
	s.Run("Liveness_v28", func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		s.Require().NoError(testutil.WaitForBlocks(timeoutCtx, 5, s.Chain))
	})

	// Phase 11: Consumer offboarded — provider module removed
	s.Run("ConsumerOffboarded_v28", func() {
		_, _, err := s.Chain.Validators[0].ExecQuery(ctx, "provider", "list-consumer-chains")
		if err != nil {
			// Provider query command doesn't exist — module removed. This is expected.
			return
		}
		// If the query somehow succeeds, the consumer should not be listed
		s.requireConsumerNotListed(ctx)
	})

	// Phase 12: IBC transfers still work after offboarding
	s.Run("IBCTransfer_v28", func() {
		err := chainsuite.SendSimpleIBCTx(ctx, s.Chain, s.consumer, s.Relayer)
		s.Require().NoError(err)
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
