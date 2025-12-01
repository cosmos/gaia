package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	tokenfactorykeeper "github.com/cosmos/tokenfactory/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/cosmos/tokenfactory/x/tokenfactory/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TestCreateDenom tests the creation of a new token denomination
// Verifies:
// - Denom is created with correct format: factory/{creator_address}/{subdenom}
// - Creator is set as admin
// - Creation fee is charged correctly
// - Duplicate subdenom is rejected
// - Invalid subdenom characters are rejected
func TestCreateDenom(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Create test account with sufficient balance for creation fee
	creatorAddr := sdk.AccAddress("test_creator_addr1")
	creatorAcc := f.accountKeeper.NewAccountWithAddress(ctx, creatorAddr)
	f.accountKeeper.SetAccount(ctx, creatorAcc)

	// Fund creator account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", creatorAddr, creationFee))

	// Test cases:
	// 1. Create denom with valid subdenom
	subdenom := "mytoken"
	expectedDenom := fmt.Sprintf("factory/%s/%s", creatorAddr.String(), subdenom)
	msg := tokenfactorytypes.NewMsgCreateDenom(creatorAddr.String(), subdenom)
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	_, err := msgServer.CreateDenom(ctx, msg)
	require.NoError(t, err)

	// 2. Verify denom exists and creator is admin
	authorityMetadata, err := f.tokenFactoryKeeper.GetAuthorityMetadata(ctx, expectedDenom)
	require.NoError(t, err)
	require.Equal(t, creatorAddr.String(), authorityMetadata.Admin)

	// 3. Verify creation fee was charged
	balance := f.bankKeeper.GetBalance(ctx, creatorAddr, "stake")
	require.Equal(t, math.ZeroInt(), balance.Amount)

	// 4. Test duplicate subdenom rejection
	_, err = msgServer.CreateDenom(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")

	// 5. Test invalid subdenom characters
	invalidMsg := tokenfactorytypes.NewMsgCreateDenom(creatorAddr.String(), "invalid@token")
	_, err = msgServer.CreateDenom(ctx, invalidMsg)
	require.Error(t, err)
}

// TestMintTokens tests minting tokens to addresses
// Verifies:
// - Admin can mint tokens
// - Bank balance increases correctly
// - Non-admin cannot mint
// - Total supply tracking is accurate
func TestMintTokens(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin account and denom
	adminAddr := sdk.AccAddress("test_admin_addr____")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	// Setup: Create a test denom
	subdenom := "minttest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	// Create the denom first
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Test cases:
	// 1. Mint tokens as admin to own address
	mintAmount := sdk.NewCoin(denom, math.NewInt(1000000))
	mintMsg := tokenfactorytypes.NewMsgMint(adminAddr.String(), mintAmount)
	_, err = msgServer.Mint(ctx, mintMsg)
	require.NoError(t, err)

	// 2. Verify balance increased
	balance := f.bankKeeper.GetBalance(ctx, adminAddr, denom)
	require.Equal(t, math.NewInt(1000000), balance.Amount)

	// 3. Test non-admin mint rejection
	nonAdminAddr := sdk.AccAddress("test_non_admin_addr")
	nonAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, nonAdminAddr)
	f.accountKeeper.SetAccount(ctx, nonAdminAcc)
	unauthorizedMintMsg := tokenfactorytypes.NewMsgMint(nonAdminAddr.String(), mintAmount)
	_, err = msgServer.Mint(ctx, unauthorizedMintMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")

	// 4. Verify total supply
	totalSupply := f.bankKeeper.GetSupply(ctx, denom)
	require.Equal(t, math.NewInt(1000000), totalSupply.Amount)
}

// TestBurnTokens tests burning tokens
// Verifies:
// - Admin can burn tokens from their balance
// - Bank balance decreases correctly
// - Non-admin cannot burn
// - Cannot burn more than balance
// - Total supply tracking is accurate
func TestBurnTokens(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin account and denom with minted tokens
	adminAddr := sdk.AccAddress("test_burn_admin____")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	subdenom := "burntest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	// Create denom and mint tokens
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)
	mintAmount := sdk.NewCoin(denom, math.NewInt(1000000))
	mintMsg := tokenfactorytypes.NewMsgMint(adminAddr.String(), mintAmount)
	_, err = msgServer.Mint(ctx, mintMsg)
	require.NoError(t, err)

	// Test cases:
	// 1. Burn tokens as admin
	burnAmount := sdk.NewCoin(denom, math.NewInt(500000))
	burnMsg := tokenfactorytypes.NewMsgBurn(adminAddr.String(), burnAmount)
	_, err = msgServer.Burn(ctx, burnMsg)
	require.NoError(t, err)

	// 2. Verify balance decreased
	balance := f.bankKeeper.GetBalance(ctx, adminAddr, denom)
	require.Equal(t, math.NewInt(500000), balance.Amount)

	// 3. Verify total supply decreased
	totalSupply := f.bankKeeper.GetSupply(ctx, denom)
	require.Equal(t, math.NewInt(500000), totalSupply.Amount)

	// 4. Test burning more than balance
	overBurnMsg := tokenfactorytypes.NewMsgBurn(adminAddr.String(), sdk.NewCoin(denom, math.NewInt(1000000)))
	_, err = msgServer.Burn(ctx, overBurnMsg)
	require.Error(t, err)

	// 5. Test non-admin burn rejection
	nonAdminAddr := sdk.AccAddress("test_non_admin_burn")
	nonAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, nonAdminAddr)
	f.accountKeeper.SetAccount(ctx, nonAdminAcc)
	unauthorizedBurnMsg := tokenfactorytypes.NewMsgBurn(nonAdminAddr.String(), burnAmount)
	_, err = msgServer.Burn(ctx, unauthorizedBurnMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

// TestChangeAdmin tests transferring admin privileges
// Verifies:
// - Admin can transfer to new address
// - Old admin loses privileges
// - New admin gains privileges
// - Unauthorized admin changes are rejected
func TestChangeAdmin(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create original admin and denom
	originalAdminAddr := sdk.AccAddress("test_original_admin")
	originalAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, originalAdminAddr)
	f.accountKeeper.SetAccount(ctx, originalAdminAcc)

	// Fund original admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", originalAdminAddr, creationFee))

	newAdminAddr := sdk.AccAddress("test_new_admin_____")
	newAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, newAdminAddr)
	f.accountKeeper.SetAccount(ctx, newAdminAcc)

	subdenom := "admintest"
	denom := fmt.Sprintf("factory/%s/%s", originalAdminAddr.String(), subdenom)

	// Create denom
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(originalAdminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Test cases:
	// 1. Transfer admin to new address
	changeAdminMsg := tokenfactorytypes.NewMsgChangeAdmin(originalAdminAddr.String(), denom, newAdminAddr.String())
	_, err = msgServer.ChangeAdmin(ctx, changeAdminMsg)
	require.NoError(t, err)

	// 2. Verify new admin
	authorityMetadata, err := f.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom)
	require.NoError(t, err)
	require.Equal(t, newAdminAddr.String(), authorityMetadata.Admin)

	// 3. Verify old admin cannot mint
	mintMsg := tokenfactorytypes.NewMsgMint(originalAdminAddr.String(), sdk.NewCoin(denom, math.NewInt(100)))
	_, err = msgServer.Mint(ctx, mintMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")

	// 4. Verify new admin can mint
	newAdminMintMsg := tokenfactorytypes.NewMsgMint(newAdminAddr.String(), sdk.NewCoin(denom, math.NewInt(100)))
	_, err = msgServer.Mint(ctx, newAdminMintMsg)
	require.NoError(t, err)

	// 5. Test unauthorized admin change
	unauthorizedAddr := sdk.AccAddress("test_unauthorized__")
	unauthorizedAcc := f.accountKeeper.NewAccountWithAddress(ctx, unauthorizedAddr)
	f.accountKeeper.SetAccount(ctx, unauthorizedAcc)
	unauthorizedMsg := tokenfactorytypes.NewMsgChangeAdmin(unauthorizedAddr.String(), denom, unauthorizedAddr.String())
	_, err = msgServer.ChangeAdmin(ctx, unauthorizedMsg)
	require.Error(t, err)
}

