package gaia_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	gaiahelpers "github.com/cosmos/gaia/v29/app/helpers"
)

const legacyParamsTypeURL = "/cosmos.params.v1beta1.ParameterChangeProposal"

type historicalGovProposal struct {
	ID    uint64 `json:"id"`
	Value []byte `json:"value"`
}

func historicalGovProposals(t *testing.T) []historicalGovProposal {
	t.Helper()
	bz, err := os.ReadFile("testdata/legacy_gov_proposals.json")
	require.NoError(t, err)
	var fixture struct {
		Proposals []historicalGovProposal `json:"proposals"`
	}
	require.NoError(t, json.Unmarshal(bz, &fixture))
	return fixture.Proposals
}

// Use persisted bytes and the real app registry: constructing an Any with a
// cached concrete value or registering types in this test would hide missing
// registration in NewGaiaApp. The fixtures also survive replacing SDK types
// with local compatibility types when upgrading past SDK v0.53.
func TestGaiaAppLegacyParamsProposals(t *testing.T) {
	app := gaiahelpers.Setup(t)
	t.Cleanup(func() { require.NoError(t, app.Close()) })
	ctx := app.NewContext(false)
	fixtures := historicalGovProposals(t)
	store := ctx.KVStore(app.GetKey(govtypes.StoreKey))
	for _, fixture := range fixtures {
		key := binary.BigEndian.AppendUint64([]byte{0}, fixture.ID)
		store.Set(key, fixture.Value)
	}
	require.NoError(t, app.GovKeeper.ProposalID.Set(ctx, fixtures[len(fixtures)-1].ID+1))
	app.SimWriteState()
	_, err := app.Commit()
	require.NoError(t, err)
	ctx = app.NewUncachedContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	govv1.RegisterQueryServer(queryHelper, govkeeper.NewQueryServer(app.GovKeeper))
	govv1beta1.RegisterQueryServer(queryHelper, govkeeper.NewLegacyQueryServer(app.GovKeeper))
	client := govv1.NewQueryClient(queryHelper)
	legacyClient := govv1beta1.NewQueryClient(queryHelper)
	queryCtx := context.Background()

	t.Run("single proposal and tally", func(t *testing.T) {
		for _, fixture := range fixtures {
			t.Run(fmt.Sprint(fixture.ID), func(t *testing.T) {
				response, err := client.Proposal(queryCtx, &govv1.QueryProposalRequest{ProposalId: fixture.ID})
				require.NoError(t, err)
				require.Equal(t, fixture.ID, response.Proposal.Id)
				bz, err := app.AppCodec().Marshal(response.Proposal)
				require.NoError(t, err)
				require.Equal(t, fixture.Value, bz, "binary round-trip must preserve historical fields")
				jsonBytes, err := app.AppCodec().MarshalJSON(response)
				require.NoError(t, err)
				var decoded govv1.QueryProposalResponse
				require.NoError(t, app.AppCodec().UnmarshalJSON(jsonBytes, &decoded))
				bz, err = app.AppCodec().Marshal(decoded.Proposal)
				require.NoError(t, err)
				require.Equal(t, fixture.Value, bz, "JSON round-trip must preserve historical fields")

				tally, err := client.TallyResult(queryCtx, &govv1.QueryTallyResultRequest{ProposalId: fixture.ID})
				require.NoError(t, err)
				require.Equal(t, response.Proposal.FinalTallyResult, tally.Tally)

				legacy, err := legacyClient.Proposal(queryCtx, &govv1beta1.QueryProposalRequest{ProposalId: fixture.ID})
				require.NoError(t, err)
				var content govv1.MsgExecLegacyContent
				require.NoError(t, app.AppCodec().Unmarshal(response.Proposal.Messages[0].Value, &content))
				require.Equal(t, content.Content.TypeUrl, legacy.Proposal.Content.TypeUrl)
				require.Equal(t, content.Content.Value, legacy.Proposal.Content.Value)
			})
		}
	})

	t.Run("list with total", func(t *testing.T) {
		response, err := client.Proposals(queryCtx, &govv1.QueryProposalsRequest{
			Pagination: &query.PageRequest{Limit: 1000, CountTotal: true},
		})
		require.NoError(t, err)
		require.Len(t, response.Proposals, len(fixtures))
		require.Equal(t, uint64(len(fixtures)), response.Pagination.Total)
		_, err = app.AppCodec().MarshalJSON(response)
		require.NoError(t, err)
		response, err = client.Proposals(queryCtx, &govv1.QueryProposalsRequest{
			Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
		})
		require.NoError(t, err)
		require.Len(t, response.Proposals, 1)
		require.Equal(t, uint64(len(fixtures)), response.Pagination.Total)
		legacy, err := legacyClient.Proposals(queryCtx, &govv1beta1.QueryProposalsRequest{
			Pagination: &query.PageRequest{Limit: 1000, CountTotal: true},
		})
		require.NoError(t, err)
		require.Len(t, legacy.Proposals, len(fixtures))
		_, err = app.AppCodec().MarshalJSON(legacy)
		require.NoError(t, err)
	})

	t.Run("status filter excludes historical proposals", func(t *testing.T) {
		response, err := client.Proposals(queryCtx, &govv1.QueryProposalsRequest{ProposalStatus: govv1.StatusVotingPeriod})
		require.NoError(t, err, "historical proposals must decode before the status filter is applied")
		require.Empty(t, response.Proposals)
		_, err = legacyClient.Proposals(queryCtx, &govv1beta1.QueryProposalsRequest{ProposalStatus: govv1beta1.StatusVotingPeriod})
		require.NoError(t, err)
	})

	t.Run("export", func(t *testing.T) {
		exported, err := app.ExportAppStateAndValidators(false, nil, nil)
		require.NoError(t, err)
		var modules map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(exported.AppState, &modules))
		var govGenesis govv1.GenesisState
		require.NoError(t, app.AppCodec().UnmarshalJSON(modules[govtypes.ModuleName], &govGenesis))
		require.Len(t, govGenesis.Proposals, len(fixtures))
		for i, fixture := range fixtures {
			bz, err := app.AppCodec().Marshal(govGenesis.Proposals[i])
			require.NoError(t, err)
			require.Equal(t, fixture.Value, bz)
		}
	})

	t.Run("params execution remains disabled", func(t *testing.T) {
		require.False(t, app.GovKeeper.LegacyRouter().HasRoute("params"))
		response, err := client.Proposal(queryCtx, &govv1.QueryProposalRequest{ProposalId: 30})
		require.NoError(t, err)
		var content govv1.MsgExecLegacyContent
		require.NoError(t, app.AppCodec().Unmarshal(response.Proposal.Messages[0].Value, &content))
		require.Equal(t, legacyParamsTypeURL, content.Content.TypeUrl)
		_, err = govkeeper.NewMsgServerImpl(app.GovKeeper).ExecLegacyContent(ctx, &content)
		require.ErrorIs(t, err, govtypes.ErrNoProposalHandlerExists)
	})
}

