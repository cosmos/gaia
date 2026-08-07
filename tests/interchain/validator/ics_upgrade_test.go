package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	providerCap     = 5
	stakingCap      = 8
	haltHeightDelta = 60
)

// GroupTransition describes expected validator group changes across an upgrade.
type GroupTransition struct {
	FromGroup string // "A", "B", "N"
	ToGroup   string
	// Expected token delta (positive = delegation, negative = unbond). Zero means no change.
	TokenDelta int64
}

type ICSUpgradeSuite struct {
	*chainsuite.Suite
}

func icsUpgradeGenesis() []cosmos.GenesisKV {
	return append(chainsuite.DefaultGenesis(),
		cosmos.NewGenesisKV("app_state.staking.params.max_validators", stakingCap),
		cosmos.NewGenesisKV("app_state.provider.params.max_provider_consensus_validators", providerCap),
	)
}

// --- Suite helpers (lifted/adapted from valset_upgrade_test.go) ---

func (s *ICSUpgradeSuite) proposeUpgrade() int64 {
	ctx := s.GetContext()
	height, err := s.Chain.Height(ctx)
	s.Require().NoError(err)

	haltHeight := height + haltHeightDelta

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     chainsuite.GovDepositAmount,
		Title:       "Upgrade to " + s.Env.UpgradeName,
		Name:        s.Env.UpgradeName,
		Description: "Upgrade to " + s.Env.UpgradeName,
		Height:      haltHeight,
	}
	upgradeTx, err := s.Chain.UpgradeProposal(ctx, interchaintest.FaucetAccountKeyName, proposal)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(ctx, upgradeTx.ProposalID))

	return haltHeight
}

func (s *ICSUpgradeSuite) waitForHalt(haltHeight int64) {
	ctx := s.GetContext()
	height, err := s.Chain.Height(ctx)
	s.Require().NoError(err)

	timeoutCtx, cancel := context.WithTimeout(ctx, (time.Duration(haltHeight-height)+10)*chainsuite.CommitTimeout)
	defer cancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+3, s.Chain)
	if err == nil {
		s.Require().Fail("chain should not produce blocks after halt height")
	} else if timeoutCtx.Err() == nil {
		s.Require().Fail("chain should not produce blocks after halt height")
	}

	height, err = s.Chain.Height(ctx)
	s.Require().NoError(err)
	s.Require().LessOrEqual(height-haltHeight, int64(1), "chain isn't halted at expected height")
}

func (s *ICSUpgradeSuite) completeUpgrade() {
	s.Require().NoError(s.Chain.ReplaceImagesAndRestart(s.GetContext(), s.Env.NewGaiaImageVersion))
}

func (s *ICSUpgradeSuite) waitUntilHeight(targetHeight int64) {
	ctx := s.GetContext()
	height, err := s.Chain.Height(ctx)
	s.Require().NoError(err)
	if height >= targetHeight {
		return
	}
	s.Require().NoError(testutil.WaitForBlocks(ctx, int(targetHeight-height), s.Chain))
}

// asyncStakingTx broadcasts a staking tx without waiting for 2 blocks.
// Needed for halt-1 timing where ExecTx's 2-block wait would time out.
func (s *ICSUpgradeSuite) asyncStakingTx(valIdx int, command ...string) {
	ctx := s.GetContext()
	cmd := s.Chain.Validators[valIdx].TxCommand(
		s.Chain.ValidatorWallets[valIdx].Moniker,
		command...,
	)
	stdout, _, err := s.Chain.Validators[valIdx].Exec(ctx, cmd, nil)
	s.Require().NoError(err)
	tx := cosmos.CosmosTx{}
	s.Require().NoError(json.Unmarshal(stdout, &tx))
	s.Require().Equal(0, tx.Code, "async tx failed: %s", tx.RawLog)
}

func (s *ICSUpgradeSuite) snapshotState() *chainsuite.ChainSnapshot {
	ctx := s.GetContext()
	snap, err := s.Chain.SnapshotValidatorState(ctx, s.Chain.ValidatorWallets)
	s.Require().NoError(err)
	return snap
}