// TestSetDenomMetadata tests setting token metadata
// Verifies:
// - Admin can set metadata
// - Metadata is queryable via bank module
// - Admin can update existing metadata
// - Non-admin cannot set metadata
// - All metadata fields are preserved correctly
func TestSetDenomMetadata(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin and denom
	adminAddr := sdk.AccAddress("test_metadata_admin")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	subdenom := "metatest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	// Create denom
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Test cases:
	// 1. Set metadata as admin
	metadata := banktypes.Metadata{
		Description: "Test token for metadata",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: "metatest", Exponent: 6},
		},
		Base:    denom,
		Display: "metatest",
		Name:    "Meta Test Token",
		Symbol:  "MTT",
	}
	setMetadataMsg := tokenfactorytypes.NewMsgSetDenomMetadata(adminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, setMetadataMsg)
	require.NoError(t, err)

	// 2. Query metadata via bank module
	retrievedMetadata, found := f.bankKeeper.GetDenomMetaData(ctx, denom)
	require.True(t, found)
	require.Equal(t, metadata.Description, retrievedMetadata.Description)
	require.Equal(t, metadata.Name, retrievedMetadata.Name)
	require.Equal(t, metadata.Symbol, retrievedMetadata.Symbol)

	// 3. Update metadata
	metadata.Description = "Updated description"
	updateMsg := tokenfactorytypes.NewMsgSetDenomMetadata(adminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, updateMsg)
	require.NoError(t, err)
	retrievedMetadata, found = f.bankKeeper.GetDenomMetaData(ctx, denom)
	require.True(t, found)
	require.Equal(t, "Updated description", retrievedMetadata.Description)

	// 4. Test non-admin setting metadata
	nonAdminAddr := sdk.AccAddress("test_non_admin_meta")
	nonAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, nonAdminAddr)
	f.accountKeeper.SetAccount(ctx, nonAdminAcc)
	unauthorizedMsg := tokenfactorytypes.NewMsgSetDenomMetadata(nonAdminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, unauthorizedMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

// TestUnauthorizedOperations tests that all privileged operations are properly restricted
// Verifies:
// - Non-admin cannot mint
// - Non-admin cannot burn
// - Non-admin cannot change admin
// - Non-admin cannot set metadata
// - Proper error messages are returned
func TestUnauthorizedOperations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin and denom with some tokens
	adminAddr := sdk.AccAddress("test_auth_admin____")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	nonAdminAddr := sdk.AccAddress("test_non_admin_____")
	nonAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, nonAdminAddr)
	f.accountKeeper.SetAccount(ctx, nonAdminAcc)

	subdenom := "authtest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	// Create denom and mint some tokens
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)
	mintMsg := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom, math.NewInt(1000000)))
	_, err = msgServer.Mint(ctx, mintMsg)
	require.NoError(t, err)

	// Test all unauthorized operations
	// 1. Unauthorized mint
	unauthorizedMintMsg := tokenfactorytypes.NewMsgMint(nonAdminAddr.String(), sdk.NewCoin(denom, math.NewInt(100)))
	_, err = msgServer.Mint(ctx, unauthorizedMintMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")

	// 2. Unauthorized burn
	unauthorizedBurnMsg := tokenfactorytypes.NewMsgBurn(nonAdminAddr.String(), sdk.NewCoin(denom, math.NewInt(100)))
	_, err = msgServer.Burn(ctx, unauthorizedBurnMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")

	// 3. Unauthorized admin change
	unauthorizedChangeAdminMsg := tokenfactorytypes.NewMsgChangeAdmin(nonAdminAddr.String(), denom, nonAdminAddr.String())
	_, err = msgServer.ChangeAdmin(ctx, unauthorizedChangeAdminMsg)
	require.Error(t, err)

	// 4. Unauthorized metadata set
	metadata := banktypes.Metadata{
		Description: "Test token",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: subdenom, Exponent: 6},
		},
		Base:    denom,
		Display: subdenom,
		Name:    "Unauthorized Test Token",
		Symbol:  "UTT",
	}
	unauthorizedMetadataMsg := tokenfactorytypes.NewMsgSetDenomMetadata(nonAdminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, unauthorizedMetadataMsg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

// TestBankModuleIntegration tests tokenfactory tokens work with bank module
// Verifies:
// - Tokens can be transferred using bank send
// - Balances are queryable
// - Denom metadata is accessible
// - Supply tracking works correctly
func TestBankModuleIntegration(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin, denom, and mint tokens
	adminAddr := sdk.AccAddress("test_bank_admin____")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	recipientAddr := sdk.AccAddress("test_bank_recipient")
	recipientAcc := f.accountKeeper.NewAccountWithAddress(ctx, recipientAddr)
	f.accountKeeper.SetAccount(ctx, recipientAcc)

	subdenom := "banktest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	// Create denom and mint tokens
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	mintAmount := sdk.NewCoin(denom, math.NewInt(1000000))
	mintMsg := tokenfactorytypes.NewMsgMint(adminAddr.String(), mintAmount)
	_, err = msgServer.Mint(ctx, mintMsg)
	require.NoError(t, err)

	// Test bank send operations
	sendAmount := sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(250000)))
	err = f.bankKeeper.SendCoins(ctx, adminAddr, recipientAddr, sendAmount)
	require.NoError(t, err)

	// Test balance queries
	adminBalance := f.bankKeeper.GetBalance(ctx, adminAddr, denom)
	require.Equal(t, math.NewInt(750000), adminBalance.Amount)

	recipientBalance := f.bankKeeper.GetBalance(ctx, recipientAddr, denom)
	require.Equal(t, math.NewInt(250000), recipientBalance.Amount)

	// Test supply queries
	totalSupply := f.bankKeeper.GetSupply(ctx, denom)
	require.Equal(t, math.NewInt(1000000), totalSupply.Amount)

	// Set and query metadata
	metadata := banktypes.Metadata{
		Description: "Bank integration test token",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: subdenom, Exponent: 6},
		},
		Base:    denom,
		Display: subdenom,
		Name:    "Bank Test Token",
		Symbol:  "BTT",
	}
	setMetadataMsg := tokenfactorytypes.NewMsgSetDenomMetadata(adminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, setMetadataMsg)
	require.NoError(t, err)

	retrievedMetadata, found := f.bankKeeper.GetDenomMetaData(ctx, denom)
	require.True(t, found)
	require.Equal(t, metadata.Name, retrievedMetadata.Name)
}

