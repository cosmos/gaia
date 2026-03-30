package validator_test

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"
)

const (
	// downtimeTestJailDuration is the downtime_jail_duration set in genesis for this test.
	// Kept short so the unjail window elapses quickly.
	downtimeTestJailDuration = 5 * time.Second
	// downtimeMinSigned is min_signed_per_window: 0.5 means jailing triggers after
	// missing more than 50% of the signed_blocks_window (i.e. >5 of 10 blocks).
	downtimeMinSigned = "0.500000000000000000"
)

// downtimeTestGenesis extends the default genesis with a short downtime jail duration
// and a higher min_signed_per_window so that jailing is triggered quickly.
func downtimeTestGenesis() []cosmos.GenesisKV {
	return append(
		chainsuite.DefaultGenesis(),
		cosmos.NewGenesisKV("app_state.slashing.params.downtime_jail_duration", downtimeTestJailDuration.String()),
		cosmos.NewGenesisKV("app_state.slashing.params.min_signed_per_window", downtimeMinSigned),
	)
}

type DowntimeSuite struct {
	*chainsuite.Suite
}

func (s *DowntimeSuite) TestDowntimeSlashing() {
	ctx := s.GetContext()

	// --- Query slashing params ---
	signedBlocksWindowJSON, err := s.Chain.QueryJSON(ctx, "params.signed_blocks_window", "slashing", "params")
	s.Require().NoError(err)
	signedBlocksWindow := int(signedBlocksWindowJSON.Int())

	slashFractionJSON, err := s.Chain.QueryJSON(ctx, "params.slash_fraction_downtime", "slashing", "params")
	s.Require().NoError(err)
	slashFraction, err := sdkmath.LegacyNewDecFromStr(slashFractionJSON.String())
	s.Require().NoError(err)

	// --- Identify the target validator (lowest bonded power) ---
	// With the default SixValidators topology the stakes are:
	//   idx 0: 30M, 1: 29M, 2: 20M, 3: 10M, 4: 7M, 5: 4M  (total 100M)
	// Validator 5 holds 4%, well under the 30% threshold.
	const targetIdx = 5

	var totalBonded sdkmath.Int = sdkmath.ZeroInt()
	for _, w := range s.Chain.ValidatorWallets {
		val, err := s.Chain.StakingQueryValidator(ctx, w.ValoperAddress)
		s.Require().NoError(err)
		if val.Status == stakingtypes.Bonded {
			totalBonded = totalBonded.Add(val.Tokens)
		}
	}

	targetWallet := s.Chain.ValidatorWallets[targetIdx]
	targetVal, err := s.Chain.StakingQueryValidator(ctx, targetWallet.ValoperAddress)
	s.Require().NoError(err)
	s.Require().Equal(stakingtypes.Bonded, targetVal.Status, "target validator must start bonded")

	threshold30Pct := totalBonded.MulRaw(30).Quo(sdkmath.NewInt(100))
	s.Require().True(
		targetVal.Tokens.LTE(threshold30Pct),
		"target validator tokens (%s) must be <=30%% of total bonded (%s), threshold: %s",
		targetVal.Tokens, totalBonded, threshold30Pct,
	)

	tokensBefore := targetVal.Tokens

	// --- Stop the lowest-power validator node ---
	s.Require().NoError(s.Chain.Validators[targetIdx].StopContainer(ctx))

	// Wait for enough blocks so the missed-block counter exceeds the jailing threshold.
	// With min_signed_per_window=0.5 and window=10, jailing triggers after >5 missed blocks.
	// Waiting window+2 blocks gives a comfortable margin.
	blockWaitCtx, blockWaitCancel := context.WithTimeout(
		ctx,
		time.Duration(signedBlocksWindow+5)*chainsuite.CommitTimeout*2,
	)
	defer blockWaitCancel()
	s.Require().NoError(testutil.WaitForBlocks(blockWaitCtx, signedBlocksWindow+2, s.Chain))

	// --- Poll until the validator is jailed ---
	jailPollCtx, jailPollCancel := context.WithTimeout(ctx, 10*chainsuite.CommitTimeout)
	defer jailPollCancel()
	for jailPollCtx.Err() == nil {
		jailed, err := s.Chain.IsValoperJailed(ctx, targetWallet.ValoperAddress)
		s.Require().NoError(err)
		if jailed {
			break
		}
		time.Sleep(chainsuite.CommitTimeout)
	}
	s.Require().NoError(jailPollCtx.Err(), "timed out waiting for validator to be jailed")

	// --- Verify validator status: BOND_STATUS_UNBONDING, jailed=true ---
	targetVal, err = s.Chain.StakingQueryValidator(ctx, targetWallet.ValoperAddress)
	s.Require().NoError(err)
	s.Require().Equal(stakingtypes.Unbonding, targetVal.Status,
		"jailed validator should have BOND_STATUS_UNBONDING")
	s.Require().True(targetVal.Jailed, "jailed validator should have jailed=true")

	// --- Verify delegations were slashed (exact calculation) ---
	// Expected: tokensAfter = tokensBefore - floor(tokensBefore * slash_fraction_downtime)
	tokensAfter := targetVal.Tokens
	expectedSlash := slashFraction.MulInt(tokensBefore).TruncateInt()
	expectedTokens := tokensBefore.Sub(expectedSlash)
	s.Require().Equal(expectedTokens, tokensAfter,
		"tokens after slash should be %s (slashed %s from %s)",
		expectedTokens, expectedSlash, tokensBefore,
	)

	// --- Wait out the downtime jail window before unjailing ---
	time.Sleep(downtimeTestJailDuration)

	// --- Restart the validator node ---
	s.Require().NoError(s.Chain.Validators[targetIdx].StartContainer(ctx))
	s.Require().NoError(testutil.WaitForBlocks(ctx, 2, s.Chain))

	// --- Submit unjail transaction ---
	_, err = s.Chain.Validators[targetIdx].ExecTx(
		ctx,
		targetWallet.Moniker,
		"slashing", "unjail",
	)
	s.Require().NoError(err)

	// --- Poll until the validator returns to BOND_STATUS_BONDED ---
	bondedCtx, bondedCancel := context.WithTimeout(ctx, 30*chainsuite.CommitTimeout)
	defer bondedCancel()
	for bondedCtx.Err() == nil {
		targetVal, err = s.Chain.StakingQueryValidator(ctx, targetWallet.ValoperAddress)
		s.Require().NoError(err)
		if targetVal.Status == stakingtypes.Bonded {
			break
		}
		time.Sleep(chainsuite.CommitTimeout)
	}
	s.Require().NoError(bondedCtx.Err(), "timed out waiting for validator to become bonded again")
	s.Require().False(targetVal.Jailed, "unjailed validator should not have jailed=true")

	// --- Verify the validator is signing blocks ---
	// A non-zero voting power in the CometBFT validator set confirms it is back in consensus.
	hexAddr, err := s.Chain.GetValidatorHex(ctx, targetIdx)
	s.Require().NoError(err)
	power, err := s.Chain.GetValidatorPower(ctx, hexAddr)
	s.Require().NoError(err)
	s.Require().NotZero(power, "unjailed validator should appear in the CometBFT validator set")
}

func TestDowntime(t *testing.T) {
	s := DowntimeSuite{chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		ChainSpec: &interchaintest.ChainSpec{
			NumValidators: &chainsuite.SixValidators,
			ChainConfig: ibc.ChainConfig{
				ModifyGenesis: cosmos.ModifyGenesis(downtimeTestGenesis()),
			},
		},
	})}
	suite.Run(t, &s)
}