// assertTopology verifies the group sizes match expectations.
func (s *ICSUpgradeSuite) assertTopology(snap *chainsuite.ChainSnapshot, expectedA, expectedB, expectedN int) {
	var countA, countB, countN int
	for _, v := range snap.Validators {
		switch {
		case v.Status == stakingtypes.Bonded && v.InCometSet:
			countA++
		case v.Status == stakingtypes.Bonded && !v.InCometSet:
			countB++
		default:
			countN++
		}
	}
	s.Require().Equal(expectedA, countA, "expected %d validators in group A, got %d", expectedA, countA)
	s.Require().Equal(expectedB, countB, "expected %d validators in group B, got %d", expectedB, countB)
	s.Require().Equal(expectedN, countN, "expected %d validators in group N, got %d", expectedN, countN)
}

// classifyValidator returns "A", "B", or "N" for a validator snapshot.
func classifyValidator(v chainsuite.ValidatorSnapshot) string {
	switch {
	case v.Status == stakingtypes.Bonded && v.InCometSet:
		return "A"
	case v.Status == stakingtypes.Bonded && !v.InCometSet:
		return "B"
	default:
		return "N"
	}
}

// assertSnapshotTransitions checks the 9 test report items from the requirements document.
// Each item runs in its own s.Run subtest for granular pass/fail visibility.
// preUpgrade is an optional snapshot taken right before completeUpgrade (i.e. at halt height).
// If provided, Item10 uses it as the jail-status baseline so that pre-upgrade staking txs
// (e.g. full unbonds that trigger jailing) do not produce false failures.
func (s *ICSUpgradeSuite) assertSnapshotTransitions(
	before, after *chainsuite.ChainSnapshot,
	expectedChanges map[int]GroupTransition,
	preUpgrade ...*chainsuite.ChainSnapshot,
) {
	s.Run("Item1_MaxValidatorsCollapse", func() {
		s.Require().Equal(uint32(after.CometSetSize), after.MaxValidators,
			"max_validators (%d) should equal CometBFT set size (%d) after ICS removal",
			after.MaxValidators, after.CometSetSize)
	})

	s.Run("Item2_CometSetSize", func() {
		expectedCometSize := 0
		for _, v := range after.Validators {
			if v.InCometSet {
				expectedCometSize++
			}
		}
		s.Require().Equal(expectedCometSize, after.CometSetSize, "CometBFT set size mismatch")
	})

	s.Run("Item3_BondedTokenConservation", func() {
		var expectedBonded int64
		for i, vBefore := range before.Validators {
			tokens := vBefore.Tokens
			afterGroup := classifyValidator(vBefore)
			if change, ok := expectedChanges[i]; ok {
				tokens += change.TokenDelta
				afterGroup = change.ToGroup
			}
			if afterGroup == "A" || afterGroup == "B" {
				expectedBonded += tokens
			}
		}
		s.Require().Equal(expectedBonded, after.TotalBondedTokens,
			"total bonded tokens: expected %d, got %d", expectedBonded, after.TotalBondedTokens)
	})

	s.Run("Item4_StakingPoolConsistency", func() {
		var sumBondedTokens int64
		for _, v := range after.Validators {
			if v.Status == stakingtypes.Bonded {
				sumBondedTokens += v.Tokens
			}
		}
		s.Require().Equal(sumBondedTokens, after.StakingPoolBonded,
			"staking pool bonded (%d) != sum of bonded validator tokens (%d)",
			after.StakingPoolBonded, sumBondedTokens)
	})

	s.Run("Item5_NoBGroup", func() {
		for _, v := range after.Validators {
			if v.Status == stakingtypes.Bonded {
				s.Require().True(v.InCometSet,
					"validator %d (%s) is bonded but not in CometBFT set, the B group should not exist after ICS removal",
					v.Index, v.OperatorAddr)
			}
		}
	})

	s.Run("Item6_PerValidatorTokens", func() {
		for i, vBefore := range before.Validators {
			vAfter := after.Validators[i]
			if change, ok := expectedChanges[i]; ok {
				expectedTokens := vBefore.Tokens + change.TokenDelta
				s.Require().Equal(expectedTokens, vAfter.Tokens,
					"validator %d tokens: expected %d, got %d", i, expectedTokens, vAfter.Tokens)
			} else {
				s.Require().Equal(vBefore.Tokens, vAfter.Tokens,
					"validator %d tokens should not change: was %d, got %d", i, vBefore.Tokens, vAfter.Tokens)
			}
		}
	})

	s.Run("Item7_GroupTransitions", func() {
		for i, vBefore := range before.Validators {
			vAfter := after.Validators[i]
			beforeGroup := classifyValidator(vBefore)
			afterGroup := classifyValidator(vAfter)
			if change, ok := expectedChanges[i]; ok {
				s.Require().Equal(change.FromGroup, beforeGroup,
					"validator %d: expected from-group %s, got %s", i, change.FromGroup, beforeGroup)
				s.Require().Equal(change.ToGroup, afterGroup,
					"validator %d: expected to-group %s, got %s", i, change.ToGroup, afterGroup)
			} else {
				s.Require().Equal(beforeGroup, afterGroup,
					"validator %d: group changed unexpectedly from %s to %s", i, beforeGroup, afterGroup)
			}
		}
	})

	s.Run("Item8_CometMembership", func() {
		for i, vBefore := range before.Validators {
			vAfter := after.Validators[i]
			if change, ok := expectedChanges[i]; ok {
				fromInComet := change.FromGroup == "A"
				toInComet := change.ToGroup == "A"
				s.Require().Equal(fromInComet, vBefore.InCometSet,
					"validator %d: expected InCometSet=%v before, got %v", i, fromInComet, vBefore.InCometSet)
				s.Require().Equal(toInComet, vAfter.InCometSet,
					"validator %d: expected InCometSet=%v after, got %v", i, toInComet, vAfter.InCometSet)
			} else {
				s.Require().Equal(vBefore.InCometSet, vAfter.InCometSet,
					"validator %d: CometBFT membership changed unexpectedly", i)
			}
		}
	})

	s.Run("Item9_VotingPowerFraction", func() {
		var totalCometPower int64
		var totalCometTokens int64
		for _, v := range after.Validators {
			if v.InCometSet {
				totalCometPower += v.CometPower
				totalCometTokens += v.Tokens
			}
		}
		if totalCometPower > 0 && totalCometTokens > 0 {
			for _, v := range after.Validators {
				if !v.InCometSet {
					continue
				}
				expectedFraction := float64(v.Tokens) / float64(totalCometTokens)
				actualFraction := float64(v.CometPower) / float64(totalCometPower)
				s.Require().InDelta(expectedFraction, actualFraction, 0.01,
					"validator %d voting power fraction: expected %.4f, got %.4f",
					v.Index, expectedFraction, actualFraction)
			}
		}
	})

	s.Run("Item10_JailStatusPreserved", func() {
		// Jail status must be preserved by the upgrade handler itself.
		// Use the pre-upgrade snapshot (state at halt height, after any pre-upgrade
		// staking txs) as the baseline so that validator jailings caused by those
		// txs (e.g. full unbonds) do not produce false failures.
		jailRef := before
		if len(preUpgrade) > 0 && preUpgrade[0] != nil {
			jailRef = preUpgrade[0]
		}
		for i, vRef := range jailRef.Validators {
			vAfter := after.Validators[i]
			s.Require().Equal(vRef.Jailed, vAfter.Jailed,
				"validator %d (%s) jail status changed across upgrade: before=%v after=%v",
				i, vRef.OperatorAddr, vRef.Jailed, vAfter.Jailed)
		}
	})
}

