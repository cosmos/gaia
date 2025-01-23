package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestTokenFactory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfgA := LocalChainConfig

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t, zaptest.Level(zapcore.DebugLevel)), []*interchaintest.ChainSpec{
		{
			Name:          "tokenfactory",
			Version:       "local",
			ChainName:     cfgA.ChainID,
			NumValidators: &vals,
			NumFullNodes:  &fullNodes,
			ChainConfig:   cfgA,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	tokenfactoryA := chains[0].(*cosmos.CosmosChain)

	// Relayer Factory
	client, network := interchaintest.DockerSetup(t)

	ic := interchaintest.NewInterchain().
		AddChain(tokenfactoryA)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))

	// Chains
	appChain := chains[0].(*cosmos.CosmosChain)

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", DefaultGenesisAmt, appChain, appChain)
	user := users[0]
	uaddr := user.FormattedAddress()

	user2 := users[1]
	uaddr2 := user2.FormattedAddress()

	node := appChain.GetNode()

	tfDenom, _, err := node.TokenFactoryCreateDenom(ctx, user, "ictestdenom", 2_500_00)
	t.Log("TF Denom: ", tfDenom)
	require.NoError(t, err)

	t.Log("Mint TF Denom to user")
	node.TokenFactoryMintDenom(ctx, user.FormattedAddress(), tfDenom, 100)
	if balance, err := appChain.GetBalance(ctx, uaddr, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 100 {
		t.Fatal("balance not 100")
	}

	t.Log("Mint TF Denom to another user")
	node.TokenFactoryMintDenomTo(ctx, user.FormattedAddress(), tfDenom, 70, user2.FormattedAddress())
	if balance, err := appChain.GetBalance(ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 70 {
		t.Fatal("balance not 70")
	}

	// This allows the uaddr here to mint tokens on behalf of the contract. Typically you only allow a contract here, but this is testing.
	coreInitMsg := fmt.Sprintf(`{"allowed_mint_addresses":["%s"],"existing_denoms":["%s"]}`, uaddr, tfDenom)
	codeId, err := node.StoreContract(ctx, user.KeyName(), "contracts/tokenfactory_core.wasm")
	require.NoError(t, err)

	contract, err := node.InstantiateContract(ctx, user.KeyName(), codeId, coreInitMsg, true)
	require.NoError(t, err)

	// change admin to the contract
	_, err = node.TokenFactoryChangeAdmin(ctx, user.KeyName(), tfDenom, contract)
	require.NoError(t, err)

	// ensure the admin is the contract
	admin, err := appChain.TokenFactoryQueryAdmin(ctx, tfDenom)
	t.Log("admin", admin)
	if admin.AuthorityMetadata.Admin != contract {
		t.Fatal("admin not coreTFContract. Did not properly transfer.")
	}

	// Mint on the contract for the user to ensure mint bindings work.
	mintAmt := 31
	mintMsg := fmt.Sprintf(`{"mint":{"address":"%s","denom":[{"denom":"%s","amount":"%d"}]}}`, uaddr2, tfDenom, mintAmt)
	if _, err := appChain.ExecuteContract(ctx, user.KeyName(), contract, mintMsg); err != nil {
		t.Fatal(err)
	}

	// ensure uaddr2 has 31+70 = 101
	if balance, err := appChain.GetBalance(ctx, uaddr2, tfDenom); err != nil {
		t.Fatal(err)
	} else if balance.Int64() != 101 {
		t.Fatal("balance not 101")
	}

	t.Cleanup(func() {
		_ = ic.Close()
	})
}
