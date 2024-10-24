package interchain_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gaia/v21/tests/interchain/chainsuite"
	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ICAControllerSuite struct {
	*chainsuite.Suite
	Host *chainsuite.Chain
}

func (s *ICAControllerSuite) SetupSuite() {
	s.Suite.SetupSuite()
	host, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.Host = host
}

func (s *ICAControllerSuite) TestICAController() {
	const amountToSend = int64(3_300_000_000)
	wallets := s.Chain.ValidatorWallets
	valIdx := 0

	var icaAddress, srcAddress string
	var err error
	for ; valIdx < len(wallets); valIdx++ {
		srcAddress = wallets[valIdx].Address
		icaAddress, err = s.Chain.SetupICAAccount(s.GetContext(), s.Host, s.Relayer, srcAddress, valIdx, amountToSend)
		if err == nil {
			break
		} else if strings.Contains(err.Error(), "active channel already set for this owner") {
			chainsuite.GetLogger(s.GetContext()).Sugar().Warnf("error setting up ICA account: %s", err)
			valIdx++
			continue
		}
		// if we get here, fail the test. Unexpected error.
		s.Require().NoError(err)
	}
	if icaAddress == "" {
		// this'll happen if every validator has an ICA account already
		s.Require().Fail("unable to create ICA account")
	}

	srcChannel, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Host)
	s.Require().NoError(err)

	_, err = s.Chain.SendIBCTransfer(s.GetContext(), srcChannel.ChannelID, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: icaAddress,
		Amount:  sdkmath.NewInt(amountToSend),
		Denom:   s.Chain.Config().Denom,
	}, ibc.TransferOptions{})
	s.Require().NoError(err)

	wallets = s.Host.ValidatorWallets
	s.Require().NoError(err)
	dstAddress := wallets[0].Address

	var ibcStakeDenom string
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balances, err := s.Host.BankQueryAllBalances(s.GetContext(), icaAddress)
		s.Require().NoError(err)
		s.Require().NotEmpty(balances)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc") {
				ibcStakeDenom = c.Denom
				break
			}
		}
		assert.NotEmpty(c, ibcStakeDenom)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	recipientBalanceBefore, err := s.Host.GetBalance(s.GetContext(), dstAddress, ibcStakeDenom)
	s.Require().NoError(err)

	icaAmount := int64(amountToSend / 3)

	srcConnection := srcChannel.ConnectionHops[0]

	s.Require().NoError(s.sendICATx(s.GetContext(), valIdx, srcAddress, dstAddress, icaAddress, srcConnection, icaAmount, ibcStakeDenom))

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		recipientBalanceAfter, err := s.Host.GetBalance(s.GetContext(), dstAddress, ibcStakeDenom)
		assert.NoError(c, err)

		assert.Equal(c, recipientBalanceBefore.Add(sdkmath.NewInt(icaAmount)), recipientBalanceAfter)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

}

func TestICAController(t *testing.T) {
	s := &ICAControllerSuite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		CreateRelayer:  true,
	})}
	suite.Run(t, s)
}

func (s *ICAControllerSuite) sendICATx(ctx context.Context, valIdx int, srcAddress string, dstAddress string, icaAddress string, srcConnection string, amount int64, denom string) error {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	bankSendMsg := banktypes.NewMsgSend(
		sdk.MustAccAddressFromBech32(icaAddress),
		sdk.MustAccAddressFromBech32(dstAddress),
		sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(amount))),
	)
	data, err := icatypes.SerializeCosmosTx(cdc, []proto.Message{bankSendMsg}, icatypes.EncodingProtobuf)
	if err != nil {
		return err
	}

	msg, err := json.Marshal(icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	})
	if err != nil {
		return err
	}
	msgPath := "msg.json"
	if err := s.Chain.Validators[valIdx].WriteFile(ctx, msg, msgPath); err != nil {
		return err
	}
	msgPath = s.Chain.Validators[valIdx].HomeDir() + "/" + msgPath
	_, err = s.Chain.Validators[valIdx].ExecTx(ctx, srcAddress,
		"interchain-accounts", "controller", "send-tx",
		srcConnection, msgPath,
	)
	if err != nil {
		return err
	}
	return nil
}
