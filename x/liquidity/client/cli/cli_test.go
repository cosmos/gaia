//go:build norace
// +build norace

package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"

	lapp "github.com/cosmos/gaia/v9/app"
	"github.com/cosmos/gaia/v9/x/liquidity"
	"github.com/cosmos/gaia/v9/x/liquidity/client/cli"
	liquiditytestutil "github.com/cosmos/gaia/v9/x/liquidity/client/testutil"
	liquiditytypes "github.com/cosmos/gaia/v9/x/liquidity/types"

	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	db *tmdb.MemDB // in-memory database is needed for exporting genesis cli integration test
}

// SetupTest creates a new network for _each_ integration test. We create a new
// network for each test because there are some state modifications that are
// needed to be made in order to make useful queries. However, we don't want
// these state changes to be present in other tests.
func (s *IntegrationTestSuite) SetupTest() {
	s.T().Log("setting up integration test suite")

	db := tmdb.NewMemDB()

	cfg := liquiditytestutil.NewConfig(db)
	cfg.NumValidators = 1

	var liquidityGenState liquiditytypes.GenesisState
	err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[liquiditytypes.ModuleName], &liquidityGenState)
	s.Require().NoError(err)

	liquidityGenState.Params = liquiditytypes.DefaultParams()

	cfg.GenesisState[liquiditytypes.ModuleName] = cfg.Codec.MustMarshalJSON(&liquidityGenState)
	cfg.AccountTokens = sdk.NewInt(100_000_000_000) // node0token denom
	cfg.StakingTokens = sdk.NewInt(100_000_000_000) // stake denom

	genesisStateGov := govtypes.DefaultGenesisState()
	genesisStateGov.DepositParams = govtypes.NewDepositParams(sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, govtypes.DefaultMinDepositTokens)), time.Duration(15)*time.Second)
	genesisStateGov.VotingParams = govtypes.NewVotingParams(time.Duration(3) * time.Second)
	genesisStateGov.TallyParams.Quorum = sdk.MustNewDecFromStr("0.01")
	bz, err := cfg.Codec.MarshalJSON(genesisStateGov)
	s.Require().NoError(err)
	cfg.GenesisState["gov"] = bz

	s.cfg = cfg
	s.network = network.New(s.T(), s.cfg)
	s.db = db

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