// assertLiveness waits for a few blocks after upgrade to prove chain liveness.
func (s *ICSUpgradeSuite) assertLiveness() {
	ctx := s.GetContext()
	timeoutCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	s.Require().NoError(testutil.WaitForBlocks(timeoutCtx, 5, s.Chain))
}

// --- Setup validation ---

func (s *ICSUpgradeSuite) TestSetup_Topology() {
	snap := s.snapshotState()

	// Assert exactly providerCap (5) validators in CometBFT set
	s.Require().Equal(providerCap, snap.CometSetSize,
		"expected %d validators in CometBFT set, got %d", providerCap, snap.CometSetSize)

	// Assert exactly stakingCap (8) validators bonded
	bondedCount := 0
	for _, v := range snap.Validators {
		if v.Status == stakingtypes.Bonded {
			bondedCount++
		}
	}
	s.Require().Equal(stakingCap, bondedCount,
		"expected %d bonded validators, got %d", stakingCap, bondedCount)

	// Assert val[8] and val[9] are not bonded
	s.Require().NotEqual(stakingtypes.Bonded, snap.Validators[8].Status,
		"val[8] should not be bonded")
	s.Require().NotEqual(stakingtypes.Bonded, snap.Validators[9].Status,
		"val[9] should not be bonded")

	// Assert full topology: 5A / 3B / 2N
	s.assertTopology(snap, providerCap, stakingCap-providerCap, chainsuite.TenValidators-stakingCap)
}

