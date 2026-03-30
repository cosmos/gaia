package validator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
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
	"github.com/tidwall/gjson"
)

const (
	newValMoniker  = "new-validator"
	selfDelegation = 5_000_000  // uatom
	newValFunds    = 10_000_000 // uatom (delegation + gas headroom)
)

type CreateValidatorSuite struct {
	*chainsuite.Suite
}

// TestCreateValidatorFromFullNode verifies that after the upgrade a new validator
// can be created using a self-delegation submitted from a node that starts as a
// non-validator (full node).
func (s *CreateValidatorSuite) TestCreateValidatorFromFullNode() {
	ctx := s.GetContext()
	fullNode := s.Chain.FullNodes[0]

	// --- Create a new key on the full node's keyring ---
	// ExecBin automatically prepends --home so the key lands in the same directory
	// that ExecTx and KeyBech32 will read from.
	_, _, err := fullNode.ExecBin(ctx,
		"keys", "add", newValMoniker,
		"--keyring-backend", "test",
	)
	s.Require().NoError(err)

	// --- Derive bech32 addresses ---
	accAddr, err := fullNode.KeyBech32(ctx, newValMoniker, "acc")
	s.Require().NoError(err)

	valoperAddr, err := fullNode.KeyBech32(ctx, newValMoniker, "val")
	s.Require().NoError(err)

	// --- Fund the new account from the faucet ---
	s.Require().NoError(s.Chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   s.Chain.Config().Denom,
		Amount:  sdkmath.NewInt(newValFunds),
		Address: accAddr,
	}))

	// --- Get the consensus public key from the full node ---
	// `comet show-validator` returns JSON: {"@type":"/cosmos.crypto.ed25519.PubKey","key":"..."}
	pubkeyBz, _, err := fullNode.ExecBin(ctx, "comet", "show-validator")
	s.Require().NoError(err)

	// --- Build the create-validator JSON message ---
	type createValidatorMsg struct {
		Pubkey              json.RawMessage `json:"pubkey"`
		Amount              string          `json:"amount"`
		Moniker             string          `json:"moniker"`
		CommissionRate      string          `json:"commission-rate"`
		CommissionMaxRate   string          `json:"commission-max-rate"`
		CommissionMaxChange string          `json:"commission-max-change-rate"`
		MinSelfDelegation   string          `json:"min-self-delegation"`
	}

	msg := createValidatorMsg{
		Pubkey:              json.RawMessage(strings.TrimSpace(string(pubkeyBz))),
		Amount:              fmt.Sprintf("%d%s", selfDelegation, s.Chain.Config().Denom),
		Moniker:             newValMoniker,
		CommissionRate:      "0.1",
		CommissionMaxRate:   "0.2",
		CommissionMaxChange: "0.01",
		MinSelfDelegation:   "1",
	}

	createValBz, err := json.Marshal(msg)
	s.Require().NoError(err)

	// Write the JSON to the node container so the CLI can read it
	const createValFile = "create-validator.json"
	s.Require().NoError(fullNode.WriteFile(ctx, createValBz, createValFile))

	// --- Submit the create-validator transaction ---
	// ExecTx broadcasts the tx and waits for 2 blocks for confirmation.
	_, err = fullNode.ExecTx(
		ctx,
		newValMoniker,
		"staking", "create-validator",
		path.Join(fullNode.HomeDir(), createValFile),
	)
	s.Require().NoError(err)

	s.Require().NoError(testutil.WaitForBlocks(ctx, 2, s.Chain))

	// --- Verify the validator was created with the correct parameters ---
	val, err := s.Chain.StakingQueryValidator(ctx, valoperAddr)
	s.Require().NoError(err)
	s.Require().Equal(newValMoniker, val.Description.Moniker,
		"new validator moniker should match")
	s.Require().False(val.Jailed, "new validator should not be jailed")
	s.Require().Equal(sdkmath.NewInt(selfDelegation), val.Tokens,
		"validator tokens should equal the self-delegation amount")
	s.Require().Equal(stakingtypes.Bonded, val.Status,
		"new validator should be in Bonded state")

	// --- Verify the validator is signing blocks (positive CometBFT voting power) ---
	// Parse the consensus address from the full node's priv_validator_key.json.
	privValKeyBz, err := fullNode.ReadFile(ctx, "config/priv_validator_key.json")
	s.Require().NoError(err)

	hexAddr := gjson.GetBytes(privValKeyBz, "address").String()
	s.Require().NotEmpty(hexAddr, "priv_validator_key.json must contain the validator address")

	// Poll until the new validator appears in the CometBFT set with positive power.
	powerCtx, powerCancel := context.WithTimeout(ctx, 30*chainsuite.CommitTimeout)
	defer powerCancel()

	var power int64
	for powerCtx.Err() == nil {
		power, err = s.Chain.GetValidatorPower(ctx, hexAddr)
		if err == nil && power > 0 {
			break
		}
		time.Sleep(chainsuite.CommitTimeout)
	}
	s.Require().NoError(powerCtx.Err(), "timed out waiting for new validator to appear in CometBFT set")
	s.Require().Positive(power, "new validator should have positive voting power in CometBFT set")
}

func TestCreateValidator(t *testing.T) {
	oneFullNode := 1
	s := CreateValidatorSuite{chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		ChainSpec: &interchaintest.ChainSpec{
			NumValidators: &chainsuite.SixValidators,
			NumFullNodes:  &oneFullNode,
			ChainConfig: ibc.ChainConfig{
				ModifyGenesis: cosmos.ModifyGenesis(chainsuite.DefaultGenesis()),
			},
		},
	})}
	suite.Run(t, &s)
}