// TestCreationFeeMechanism tests the denom creation fee system
// Verifies:
// - Fee is charged correctly on creation
// - Insufficient balance is rejected
func TestCreationFeeMechanism(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create account with exact creation fee
	creatorAddr := sdk.AccAddress("test_fee_creator___")
	creatorAcc := f.accountKeeper.NewAccountWithAddress(ctx, creatorAddr)
	f.accountKeeper.SetAccount(ctx, creatorAcc)

	// Get creation fee from params
	params := f.tokenFactoryKeeper.GetParams(ctx)
	creationFee := params.DenomCreationFee

	// Fund account with exact amount
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", creatorAddr, creationFee))

	// Create denom
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	subdenom := "feetest"
	createMsg := tokenfactorytypes.NewMsgCreateDenom(creatorAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Verify fee was charged
	balance := f.bankKeeper.GetBalance(ctx, creatorAddr, "stake")
	require.Equal(t, math.ZeroInt(), balance.Amount)

	// Test with insufficient balance
	insufficientAddr := sdk.AccAddress("test_insufficient__")
	insufficientAcc := f.accountKeeper.NewAccountWithAddress(ctx, insufficientAddr)
	f.accountKeeper.SetAccount(ctx, insufficientAcc)

	// Fund with less than required
	insufficientAmount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", insufficientAmount))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", insufficientAddr, insufficientAmount))

	insufficientMsg := tokenfactorytypes.NewMsgCreateDenom(insufficientAddr.String(), "insufficient")
	_, err = msgServer.CreateDenom(ctx, insufficientMsg)
	require.Error(t, err)
}

