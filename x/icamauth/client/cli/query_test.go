package cli

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/stretchr/testify/require"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"google.golang.org/grpc/status"

	"github.com/cosmos/gaia/v8/x/icamauth/types"
)

// TODO add suite tests framework for test the CLI commands
func TestGetInterchainAccountCmd(t *testing.T) {
	t.Skipf("needs populate the suite first")

	clientCtx := client.Context{}
	for _, tc := range []struct {
		name         string
		connectionID string
		owner        string
		want         string
		err          error
	}{
		{
			name:         "should allow valid query",
			connectionID: "connection-0",
			owner:        "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			want:         "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			err:          nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{
				tc.connectionID,
				tc.owner,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}
			out, err := clitestutil.ExecTestCLICmd(
				clientCtx,
				getInterchainAccountCmd(),
				args,
			)
			if tc.err != nil {
				stat, ok := status.FromError(tc.err)
				require.True(t, ok)
				require.ErrorIs(t, stat.Err(), tc.err)
				return
			}
			require.NoError(t, err)

			var resp types.QueryInterchainAccountResponse
			require.NoError(t, types.ModuleCdc.UnmarshalJSON(out.Bytes(), &resp))
			require.NotNil(t, resp.InterchainAccountAddress)
			require.Equal(t, tc.want, resp.InterchainAccountAddress)
		})
	}
}