// --- Dimension 3: Delegation scenarios ---

func (s *ICSUpgradeSuite) Test3_1_NoDelegations() {
	before := s.snapshotState()
	s.assertTopology(before, 5, 3, 2)

	haltHeight := s.proposeUpgrade()
	s.waitForHalt(haltHeight)
	s.completeUpgrade()

	// Wait for state to settle
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 10, s.Chain))
	after := s.snapshotState()

	// Upgrade collapses max_validators 8→5; all B validators become N
	expectedChanges := map[int]GroupTransition{
		0: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		1: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		2: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		3: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		4: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		5: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		6: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		7: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		8: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
		9: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
	}
	s.assertSnapshotTransitions(before, after, expectedChanges)
	s.assertTopology(after, 5, 0, 5)
	s.assertLiveness()
}

func (s *ICSUpgradeSuite) Test3_2_DelegationsNoGroupChange() {
	before := s.snapshotState()

	haltHeight := s.proposeUpgrade()

	// At halt-15: delegate small amounts that don't cause group transitions
	s.waitUntilHeight(haltHeight - 15)

	// Delegate 1M to val[0] (A→A, 30M→31M)
	delegate0 := fmt.Sprintf("%d%s", 1_000_000, s.Chain.Config().Denom)
	s.Require().NoError(s.Chain.Validators[0].StakingDelegate(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Moniker,
		s.Chain.ValidatorWallets[0].ValoperAddress,
		delegate0,
	))

	// Delegate 1M to val[5] (B→N after upgrade, 9M→10M)
	delegate5 := fmt.Sprintf("%d%s", 1_000_000, s.Chain.Config().Denom)
	s.Require().NoError(s.Chain.Validators[5].StakingDelegate(
		s.GetContext(),
		s.Chain.ValidatorWallets[5].Moniker,
		s.Chain.ValidatorWallets[5].ValoperAddress,
		delegate5,
	))

	s.waitForHalt(haltHeight)
	preUpgrade := s.snapshotState()
	s.completeUpgrade()

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 10, s.Chain))
	after := s.snapshotState()

	expectedChanges := map[int]GroupTransition{
		0: {FromGroup: "A", ToGroup: "A", TokenDelta: 1_000_000},
		1: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		2: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		3: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		4: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		5: {FromGroup: "B", ToGroup: "N", TokenDelta: 1_000_000},
		6: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		7: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		8: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
		9: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
	}
	s.assertSnapshotTransitions(before, after, expectedChanges, preUpgrade)
	s.assertTopology(after, 5, 0, 5)
	s.assertLiveness()
}