// TearDownTest cleans up the curret test network after each test in the suite.
func (s *IntegrationTestSuite) TearDownTest() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// TestIntegrationTestSuite every integration test suite.
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestNewCreatePoolCmd() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid pool type id",
			[]string{
				"invalidpooltypeid",
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"pool type id is not supported",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000)), sdk.NewCoin("denomZ", sdk.NewInt(100_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"invalid number of denoms",
			[]string{
				fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000)), sdk.NewCoin("denomZ", sdk.NewInt(100_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"deposit coin less than minimum deposit amount",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(1_000)), sdk.NewCoin(denomY, sdk.NewInt(10_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 9,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewCreatePoolCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewDepositWithinBatchCmd() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid pool id",
			[]string{
				"invalidpoolid",
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(1_000_000)), sdk.NewCoin(denomY, sdk.NewInt(1_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"invalid number of denoms",
			[]string{
				fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(1_000_000)), sdk.NewCoin(denomY, sdk.NewInt(1_000_000)), sdk.NewCoin("denomZ", sdk.NewInt(1_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000_000)), sdk.NewCoin(denomY, sdk.NewInt(10_000_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewDepositWithinBatchCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewWithdrawWithinBatchCmd() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid pool id",
			[]string{
				"invalidpoolid",
				sdk.NewCoins(sdk.NewCoin("poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D", sdk.NewInt(10_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"bad pool coin",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin("badpoolcoindenom", sdk.NewInt(10_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 29,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin("poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D", sdk.NewInt(10_000))).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewWithdrawWithinBatchCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestNewSwapWithinBatchCmd() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000))).String(),
				denomY,
				fmt.Sprintf("%.2f", 0.02),
				fmt.Sprintf("%.3f", 0.003),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"swap type id not supported",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", uint32(2)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000))).String(),
				denomY,
				fmt.Sprintf("%.2f", 0.02),
				fmt.Sprintf("%.2f", 0.03),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, nil, 0,
		},
		{
			"bad offer coin fee",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", uint32(1)),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000))).String(),
				denomY,
				fmt.Sprintf("%.2f", 0.02),
				fmt.Sprintf("%.2f", 0.01),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 35,
		},
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", liquiditytypes.DefaultSwapTypeID),
				sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000))).String(),
				denomY,
				fmt.Sprintf("%.2f", 1.0),
				fmt.Sprintf("%.3f", 0.003),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewSwapWithinBatchCmd()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"pool_types":[{"id":1,"name":"StandardLiquidityPool","min_reserve_coin_num":2,"max_reserve_coin_num":2,"description":"Standard liquidity pool with pool price function X/Y, ESPM constraint, and two kinds of reserve coins"}],"min_init_deposit_amount":"1000000","init_pool_coin_mint_amount":"1000000","max_reserve_coin_amount":"0","pool_creation_fee":[{"denom":"stake","amount":"40000000"}],"swap_fee_rate":"0.003000000000000000","withdraw_fee_rate":"0.000000000000000000","max_order_amount_ratio":"0.100000000000000000","unit_batch_height":1,"circuit_breaker_enabled":false}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`circuit_breaker_enabled: false
init_pool_coin_mint_amount: "1000000"
max_order_amount_ratio: "0.100000000000000000"
max_reserve_coin_amount: "0"
min_init_deposit_amount: "1000000"
pool_creation_fee:
- amount: "40000000"
  denom: stake
pool_types:
- description: Standard liquidity pool with pool price function X/Y, ESPM constraint,
    and two kinds of reserve coins
  id: 1
  max_reserve_coin_num: 2
  min_reserve_coin_num: 2
  name: StandardLiquidityPool
swap_fee_rate: "0.003000000000000000"
unit_batch_height: 1
withdraw_fee_rate: "0.000000000000000000"`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryLiquidityPool() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case with pool id",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
		{
			"with invalid pool coin denom",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolCoinDenom, "invalid_value"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with empty pool coin denom",
			[]string{
				fmt.Sprintf("--%s", cli.FlagPoolCoinDenom),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case with pool coin denom",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagPoolCoinDenom, "poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
		{
			"with invalid reserve acc",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagReserveAcc, "invalid_value"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with empty reserve acc",
			[]string{
				fmt.Sprintf("--%s", cli.FlagReserveAcc),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case with reserve acc",
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagReserveAcc, "cosmos1cva80e6jcxpezd3k5dl7zwy2eg30u7ld3y0a67"),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryLiquidityPool()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp liquiditytypes.QueryLiquidityPoolResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
				s.Require().NoError(err)
				s.Require().Equal(uint64(1), resp.GetPool().Id)
				s.Require().Equal(uint32(1), resp.GetPool().TypeId)
				s.Require().Len(resp.GetPool().ReserveCoinDenoms, 2)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryLiquidityPools() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"valid case",
			[]string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryLiquidityPools()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resps liquiditytypes.QueryLiquidityPoolsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resps)
				s.Require().NoError(err)

				for _, pool := range resps.GetPools() {
					s.Require().Equal(uint64(1), pool.Id)
					s.Require().Equal(uint32(1), pool.TypeId)
					s.Require().Len(pool.ReserveCoinDenoms, 2)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryLiquidityPoolBatch() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryLiquidityPoolBatch()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp liquiditytypes.QueryLiquidityPoolBatchResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
				s.Require().NoError(err)
				s.Require().Equal(uint64(1), resp.GetBatch().PoolId)
				s.Require().Equal(false, resp.GetBatch().Executed)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchDepositMsg() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// create new deposit
	_, err = liquiditytestutil.MsgDepositWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000_000)), sdk.NewCoin(denomY, sdk.NewInt(10_000_000))).String(),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchDepositMsg()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp liquiditytypes.QueryPoolBatchDepositMsgResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
				s.Require().NoError(err)
				s.Require().Equal(val.Address.String(), resp.GetDeposit().Msg.DepositorAddress)
				s.Require().Equal(true, resp.GetDeposit().Executed)
				s.Require().Equal(true, resp.GetDeposit().Succeeded)
				s.Require().Equal(true, resp.GetDeposit().ToBeDeleted)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchDepositMsgs() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// create new deposit
	_, err = liquiditytestutil.MsgDepositWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000_000)), sdk.NewCoin(denomY, sdk.NewInt(10_000_000))).String(),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchDepositMsgs()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resps liquiditytypes.QueryPoolBatchDepositMsgsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resps)
				s.Require().NoError(err)

				for _, deposit := range resps.GetDeposits() {
					s.Require().Equal(true, deposit.Executed)
					s.Require().Equal(true, deposit.Succeeded)
					s.Require().Equal(true, deposit.ToBeDeleted)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchWithdrawMsg() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// withdraw pool coin from the pool
	poolCoinDenom := "poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D"
	_, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin(poolCoinDenom, sdk.NewInt(10_000))).String(),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not suppoorted pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchWithdrawMsg()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp liquiditytypes.QueryPoolBatchWithdrawMsgResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
				s.Require().NoError(err)
				s.Require().Equal(val.Address.String(), resp.GetWithdraw().Msg.WithdrawerAddress)
				s.Require().Equal(poolCoinDenom, resp.GetWithdraw().Msg.PoolCoin.Denom)
				s.Require().Equal(true, resp.GetWithdraw().Executed)
				s.Require().Equal(true, resp.GetWithdraw().Succeeded)
				s.Require().Equal(true, resp.GetWithdraw().ToBeDeleted)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchWithdrawMsgs() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(100_000_000)), sdk.NewCoin(denomY, sdk.NewInt(100_000_000))).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// withdraw pool coin from the pool
	_, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin("poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D", sdk.NewInt(10_000))).String(),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchWithdrawMsgs()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resps liquiditytypes.QueryPoolBatchWithdrawMsgsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resps)
				s.Require().NoError(err)

				for _, withdraw := range resps.GetWithdraws() {
					s.Require().Equal(val.Address.String(), withdraw.Msg.WithdrawerAddress)
					s.Require().Equal(true, withdraw.Executed)
					s.Require().Equal(true, withdraw.Succeeded)
					s.Require().Equal(true, withdraw.ToBeDeleted)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchSwapMsg() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)
	X := sdk.NewCoin(denomX, sdk.NewInt(1_000_000_000))
	Y := sdk.NewCoin(denomY, sdk.NewInt(5_000_000_000))

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(X, Y).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// swap coins from the pool
	offerCoin := sdk.NewCoin(denomY, sdk.NewInt(50_000_000))
	_, err = liquiditytestutil.MsgSwapWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		fmt.Sprintf("%d", liquiditytypes.DefaultSwapTypeID),
		offerCoin.String(),
		denomX,
		fmt.Sprintf("%.3f", 0.019),
		fmt.Sprintf("%.3f", 0.003),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchSwapMsg()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp liquiditytypes.QueryPoolBatchSwapMsgResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
				s.Require().NoError(err)
				s.Require().Equal(val.Address.String(), resp.GetSwap().Msg.SwapRequesterAddress)
				s.Require().Equal(true, resp.GetSwap().Executed)
				s.Require().Equal(true, resp.GetSwap().Succeeded)
				s.Require().Equal(true, resp.GetSwap().ToBeDeleted)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCircuitBreaker() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)
	X := sdk.NewCoin(denomX, sdk.NewInt(1_000_000_000))
	Y := sdk.NewCoin(denomY, sdk.NewInt(5_000_000_000))

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(X, Y).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// swap coins from the pool
	offerCoin := sdk.NewCoin(denomY, sdk.NewInt(50_000_000))
	output, err := liquiditytestutil.MsgSwapWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		fmt.Sprintf("%d", liquiditytypes.DefaultSwapTypeID),
		offerCoin.String(),
		denomX,
		fmt.Sprintf("%.3f", 0.019),
		fmt.Sprintf("%.3f", 0.003),
	)
	var txRes sdk.TxResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(0))
	s.Require().NoError(err)

	circuitBreakerEnabled := true
	circuitBreakerEnabledStr, err := json.Marshal(&circuitBreakerEnabled)
	if err != nil {
		panic(err)
	}

	paramChange := paramscutils.ParamChangeProposalJSON{
		Title:       "enable-circuit-breaker",
		Description: "enable circuit breaker",
		Changes: []paramscutils.ParamChangeJSON{
			{
				Subspace: liquiditytypes.ModuleName,
				Key:      "CircuitBreakerEnabled",
				Value:    circuitBreakerEnabledStr,
			},
		},
		Deposit: sdk.NewCoin(s.cfg.BondDenom, govtypes.DefaultMinDepositTokens).String(),
	}
	paramChangeProp, err := json.Marshal(&paramChange)
	if err != nil {
		panic(err)
	}

	// create a param change proposal with deposit
	_, err = liquiditytestutil.MsgParamChangeProposalExec(
		val.ClientCtx,
		val.Address.String(),
		testutil.WriteToNewTempFile(s.T(), string(paramChangeProp)).Name(),
	)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	_, err = liquiditytestutil.MsgVote(val.ClientCtx, val.Address.String(), "1", "yes")
	s.Require().NoError(err)
	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// check if circuit breaker is enabled
	expectedOutput := `{"pool_types":[{"id":1,"name":"StandardLiquidityPool","min_reserve_coin_num":2,"max_reserve_coin_num":2,"description":"Standard liquidity pool with pool price function X/Y, ESPM constraint, and two kinds of reserve coins"}],"min_init_deposit_amount":"1000000","init_pool_coin_mint_amount":"1000000","max_reserve_coin_amount":"0","pool_creation_fee":[{"denom":"stake","amount":"40000000"}],"swap_fee_rate":"0.003000000000000000","withdraw_fee_rate":"0.000000000000000000","max_order_amount_ratio":"0.100000000000000000","unit_batch_height":1,"circuit_breaker_enabled":true}`
	out, err := clitestutil.ExecTestCLICmd(val.ClientCtx, cli.GetCmdQueryParams(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err)
	s.Require().Equal(expectedOutput, strings.TrimSpace(out.String()))

	// fail swap coins because of circuit breaker
	output, err = liquiditytestutil.MsgSwapWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		fmt.Sprintf("%d", liquiditytypes.DefaultSwapTypeID),
		offerCoin.String(),
		denomX,
		fmt.Sprintf("%.3f", 0.019),
		fmt.Sprintf("%.3f", 0.003),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(40))
	s.Require().Equal(txRes.RawLog, "failed to execute message; message index: 0: circuit breaker is triggered")

	// fail create new pool because of circuit breaker
	output, err = liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(X, Y).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(40))
	s.Require().Equal(txRes.RawLog, "failed to execute message; message index: 0: circuit breaker is triggered")

	// fail create new deposit because of circuit breaker
	_, err = liquiditytestutil.MsgDepositWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(sdk.NewCoin(denomX, sdk.NewInt(10_000_000)), sdk.NewCoin(denomY, sdk.NewInt(10_000_000))).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(40))
	s.Require().Equal(txRes.RawLog, "failed to execute message; message index: 0: circuit breaker is triggered")

	// success withdraw pool coin from the pool even though circuit breaker is true
	poolCoinDenom := "poolC33A77E752C183913636A37FE1388ACA22FE7BED792BEB2E72EF2DA857703D8D"
	output, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin(poolCoinDenom, sdk.NewInt(500000))).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(0))

	output, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin(poolCoinDenom, sdk.NewInt(499999))).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(0))

	// withdraw last pool coin
	output, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin(poolCoinDenom, sdk.NewInt(1))).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(0))

	// fail withdraw because of the pool is depleted
	output, err = liquiditytestutil.MsgWithdrawWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		sdk.NewCoins(sdk.NewCoin(poolCoinDenom, sdk.NewInt(1))).String(),
	)
	s.Require().NoError(err)
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(output.Bytes(), &txRes))
	s.Require().Equal(txRes.Code, uint32(39))
	s.Require().Equal(txRes.RawLog, "failed to execute message; message index: 0: the pool is depleted of reserve coin, reinitializing is required by deposit")
}

