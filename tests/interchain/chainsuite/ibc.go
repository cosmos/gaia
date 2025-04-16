package chainsuite

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func SendSimpleIBCTx(ctx context.Context, chainA *Chain, chainB *Chain, relayer *Relayer) error {
	addr1 := chainA.ValidatorWallets[0].Address
	addr2 := chainB.ValidatorWallets[0].Address

	senderTxChannel, err := relayer.GetTransferChannel(ctx, chainA, chainB)
	if err != nil {
		return err
	}

	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", senderTxChannel.Counterparty.ChannelID, chainA.Config().Denom))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	initial1, err := chainA.GetBalance(ctx, addr1, chainA.Config().Denom)
	if err != nil {
		return err
	}
	initial2, err := chainB.GetBalance(ctx, addr2, dstIbcDenom)
	if err != nil {
		return err
	}

	amountToSend := math.NewInt(1_000_000)
	_, err = chainA.Validators[0].SendIBCTransfer(ctx, senderTxChannel.ChannelID, chainA.ValidatorWallets[0].Moniker, ibc.WalletAmount{
		Denom:   chainA.Config().Denom,
		Amount:  amountToSend,
		Address: addr2,
	}, ibc.TransferOptions{})
	if err != nil {
		return err
	}
	tCtx, tCancel := context.WithTimeout(ctx, 30*CommitTimeout)
	defer tCancel()

	for tCtx.Err() == nil {
		time.Sleep(CommitTimeout)
		var final1, final2 math.Int
		final1, err = chainA.GetBalance(ctx, addr1, chainA.Config().Denom)
		if err != nil {
			continue
		}
		final2, err = chainB.GetBalance(ctx, addr2, dstIbcDenom)
		if err != nil {
			continue
		}

		if !initial2.Add(amountToSend).Equal(final2) {
			err = fmt.Errorf("destination balance not updated; expected %s, got %s", initial2.Add(amountToSend), final2)
			continue
		}
		if !final1.LTE(initial1.Sub(amountToSend)) {
			err = fmt.Errorf("source balance not updated; expected <= %s, got %s", initial1.Sub(amountToSend), final1)
			continue
		}
		break
	}

	return err
}