// Test3_3_NAB_ABN_ANN tests group transitions for three validators across two events:
// the staking txs landing at haltHeight-1 and the upgrade completing.
//
// The name encodes three snapshots of {val[8], val[4], val[7]}:
//
//	NAB – initial state:          val[8]=N(3M),  val[4]=A(12M), val[7]=B(5M)
//	ABN – after staking txs:      val[8]=A(18M), val[4]=B(12M), val[7]=N(0M)
//	ANN – after upgrade:          val[8]=A(18M), val[4]=N(12M), val[7]=N(0M)
//
// val[8] receives +15M and jumps into the top-5 (N→A).
// val[4] is displaced from the top-5 by val[8] (A→B), then the B group is
// eliminated by the upgrade (B→N).
// val[7] is fully unbonded before halt (B→N) and stays N after upgrade.
func (s *ICSUpgradeSuite) Test3_3_NAB_ABN_ANN() {
	before := s.snapshotState()

	haltHeight := s.proposeUpgrade()

	// At halt-2 (straddling): these txs land at halt-1, CometBFT change at halt+1
	s.waitUntilHeight(haltHeight - 2)

	// Delegate ~15M to val[8] (N, 3M → 18M) → enters top 8 (B) and top 5 (A)
	delegate8 := fmt.Sprintf("%d%s", 15_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(8,
		"staking", "delegate",
		s.Chain.ValidatorWallets[8].ValoperAddress,
		delegate8,
	)

	// Unbond val[7] (B, 5M) entirely → becomes N
	unbond7 := fmt.Sprintf("%d%s", 5_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(7,
		"staking", "unbond",
		s.Chain.ValidatorWallets[7].ValoperAddress,
		unbond7,
	)

	s.waitForHalt(haltHeight)
	preUpgrade := s.snapshotState()
	s.completeUpgrade()

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 10, s.Chain))
	after := s.snapshotState()

	// val[8]: N→A (3M + 15M = 18M, enters top 5)
	// val[4]: A→N (12M, bumped from top 5 by val[8]; B group eliminated by upgrade)
	// val[5]: B→N (9M, B group eliminated by upgrade)
	// val[6]: B→N (7M, B group eliminated by upgrade)
	// val[7]: B→N (5M fully unbonded, jailed before upgrade due to zero tokens)
	expectedChanges := map[int]GroupTransition{
		0: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		1: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		2: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		3: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		4: {FromGroup: "A", ToGroup: "N", TokenDelta: 0},
		5: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		6: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		7: {FromGroup: "B", ToGroup: "N", TokenDelta: -5_000_000},
		8: {FromGroup: "N", ToGroup: "A", TokenDelta: 15_000_000},
		9: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
	}
	s.assertSnapshotTransitions(before, after, expectedChanges, preUpgrade)
	s.assertTopology(after, 5, 0, 5)
	s.assertLiveness()
}