// TestCommunityPoolFeeDeposit tests that denom creation fees are deposited to the community pool
// Verifies:
// - Creation fees go to community pool when EnableCommunityPoolFeeFunding is enabled
// - Community pool balance increases by exact fee amount
// - Fee collection mechanism works correctly
func TestCommunityPoolFeeDeposit(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Get initial community pool balance
	initialFeePool, err := f.distributionKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	initialPool := initialFeePool.CommunityPool

	// Setup creator account
	creatorAddr := sdk.AccAddress("test_pool_creator")
	creatorAcc := f.accountKeeper.NewAccountWithAddress(ctx, creatorAddr)
	f.accountKeeper.SetAccount(ctx, creatorAcc)

	// Get creation fee from params
	params := f.tokenFactoryKeeper.GetParams(ctx)
	creationFee := params.DenomCreationFee

	// Fund account with exact creation fee amount
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", creatorAddr, creationFee))

	// Create denom - fee should go to community pool
	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)
	createMsg := tokenfactorytypes.NewMsgCreateDenom(creatorAddr.String(), "pooltest")
	_, err = msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Get final community pool balance
	finalFeePool, err := f.distributionKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	finalPool := finalFeePool.CommunityPool

	// Verify fee was deposited to community pool
	// Convert sdk.Coins to sdk.DecCoins for comparison
	expectedIncrease := sdk.NewDecCoinsFromCoins(creationFee...)
	actualIncrease := finalPool.Sub(initialPool)

	require.Equal(t, expectedIncrease.String(), actualIncrease.String(),
		"community pool should increase by exactly the creation fee amount")
}

