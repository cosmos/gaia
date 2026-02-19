package consumer_chain_test

import (
	"context"
	"encoding/json"
	"path"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v27/tests/interchain/chainsuite"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/suite"
)

type CreateDisabledSuite struct {
	*chainsuite.Suite
	DelegatorWallet  ibc.Wallet
	DelegatorWallet2 ibc.Wallet
}

func (s *CreateDisabledSuite) SetupSuite() {
	s.Suite.SetupSuite()
	wallet, err := s.Chain.BuildWallet(s.GetContext(), "delegator", "")
	s.Require().NoError(err)
	s.DelegatorWallet = wallet
	s.Require().NoError(s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DelegatorWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}))

	wallet, err = s.Chain.BuildWallet(s.GetContext(), "delegator2", "")
	s.Require().NoError(err)
	s.DelegatorWallet2 = wallet
	s.Require().NoError(s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.DelegatorWallet2.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   s.Chain.Config().Denom,
	}))
}

func (s *CreateDisabledSuite) TestCreateConsumer() {
	var chainID string = "blocked-1"
	spawnTime := time.Now().Add(1 * time.Minute)
	initParams := &providertypes.ConsumerInitializationParameters{
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
	}
	params := providertypes.MsgCreateConsumer{
		ChainId: chainID,
		Metadata: providertypes.ConsumerMetadata{
			Name:        chainID,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
		InitializationParameters: initParams,
	}
	paramsBz, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = s.Chain.GetNode().WriteFile(s.GetContext(), paramsBz, "consumer-addition.json")
	if err != nil {
		return
	}
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(), "provider", "create-consumer", path.Join(s.Chain.GetNode().HomeDir(), "consumer-addition.json"))
	s.Require().Error(err, "MsgCreateConsumer should be blocked on provider chain")
}

func (s *CreateDisabledSuite) TestCreateConsumerThroughAuthz() {
	// Grant authorization to execute MsgCreateConsumer
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"authz", "grant", s.DelegatorWallet2.FormattedAddress(), "generic",
		"--msg-type=/interchain_security.ccv.provider.v1.MsgCreateConsumer",
	)
	s.Require().NoError(err, "authz grant should succeed")

	var chainID string = "blocked-authz-1"
	spawnTime := time.Now().Add(1 * time.Minute)
	initParams := &providertypes.ConsumerInitializationParameters{
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
	}
	params := providertypes.MsgCreateConsumer{
		Submitter: s.DelegatorWallet.FormattedAddress(),
		ChainId:   chainID,
		Metadata: providertypes.ConsumerMetadata{
			Name:        chainID,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
		InitializationParameters: initParams,
	}
	paramsBz, err := json.Marshal(params)
	if err != nil {
		return
	}
	err = s.Chain.GetNode().WriteFile(s.GetContext(), paramsBz, "consumer-addition-authz.json")
	if err != nil {
		return
	}

	// Try to execute MsgCreateConsumer through authz - should be blocked by the ante handler
	err = s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "provider", "create-consumer", path.Join(s.Chain.GetNode().HomeDir(), "consumer-addition-authz.json"), "--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err, "MsgCreateConsumer should be blocked even through authz")
}

func TestConsumerCreationDisabled(t *testing.T) {
	s := &CreateDisabledSuite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}
	suite.Run(t, s)
}

func (s CreateDisabledSuite) authzGenExec(ctx context.Context, grantee ibc.Wallet, command ...string) error {
	txjson, err := s.Chain.GenerateTx(ctx, 0, command...)
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, []byte(txjson), "tx.json")
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(
		ctx,
		grantee.FormattedAddress(),
		"authz", "exec", path.Join(s.Chain.Validators[0].HomeDir(), "tx.json"),
	)
	return err
}