// Test3_4_ANN_BAA_NBN tests group transitions for three validators across two events:
// the staking txs landing at haltHeight-1 and the upgrade completing.
//
// The name encodes per-validator timelines of {val[4], val[5], val[8]}
// as [initial state][state after txs][state after upgrade]:
//
//	ANN – val[4]: A(12M) → unbonds 10M → N(2M, rank 9; drops below staking cap) → N
//	BAA – val[5]: B(9M)  → delegates 6M → A(15M, enters top 5)                  → A
//	NBN – val[8]: N(3M)  → delegates 3M → B(6M, enters top 8)                   → N (B group eliminated)
func (s *ICSUpgradeSuite) Test3_4_ANN_BAA_NBN() {
	before := s.snapshotState()

	haltHeight := s.proposeUpgrade()

	// At halt-2 (straddling)
	s.waitUntilHeight(haltHeight - 2)

	// Unbond ~10M from val[4] (A, 12M → 2M) → drops below staking cap → becomes N
	unbond4 := fmt.Sprintf("%d%s", 10_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(4,
		"staking", "unbond",
		s.Chain.ValidatorWallets[4].ValoperAddress,
		unbond4,
	)

	// Delegate ~6M to val[5] (B, 9M → 15M) → enters top 5 → becomes A
	delegate5 := fmt.Sprintf("%d%s", 6_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(5,
		"staking", "delegate",
		s.Chain.ValidatorWallets[5].ValoperAddress,
		delegate5,
	)

	// Delegate ~3M to val[8] (N, 3M → 6M) → enters top 8 as B, then B group eliminated by upgrade → N
	delegate8 := fmt.Sprintf("%d%s", 3_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(8,
		"staking", "delegate",
		s.Chain.ValidatorWallets[8].ValoperAddress,
		delegate8,
	)

	s.waitForHalt(haltHeight)
	preUpgrade := s.snapshotState()
	s.completeUpgrade()

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 10, s.Chain))
	after := s.snapshotState()

	expectedChanges := map[int]GroupTransition{
		0: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		1: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		2: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		3: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		4: {FromGroup: "A", ToGroup: "N", TokenDelta: -10_000_000},
		5: {FromGroup: "B", ToGroup: "A", TokenDelta: 6_000_000},
		6: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		7: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		8: {FromGroup: "N", ToGroup: "N", TokenDelta: 3_000_000},
		9: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
	}
	s.assertSnapshotTransitions(before, after, expectedChanges, preUpgrade)
	s.assertTopology(after, 5, 0, 5)
	s.assertLiveness()
}

// Test3_5_Combined exercises all four interesting transition patterns simultaneously.
// Each subject validator has a unique [initial state][state after txs][state after upgrade] timeline:
//
//	ABN – val[4]: A(12M) → unbonds 1M; val[5] overtakes it → B(11M, rank 6) → N (B group eliminated)
//	BAA – val[5]: B(9M)  → delegates 6M → A(15M, enters top 5)               → A
//	BNN – val[7]: B(5M)  → fully unbonds → N(0M)                              → N
//	NBN – val[8]: N(3M)  → delegates 4M → B(7M, enters top 8)                → N (B group eliminated)
func (s *ICSUpgradeSuite) Test3_5_Combined() {
	before := s.snapshotState()

	haltHeight := s.proposeUpgrade()

	// At halt-2 (straddling)
	s.waitUntilHeight(haltHeight - 2)

	// val[5] B→A: delegate 6M (9M → 15M, enters top 5)
	delegate5 := fmt.Sprintf("%d%s", 6_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(5,
		"staking", "delegate",
		s.Chain.ValidatorWallets[5].ValoperAddress,
		delegate5,
	)

	// val[4] A→N: unbond 1M (12M → 11M, bumped from top 5 by val[5]; B group eliminated)
	unbond4 := fmt.Sprintf("%d%s", 1_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(4,
		"staking", "unbond",
		s.Chain.ValidatorWallets[4].ValoperAddress,
		unbond4,
	)

	// val[8] N→B→N: delegate 4M (3M → 7M, enters top 8 as B) → N after upgrade (B group eliminated)
	delegate8 := fmt.Sprintf("%d%s", 4_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(8,
		"staking", "delegate",
		s.Chain.ValidatorWallets[8].ValoperAddress,
		delegate8,
	)

	// val[7] B→N: fully unbond (5M → 0)
	unbond7 := fmt.Sprintf("%d%s", 5_000_000, s.Chain.Config().Denom)
	s.asyncStakingTx(7,
		"staking", "unbond",
		s.Chain.ValidatorWallets[7].ValoperAddress,
		unbond7,
	)

	s.waitForHalt(haltHeight)
	preUpgrade := s.snapshotState()
	s.completeUpgrade()

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 10, s.Chain))
	after := s.snapshotState()

	expectedChanges := map[int]GroupTransition{
		0: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		1: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		2: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		3: {FromGroup: "A", ToGroup: "A", TokenDelta: 0},
		4: {FromGroup: "A", ToGroup: "N", TokenDelta: -1_000_000},
		5: {FromGroup: "B", ToGroup: "A", TokenDelta: 6_000_000},
		6: {FromGroup: "B", ToGroup: "N", TokenDelta: 0},
		7: {FromGroup: "B", ToGroup: "N", TokenDelta: -5_000_000},
		8: {FromGroup: "N", ToGroup: "N", TokenDelta: 4_000_000},
		9: {FromGroup: "N", ToGroup: "N", TokenDelta: 0},
	}
	s.assertSnapshotTransitions(before, after, expectedChanges, preUpgrade)
	s.assertLiveness()
}

// --- Test runner ---

func TestICSUpgrade(t *testing.T) {
	s := &ICSUpgradeSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			Scope: chainsuite.ChainScopeTest,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.TenValidators,
				ChainConfig: ibc.ChainConfig{
					ModifyGenesis:        cosmos.ModifyGenesis(icsUpgradeGenesis()),
					ModifyGenesisAmounts: chainsuite.TenValidatorGenesisAmounts(chainsuite.Uatom),
				},
			},
		}),
	}
	suite.Run(t, s)
}