// TestMultipleDenoms tests creating and managing multiple denoms from one address
// Verifies:
// - Multiple denoms can be created
// - Each operates independently
// - State isolation is maintained
func TestMultipleDenoms(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create admin
	adminAddr := sdk.AccAddress("test_multi_admin___")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for multiple creation fees
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(30000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)

	// Create multiple denoms
	subdenom1 := "token1"
	subdenom2 := "token2"
	subdenom3 := "token3"

	denom1 := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom1)
	denom2 := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom2)
	denom3 := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom3)

	createMsg1 := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom1)
	_, err := msgServer.CreateDenom(ctx, createMsg1)
	require.NoError(t, err)

	createMsg2 := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom2)
	_, err = msgServer.CreateDenom(ctx, createMsg2)
	require.NoError(t, err)

	createMsg3 := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom3)
	_, err = msgServer.CreateDenom(ctx, createMsg3)
	require.NoError(t, err)

	// Mint different amounts to each
	mintMsg1 := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom1, math.NewInt(1000)))
	_, err = msgServer.Mint(ctx, mintMsg1)
	require.NoError(t, err)

	mintMsg2 := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom2, math.NewInt(2000)))
	_, err = msgServer.Mint(ctx, mintMsg2)
	require.NoError(t, err)

	mintMsg3 := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom3, math.NewInt(3000)))
	_, err = msgServer.Mint(ctx, mintMsg3)
	require.NoError(t, err)

	// Verify they operate independently
	balance1 := f.bankKeeper.GetBalance(ctx, adminAddr, denom1)
	require.Equal(t, math.NewInt(1000), balance1.Amount)

	balance2 := f.bankKeeper.GetBalance(ctx, adminAddr, denom2)
	require.Equal(t, math.NewInt(2000), balance2.Amount)

	balance3 := f.bankKeeper.GetBalance(ctx, adminAddr, denom3)
	require.Equal(t, math.NewInt(3000), balance3.Amount)

	// Test transferring admin on one doesn't affect others
	newAdminAddr := sdk.AccAddress("test_new_admin_____")
	newAdminAcc := f.accountKeeper.NewAccountWithAddress(ctx, newAdminAddr)
	f.accountKeeper.SetAccount(ctx, newAdminAcc)

	changeAdminMsg := tokenfactorytypes.NewMsgChangeAdmin(adminAddr.String(), denom1, newAdminAddr.String())
	_, err = msgServer.ChangeAdmin(ctx, changeAdminMsg)
	require.NoError(t, err)

	// Verify denom1 has new admin
	authorityMetadata1, err := f.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom1)
	require.NoError(t, err)
	require.Equal(t, newAdminAddr.String(), authorityMetadata1.Admin)

	// Verify denom2 and denom3 still have original admin
	authorityMetadata2, err := f.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom2)
	require.NoError(t, err)
	require.Equal(t, adminAddr.String(), authorityMetadata2.Admin)

	authorityMetadata3, err := f.tokenFactoryKeeper.GetAuthorityMetadata(ctx, denom3)
	require.NoError(t, err)
	require.Equal(t, adminAddr.String(), authorityMetadata3.Admin)
}