func (s *IntegrationTestSuite) TestGetCmdQueryPoolBatchSwapMsgs() {
	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)
	X := sdk.NewCoin(denomX, sdk.NewInt(1_000_000_000))
	Y := sdk.NewCoin(denomY, sdk.NewInt(5_000_000_000))

	// liquidity pool should be created prior to test this integration test
	_, err := liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(X, Y).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	// swap coins from the pool
	offerCoin := sdk.NewCoin(denomY, sdk.NewInt(50_000_000))
	_, err = liquiditytestutil.MsgSwapWithinBatchExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", uint32(1)),
		fmt.Sprintf("%d", liquiditytypes.DefaultSwapTypeID),
		offerCoin.String(),
		denomX,
		fmt.Sprintf("%.3f", 0.019),
		fmt.Sprintf("%.3f", 0.003),
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid pool id",
			[]string{
				"invalidpoolid",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"with not supported pool id",
			[]string{
				fmt.Sprintf("%d", uint32(2)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
		},
		{
			"valid case",
			[]string{
				fmt.Sprintf("%d", uint32(1)),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryPoolBatchSwapMsgs()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resps liquiditytypes.QueryPoolBatchSwapMsgsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &resps)
				s.Require().NoError(err)

				for _, swap := range resps.GetSwaps() {
					s.Require().Equal(val.Address.String(), swap.Msg.SwapRequesterAddress)
					s.Require().Equal(true, swap.Executed)
					s.Require().Equal(true, swap.Succeeded)
					s.Require().Equal(true, swap.ToBeDeleted)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestInitGenesis() {
	testCases := []struct {
		name      string
		flags     func(dir string) []string
		expectErr bool
		err       error
	}{
		{
			name: "default genesis state",
			flags: func(dir string) []string {
				return []string{
					"liquidity-test",
				}
			},
			expectErr: false,
			err:       nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			testMbm := module.NewBasicManager(liquidity.AppModuleBasic{})

			home := s.T().TempDir()
			logger := tmlog.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			s.Require().NoError(err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			interfaceRegistry := types.NewInterfaceRegistry()
			marshaler := codec.NewProtoCodec(interfaceRegistry)
			clientCtx := client.Context{}.
				WithCodec(marshaler).
				WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := genutilcli.InitCmd(testMbm, home)
			cmd.SetArgs(
				tc.flags(home),
			)

			if tc.expectErr {
				err := cmd.ExecuteContext(ctx)
				s.Require().EqualError(err, tc.err.Error())
			} else {
				s.Require().NoError(cmd.ExecuteContext(ctx))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestExportGenesis() {
	clientCtx := s.network.Validators[0].ClientCtx
	serverCtx := s.network.Validators[0].Ctx

	home := clientCtx.HomeDir

	// verify genesis file saved in temp directory
	genDocFile := clientCtx.HomeDir + "/config/genesis.json"
	genDoc, err := tmtypes.GenesisDocFromFile(genDocFile)
	s.Require().NoError(err)
	s.Require().NotNil(genDoc)

	val := s.network.Validators[0]

	// use two different tokens that are minted to the test account
	denomX, denomY := liquiditytypes.AlphabeticalDenomPair("node0token", s.network.Config.BondDenom)
	X := sdk.NewCoin(denomX, sdk.NewInt(1_000_000_000))
	Y := sdk.NewCoin(denomY, sdk.NewInt(5_000_000_000))

	// create liquidity pool
	_, err = liquiditytestutil.MsgCreatePoolExec(
		val.ClientCtx,
		val.Address.String(),
		fmt.Sprintf("%d", liquiditytypes.DefaultPoolTypeID),
		sdk.NewCoins(X, Y).String(),
	)
	s.Require().NoError(err)

	err = s.network.WaitForNextBlock()
	s.Require().NoError(err)

	cmd := server.ExportCmd(
		func(_ tmlog.Logger, _ tmdb.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
			appOpts servertypes.AppOptions,
		) (servertypes.ExportedApp, error) {
			encCfg := lapp.MakeEncodingConfig()
			encCfg.Marshaler = codec.NewProtoCodec(encCfg.InterfaceRegistry)

			// get logger and in-memory database
			logger := serverCtx.Logger
			db := s.db

			var app *lapp.LiquidityApp
			if height != -1 {
				app = lapp.NewLiquidityApp(logger, db, traceStore, false, map[int64]bool{}, "", uint(1), encCfg, appOpts)

				if err := app.LoadHeight(height); err != nil {
					return servertypes.ExportedApp{}, err
				}
			} else {
				app = lapp.NewLiquidityApp(logger, db, traceStore, true, map[int64]bool{}, "", uint(1), encCfg, appOpts)
			}

			return app.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
		},
		home,
	)

	args := []string{fmt.Sprintf("--%s=%s", flags.FlagHome, home)}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
	s.Require().NoError(err)

	var exportedGenDoc tmtypes.GenesisDoc
	err = tmjson.Unmarshal(out.Bytes(), &exportedGenDoc)
	s.Require().NoError(err)

	s.Require().Equal(clientCtx.ChainID, exportedGenDoc.ChainID)
}
