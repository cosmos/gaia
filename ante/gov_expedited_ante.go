package ante

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var expeditedPropsWhitelist = map[string]struct{}{
	"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade": {},
	"/cosmos.upgrade.v1beta1.MsgCancelUpgrade":   {},
	// legacy proposals can still be submitted using govv1.MsgSubmitProposal
	"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":       {},
	"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal": {},
}

// Check if the proposal is whitelisted for expedited voting.
type GovExpeditedProposalsDecorator struct {
	cdc codec.BinaryCodec
}

func NewGovExpeditedProposalsDecorator(cdc codec.BinaryCodec) GovExpeditedProposalsDecorator {
	return GovExpeditedProposalsDecorator{
		cdc: cdc,
	}
}

// AnteHandle checks if the proposal is whitelisted for expedited voting.
// Only proposals submitted using "gaiad tx gov submit-proposal" can be expedited.
// Legacy proposals submitted using "gaiad tx gov submit-legacy-proposal" cannot be marked as expedited.
func (g GovExpeditedProposalsDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		prop, ok := msg.(*govv1.MsgSubmitProposal)
		if !ok {
			continue
		}
		if prop.Expedited {
			if err := g.validateExpeditedGovProp(prop); err != nil {
				return ctx, err
			}
		}
	}
	return next(ctx, tx, simulate)
}

func (g GovExpeditedProposalsDecorator) isWhitelisted(msgType string) bool {
	_, ok := expeditedPropsWhitelist[msgType]
	return ok
}

func (g GovExpeditedProposalsDecorator) validateExpeditedGovProp(prop *govv1.MsgSubmitProposal) error {
	msgs := prop.GetMessages()
	if len(msgs) == 0 {
		return fmt.Errorf("unsupported expedited proposal")
	}
	for _, message := range msgs {
		// in case of legacy content submitted using govv1.MsgSubmitProposal
		if sdkMsg, isLegacy := message.GetCachedValue().(*govv1.MsgExecLegacyContent); isLegacy {
			if !g.isWhitelisted(sdkMsg.Content.TypeUrl) {
				return fmt.Errorf("unsupported expedited proposal type: %s", sdkMsg.Content.TypeUrl)
			}
			continue
		}
		if !g.isWhitelisted(message.TypeUrl) {
			return fmt.Errorf("unsupported expedited proposal type: %s", message.TypeUrl)
		}
	}
	return nil
}
