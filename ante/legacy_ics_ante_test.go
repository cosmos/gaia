package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	protov2 "google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v28/ante"
	"github.com/cosmos/gaia/v28/app/helpers"
	gaiaerrors "github.com/cosmos/gaia/v28/types/errors"
	"github.com/cosmos/gaia/v28/x/legacy/ics"
)

// mockTx is a minimal sdk.Tx that holds a fixed slice of messages.
type mockTx struct{ msgs []sdk.Msg }

func (m mockTx) GetMsgs() []sdk.Msg                    { return m.msgs }
func (m mockTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (m mockTx) ValidateBasic() error                  { return nil }

// noopNext is the terminal AnteHandler used in tests.
var noopNext sdk.AnteHandler = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	return ctx, nil
}

func TestRejectLegacyICSDecorator(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	decorator := ante.NewRejectLegacyICSDecorator()

	// Register the stubs so sdk.MsgTypeURL can resolve them.
	ics.RegisterInterfaces(gaiaApp.InterfaceRegistry())

	tests := []struct {
		name      string
		msgs      []sdk.Msg
		expectErr bool
		errCode   uint32
	}{
		{
			name:      "no ICS messages - passes",
			msgs:      []sdk.Msg{},
			expectErr: false,
		},
		{
			name:      "MsgUpdateConsumer - rejected",
			msgs:      []sdk.Msg{&ics.MsgUpdateConsumer{}},
			expectErr: true,
			errCode:   gaiaerrors.ErrDeprecatedMessage.ABCICode(),
		},
		{
			name:      "MsgAssignConsumerKey - rejected",
			msgs:      []sdk.Msg{&ics.MsgAssignConsumerKey{}},
			expectErr: true,
			errCode:   gaiaerrors.ErrDeprecatedMessage.ABCICode(),
		},
		{
			name:      "MsgOptIn - rejected",
			msgs:      []sdk.Msg{&ics.MsgOptIn{}},
			expectErr: true,
			errCode:   gaiaerrors.ErrDeprecatedMessage.ABCICode(),
		},
		{
			name:      "MsgOptOut - rejected",
			msgs:      []sdk.Msg{&ics.MsgOptOut{}},
			expectErr: true,
			errCode:   gaiaerrors.ErrDeprecatedMessage.ABCICode(),
		},
		{
			name:      "MsgSetConsumerCommissionRate - rejected",
			msgs:      []sdk.Msg{&ics.MsgSetConsumerCommissionRate{}},
			expectErr: true,
			errCode:   gaiaerrors.ErrDeprecatedMessage.ABCICode(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := mockTx{msgs: tc.msgs}
			_, err := decorator.AnteHandle(ctx, tx, false, noopNext)
			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.errCode, gaiaerrors.ErrDeprecatedMessage.ABCICode())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
