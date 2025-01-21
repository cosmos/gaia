package validator_test

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/sjson"
	"golang.org/x/sync/errgroup"
)

type FeemarketSuite struct {
	*chainsuite.Suite
}

func (s *FeemarketSuite) TestGasGoesUp() {
	const (
		txsPerBlock         = 600
		blocksToPack        = 5
		maxBlockUtilization = 1000000
	)

	s.setMaxBlockUtilization(maxBlockUtilization)

	s.setCommitTimeout(150 * time.Second)

	s.packBlocks(txsPerBlock, blocksToPack)
}

func (s *FeemarketSuite) setMaxBlockUtilization(utilization int) {
	params, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "feemarket", "params")
	s.Require().NoError(err)

	params, err = sjson.SetBytes(params, "max_block_utilization", fmt.Sprint(utilization))
	s.Require().NoError(err)

	govAuthority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	proposalJson := fmt.Sprintf(`{
		"@type": "/feemarket.feemarket.v1.MsgParams",
		"authority": "%s"
}`, govAuthority)
	proposalJson, err = sjson.SetRaw(proposalJson, "params", string(params))
	s.Require().NoError(err)

	txhash, err := s.Chain.GetNode().SubmitProposal(s.GetContext(), interchaintest.FaucetAccountKeyName,
		cosmos.TxProposalv1{
			Title:    "Set Block Params",
			Deposit:  chainsuite.GovDepositAmount,
			Messages: []json.RawMessage{json.RawMessage(proposalJson)},
			Summary:  "Set Block Params",
			Metadata: "ipfs://CID",
		})
	s.Require().NoError(err)

	propId, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propId))
	maxBlockResult, err := s.Chain.QueryJSON(s.GetContext(), "max_block_utilization", "feemarket", "params")
	s.Require().NoError(err)
	maxBlock := maxBlockResult.String()
	s.Require().Equal(fmt.Sprint(utilization), maxBlock)
}

func (s *FeemarketSuite) packBlocks(txsPerBlock, blocksToPack int) {
	script := `
#!/bin/sh

set -ue
set -o pipefail

TX_COUNT=$1
CHAIN_BINARY=$2
FROM=$3
TO=$4
DENOM=$5
GAS_PRICES=$6
CHAIN_ID=$7
VAL_HOME=$8
NODE=$9

i=0


cd $HOME

SEQUENCE=$($CHAIN_BINARY query auth account $FROM --chain-id $CHAIN_ID --node $NODE --home $VAL_HOME -o json | jq -r .account.value.sequence)
ACCOUNT=$($CHAIN_BINARY query auth account $FROM --chain-id $CHAIN_ID --node $NODE --home $VAL_HOME -o json | jq -r .account.value.account_number)

if [ $SEQUENCE == "null" ]; then
	$CHAIN_BINARY query auth account $FROM --chain-id $CHAIN_ID --node $NODE --home $VAL_HOME -o json >&2
	exit 1
fi

if [ $ACCOUNT == "null" ]; then
	ACCOUNT=0
fi

$CHAIN_BINARY tx bank send $FROM $TO 1$DENOM --keyring-backend test --generate-only --account-number $ACCOUNT --from $FROM --chain-id $CHAIN_ID --gas 500000 --gas-adjustment 2.0 --gas-prices $GAS_PRICES$DENOM --home $VAL_HOME --node $NODE -o json > tx.json

while [ $i -lt $TX_COUNT ]; do
	$CHAIN_BINARY tx sign tx.json --from $FROM --chain-id $CHAIN_ID --sequence $SEQUENCE --keyring-backend test --account-number $ACCOUNT --offline --home $VAL_HOME > tx.json.signed
	tx=$($CHAIN_BINARY tx broadcast tx.json.signed --node $NODE --chain-id $CHAIN_ID --home $VAL_HOME -o json)
	if [ $(echo $tx | jq -r .code) -ne 0 ]; then
		echo "$tx" >&2
		$CHAIN_BINARY query tx $(echo $tx | jq -r .txhash) --chain-id $CHAIN_ID --node $NODE --home $VAL_HOME >&2
		exit 1
	else
		echo $(echo $tx | jq -r .txhash)
	fi
	SEQUENCE=$((SEQUENCE+1))
	i=$((i+1))
done
`
	for _, val := range s.Chain.Validators {
		err := val.WriteFile(s.GetContext(), []byte(script), "pack.sh")
		s.Require().NoError(err)
	}
	wallets := s.Chain.ValidatorWallets

	gasResult, err := s.Chain.QueryJSON(s.GetContext(), "price.amount", "feemarket", "gas-price", s.Chain.Config().Denom)
	s.Require().NoError(err)
	gasStr := gasResult.String()
	gasBefore, err := strconv.ParseFloat(gasStr, 64)
	s.Require().NoError(err)
	gasNow := gasBefore

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 1, s.Chain))

	prevBlock, err := s.Chain.Height(s.GetContext())
	s.Require().NoError(err)

	for i := 0; i < blocksToPack; i++ {
		eg := errgroup.Group{}
		for v, val := range s.Chain.Validators {
			val := val
			v := v
			eg.Go(func() error {
				_, stderr, err := val.Exec(s.GetContext(), []string{
					"sh", path.Join(val.HomeDir(), "pack.sh"),
					strconv.Itoa(txsPerBlock / len(s.Chain.Validators)),
					s.Chain.Config().Bin,
					wallets[v].Address,
					wallets[(v+1)%len(s.Chain.Validators)].Address,
					s.Chain.Config().Denom,
					fmt.Sprint(gasNow),
					s.Chain.Config().ChainID,
					val.HomeDir(),
					fmt.Sprintf("tcp://%s:26657", val.HostName()),
				}, nil)

				if err != nil {
					return fmt.Errorf("validator %d, err %w, stderr: %s", v, err, stderr)
				} else if len(stderr) > 0 {
					return fmt.Errorf("stderr: %s", stderr)
				}
				return nil
			})
		}
		s.Require().NoError(eg.Wait())
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 1, s.Chain))
		time.Sleep(5 * time.Second) // ensure the feemarket has time to update
		currentBlock, err := s.Chain.Height(s.GetContext())
		s.Require().NoError(err)
		s.Require().Equal(prevBlock+1, currentBlock)
		prevBlock = currentBlock

		gasResult, err := s.Chain.QueryJSON(s.GetContext(), "price.amount", "feemarket", "gas-price", s.Chain.Config().Denom)
		s.Require().NoError(err)
		gasStr = gasResult.String()
		gasNow, err = strconv.ParseFloat(gasStr, 64)
		s.Require().NoError(err)
		s.Require().Greater(gasNow, gasBefore)
		gasBefore = gasNow
	}
}

func (s *FeemarketSuite) setCommitTimeout(timeout time.Duration) {
	configToml := make(testutil.Toml)
	consensusToml := make(testutil.Toml)
	consensusToml["timeout_commit"] = timeout.String()
	configToml["consensus"] = consensusToml
	s.Chain.ModifyConfig(s.GetContext(), s.T(), map[string]testutil.Toml{"config/config.toml": configToml})
}

func TestFeemarket(t *testing.T) {
	s := &FeemarketSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			UpgradeOnSetup: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
	}
	suite.Run(t, s)
}