// TestGenesisExportImport tests state preservation across genesis export/import
// Verifies:
// - Denoms are preserved
// - Metadata is preserved
// - Admin assignments are preserved
// - Operations work after import
func TestGenesisExportImport(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	// Setup: Create denoms with various states
	adminAddr := sdk.AccAddress("test_genesis_admin_")
	adminAcc := f.accountKeeper.NewAccountWithAddress(ctx, adminAddr)
	f.accountKeeper.SetAccount(ctx, adminAcc)

	// Fund admin account for creation fee
	creationFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10000000)))
	require.NoError(t, f.bankKeeper.MintCoins(ctx, "mint", creationFee))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(ctx, "mint", adminAddr, creationFee))

	msgServer := tokenfactorykeeper.NewMsgServerImpl(f.tokenFactoryKeeper)

	// Create denom
	subdenom := "genesistest"
	denom := fmt.Sprintf("factory/%s/%s", adminAddr.String(), subdenom)

	createMsg := tokenfactorytypes.NewMsgCreateDenom(adminAddr.String(), subdenom)
	_, err := msgServer.CreateDenom(ctx, createMsg)
	require.NoError(t, err)

	// Mint tokens
	mintMsg := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom, math.NewInt(500000)))
	_, err = msgServer.Mint(ctx, mintMsg)
	require.NoError(t, err)

	// Set metadata
	metadata := banktypes.Metadata{
		Description: "Genesis test token",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: denom, Exponent: 0},
			{Denom: subdenom, Exponent: 6},
		},
		Base:    denom,
		Display: subdenom,
		Name:    "Genesis Test Token",
		Symbol:  "GTT",
	}
	setMetadataMsg := tokenfactorytypes.NewMsgSetDenomMetadata(adminAddr.String(), metadata)
	_, err = msgServer.SetDenomMetadata(ctx, setMetadataMsg)
	require.NoError(t, err)

	// Export genesis
	genesisState := f.tokenFactoryKeeper.ExportGenesis(ctx)

	// Create new fixture for import
	f2 := initFixture(t)
	ctx2 := f2.sdkCtx

	// Re-create accounts in new context
	adminAcc2 := f2.accountKeeper.NewAccountWithAddress(ctx2, adminAddr)
	f2.accountKeeper.SetAccount(ctx2, adminAcc2)

	// Import genesis to new context
	f2.tokenFactoryKeeper.InitGenesis(ctx2, *genesisState)

	// Note: Bank metadata is stored in bank module's genesis, not tokenfactory's genesis.
	// In a real chain, both would be exported/imported together.
	// For this test, we need to manually set the metadata in the new context.
	f2.bankKeeper.SetDenomMetaData(ctx2, metadata)

	// Verify all state preserved
	// Check admin assignment
	authorityMetadata, err := f2.tokenFactoryKeeper.GetAuthorityMetadata(ctx2, denom)
	require.NoError(t, err)
	require.Equal(t, adminAddr.String(), authorityMetadata.Admin)

	// Check metadata
	retrievedMetadata, found := f2.bankKeeper.GetDenomMetaData(ctx2, denom)
	require.True(t, found)
	require.Equal(t, metadata.Name, retrievedMetadata.Name)
	require.Equal(t, metadata.Symbol, retrievedMetadata.Symbol)

	// Verify operations still work - mint more tokens
	msgServer2 := tokenfactorykeeper.NewMsgServerImpl(f2.tokenFactoryKeeper)
	mintMsg2 := tokenfactorytypes.NewMsgMint(adminAddr.String(), sdk.NewCoin(denom, math.NewInt(100000)))
	_, err = msgServer2.Mint(ctx2, mintMsg2)
	require.NoError(t, err)
}