// Registering a historical Content also allows incoming transactions to decode.
// Exercise FinalizeBlock to retain the intentional coordinated-upgrade behavior:
// ante charges fees/increments sequence, while gov still rejects execution.
func TestGaiaAppLegacyParamsTransactions(t *testing.T) {
	for _, version := range []string{"v1", "v1beta1"} {
		t.Run(version, func(t *testing.T) {
			app := gaiahelpers.Setup(t)
			t.Cleanup(func() { require.NoError(t, app.Close()) })
			ctx := app.NewContext(false)
			key := secp256k1.GenPrivKeyFromSecret([]byte("legacy params regression test"))
			address := sdk.AccAddress(key.PubKey().Address())
			account := app.AccountKeeper.NewAccountWithAddress(ctx, address)
			app.AccountKeeper.SetAccount(ctx, account)
			feeParams, err := app.FeeMarketKeeper.GetParams(ctx)
			require.NoError(t, err)
			govParams, err := app.GovKeeper.Params.Get(ctx)
			require.NoError(t, err)
			funds := sdk.NewCoins(sdk.NewInt64Coin(feeParams.FeeDenom, 1_000_000_000_000)).Add(govParams.MinDeposit...)
			require.NoError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, funds))
			require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, address, funds))
			app.SimWriteState()
			_, err = app.Commit()
			require.NoError(t, err)

			var proposal govv1.Proposal
			require.NoError(t, app.AppCodec().Unmarshal(historicalGovProposals(t)[1].Value, &proposal))
			var content govv1.MsgExecLegacyContent
			require.NoError(t, app.AppCodec().Unmarshal(proposal.Messages[0].Value, &content))
			require.Equal(t, legacyParamsTypeURL, content.Content.TypeUrl)
			var msg sdk.Msg = &govv1.MsgSubmitProposal{
				Messages: proposal.Messages, Proposer: address.String(), InitialDeposit: govParams.MinDeposit,
				Title: proposal.Title, Summary: proposal.Summary,
			}
			if version == "v1beta1" {
				msg = &govv1beta1.MsgSubmitProposal{
					Content: content.Content, Proposer: address.String(), InitialDeposit: govParams.MinDeposit,
				}
			}
			transaction, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), app.GetTxConfig(),
				[]sdk.Msg{msg}, sdk.NewCoins(sdk.NewInt64Coin(feeParams.FeeDenom, 10_000_000)), 500_000,
				"", []uint64{account.GetAccountNumber()}, []uint64{0}, key)
			require.NoError(t, err)
			txBytes, err := app.GetTxConfig().TxEncoder()(transaction)
			require.NoError(t, err)
			_, err = app.GetTxConfig().TxDecoder()(txBytes)
			require.NoError(t, err, "historical content remains decodable in transaction bodies")
			result, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: 2, Time: time.Unix(2, 0), Txs: [][]byte{txBytes},
			})
			require.NoError(t, err)
			require.Len(t, result.TxResults, 1)
			require.Equal(t, govtypes.ModuleName, result.TxResults[0].Codespace) // "gov"
			require.Equal(t, govtypes.ErrNoProposalHandlerExists.ABCICode(), result.TxResults[0].Code)
			require.Positive(t, result.TxResults[0].GasUsed)
			_, err = app.Commit()
			require.NoError(t, err)
			ctx = app.NewUncachedContext(false, tmproto.Header{})
			require.Equal(t, uint64(1), app.AccountKeeper.GetAccount(ctx, address).GetSequence())
			require.True(t, app.BankKeeper.GetBalance(ctx, address, feeParams.FeeDenom).Amount.LT(funds.AmountOf(feeParams.FeeDenom)))
		})
	}
}
