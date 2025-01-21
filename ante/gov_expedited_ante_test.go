package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/gaia/v23/ante"
	"github.com/cosmos/gaia/v23/app/helpers"
)

func TestGovExpeditedProposalsDecorator(t *testing.T) {
	gaiaApp := helpers.Setup(t)

	testCases := []struct {
		name      string
		ctx       sdk.Context
		msgs      []sdk.Msg
		expectErr bool
	}{
		// these cases should pass
		{
			name: "expedited - govv1.MsgSubmitProposal - MsgSoftwareUpgrade",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&upgradetypes.MsgSoftwareUpgrade{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Plan: upgradetypes.Plan{
						Name:   "upgrade plan-plan",
						Info:   "some text here",
						Height: 123456789,
					},
				}}, true),
			},
			expectErr: false,
		},
		{
			name: "expedited - govv1.MsgSubmitProposal - MsgCancelUpgrade",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&upgradetypes.MsgCancelUpgrade{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				}}, true),
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - TextProposal",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newLegacyTextProp(false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgCommunityPoolSpend",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&distrtypes.MsgCommunityPoolSpend{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Recipient: sdk.AccAddress{}.String(),
					Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgTransfer",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgSend{
					FromAddress: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					ToAddress:   "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Amount:      sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, false), // normal
			},
			expectErr: false,
		},
		{
			name: "normal - govv1.MsgSubmitProposal - MsgUpdateParams",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgUpdateParams{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				}}, false),
			},
			expectErr: false,
		},
		// legacy proposals - antehandler should not affect them
		// submitted using "gaiad tx gov submit-legacy-proposal"
		{
			name:      "normal - govv1beta.MsgSubmitProposal - LegacySoftwareUpgrade",
			ctx:       sdk.Context{},
			msgs:      []sdk.Msg{newGovV1BETA1LegacyUpgradeProp()},
			expectErr: false,
		},
		{
			name:      "normal - govv1beta.MsgSubmitProposal - LegacyCancelSoftwareUpgrade",
			ctx:       sdk.Context{},
			msgs:      []sdk.Msg{newGovV1BETA1LegacyCancelUpgradeProp()},
			expectErr: false,
		},
		// these cases should fail
		// these are normal proposals, not whitelisted for expedited voting
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - Empty",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - TextProposal",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newLegacyTextProp(true), // expedite
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgCommunityPoolSpend",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&distrtypes.MsgCommunityPoolSpend{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Recipient: sdk.AccAddress{}.String(),
					Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgTransfer",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgSend{
					FromAddress: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					ToAddress:   "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
					Amount:      sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				}}, true),
			},
			expectErr: true,
		},
		{
			name: "fail - expedited - govv1.MsgSubmitProposal - MsgUpdateParams",
			ctx:  sdk.Context{},
			msgs: []sdk.Msg{
				newGovProp([]sdk.Msg{&banktypes.MsgUpdateParams{
					Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				}}, true),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			txCfg := gaiaApp.GetTxConfig()
			decorator := ante.NewGovExpeditedProposalsDecorator(gaiaApp.AppCodec())

			txBuilder := txCfg.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msgs...))

			_, err := decorator.AnteHandle(tc.ctx, txBuilder.GetTx(), false,
				func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil })
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func newLegacyTextProp(expedite bool) *govv1.MsgSubmitProposal {
	testProposal := govv1beta1.NewTextProposal("Proposal", "Test as normal proposal")
	msgContent, err := govv1.NewLegacyContent(testProposal, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	if err != nil {
		return nil
	}
	return newGovProp([]sdk.Msg{msgContent}, expedite)
}

func newGovV1BETA1LegacyUpgradeProp() *govv1beta1.MsgSubmitProposal {
	legacyContent := upgradetypes.NewSoftwareUpgradeProposal("test legacy upgrade", "test legacy upgrade", upgradetypes.Plan{
		Name:   "upgrade plan-plan",
		Info:   "some text here",
		Height: 123456789,
	})

	msg, _ := govv1beta1.NewMsgSubmitProposal(legacyContent, sdk.NewCoins(), sdk.AccAddress{})
	return msg
}

func newGovV1BETA1LegacyCancelUpgradeProp() *govv1beta1.MsgSubmitProposal {
	legacyContent := upgradetypes.NewCancelSoftwareUpgradeProposal("test legacy upgrade", "test legacy upgrade")

	msg, _ := govv1beta1.NewMsgSubmitProposal(legacyContent, sdk.NewCoins(), sdk.AccAddress{})
	return msg
}

func newGovProp(msgs []sdk.Msg, expedite bool) *govv1.MsgSubmitProposal {
	msg, _ := govv1.NewMsgSubmitProposal(msgs, sdk.NewCoins(), sdk.AccAddress{}.String(), "", "expedite", "expedite", expedite)
	// fmt.Println("### msg ###", msg, "err", err)
	return msg
}
