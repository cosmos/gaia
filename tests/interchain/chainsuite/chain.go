package chainsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// This moniker is hardcoded into interchaintest
const validatorMoniker = "validator"

type Chain struct {
	*cosmos.CosmosChain
	ValidatorWallets []ValidatorWallet
	RelayerWallet    ibc.Wallet
}

type ValidatorWallet struct {
	Moniker        string
	Address        string
	ValoperAddress string
	ValConsAddress string
}

func chainFromCosmosChain(cosmos *cosmos.CosmosChain, relayerWallet ibc.Wallet) (*Chain, error) {
	c := &Chain{CosmosChain: cosmos}
	wallets, err := getValidatorWallets(context.Background(), c)
	if err != nil {
		return nil, err
	}
	c.ValidatorWallets = wallets
	c.RelayerWallet = relayerWallet
	return c, nil
}

// CreateChain creates a single new chain with the given version and returns the chain object.
func CreateChain(ctx context.Context, testName interchaintest.TestName, spec *interchaintest.ChainSpec) (*Chain, error) {
	cf := interchaintest.NewBuiltinChainFactory(
		GetLogger(ctx),
		[]*interchaintest.ChainSpec{spec},
	)

	chains, err := cf.Chains(testName.Name())
	if err != nil {
		return nil, err
	}
	cosmosChain := chains[0].(*cosmos.CosmosChain)
	relayerWallet, err := cosmosChain.BuildRelayerWallet(ctx, "relayer-"+cosmosChain.Config().ChainID)
	if err != nil {
		return nil, err
	}

	ic := interchaintest.NewInterchain().AddChain(cosmosChain, ibc.WalletAmount{
		Address: relayerWallet.FormattedAddress(),
		Denom:   cosmosChain.Config().Denom,
		Amount:  sdkmath.NewInt(ValidatorFunds),
	})

	dockerClient, dockerNetwork := GetDockerContext(ctx)

	if err := ic.Build(ctx, GetRelayerExecReporter(ctx), interchaintest.InterchainBuildOptions{
		Client:    dockerClient,
		NetworkID: dockerNetwork,
		TestName:  testName.Name(),
	}); err != nil {
		return nil, err
	}

	chain, err := chainFromCosmosChain(cosmosChain, relayerWallet)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func (c *Chain) GenerateTx(ctx context.Context, valIdx int, command ...string) (string, error) {
	command = append([]string{"tx"}, command...)
	command = append(command, "--generate-only", "--keyring-backend", "test", "--chain-id", c.Config().ChainID)
	command = c.Validators[valIdx].NodeCommand(command...)
	stdout, _, err := c.Validators[valIdx].Exec(ctx, command, nil)
	if err != nil {
		return "", err
	}
	return string(stdout), nil
}

func (c *Chain) WaitForProposalStatus(ctx context.Context, proposalID string, status govv1.ProposalStatus) error {
	propID, err := strconv.ParseInt(proposalID, 10, 64)
	if err != nil {
		return err
	}
	chainHeight, err := c.Height(ctx)
	if err != nil {
		return err
	}
	maxHeight := chainHeight + UpgradeDelta
	_, err = cosmos.PollForProposalStatusV1(ctx, c.CosmosChain, chainHeight, maxHeight, uint64(propID), status)
	return err
}

func (c *Chain) PassProposal(ctx context.Context, proposalID string) error {
	propID, err := strconv.ParseInt(proposalID, 10, 64)
	if err != nil {
		return err
	}
	err = c.VoteOnProposalAllValidators(ctx, uint64(propID), cosmos.ProposalVoteYes)
	if err != nil {
		return err
	}
	return c.WaitForProposalStatus(ctx, proposalID, govv1.StatusPassed)
}

func (c *Chain) ReplaceImagesAndRestart(ctx context.Context, version string) error {
	// bring down nodes to prepare for upgrade
	err := c.StopAllNodes(ctx)
	if err != nil {
		return err
	}

	// upgrade version on all nodes
	c.UpgradeVersion(ctx, c.GetNode().DockerClient, c.GetNode().Image.Repository, version)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and block production resumes.
	err = c.StartAllNodes(ctx)
	if err != nil {
		return err
	}

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()
	err = testutil.WaitForBlocks(timeoutCtx, 5, c)
	if err != nil {
		return fmt.Errorf("failed to wait for blocks after upgrade: %w", err)
	}

	// Flush "successfully migrated key info" messages
	for _, val := range c.Validators {
		_, _, err := val.ExecBin(ctx, "keys", "list", "--keyring-backend", "test")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) Upgrade(ctx context.Context, upgradeName, version string) error {
	height, err := c.Height(ctx)
	if err != nil {
		return err
	}

	haltHeight := height + UpgradeDelta

	proposal := cosmos.SoftwareUpgradeProposal{
		Deposit:     GovDepositAmount, // greater than min deposit
		Title:       "Upgrade to " + upgradeName,
		Name:        upgradeName,
		Description: "Upgrade to " + upgradeName,
		Height:      haltHeight,
	}
	upgradeTx, err := c.UpgradeProposal(ctx, interchaintest.FaucetAccountKeyName, proposal)
	if err != nil {
		return err
	}
	if err := c.PassProposal(ctx, upgradeTx.ProposalID); err != nil {
		return err
	}

	height, err = c.Height(ctx)
	if err != nil {
		return err
	}

	// wait for the chain to halt. We're asking for blocks after the halt height, so we should time out.
	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, (time.Duration(haltHeight-height)+10)*CommitTimeout)
	defer timeoutCtxCancel()
	err = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+3, c)
	if err == nil {
		return fmt.Errorf("chain should not produce blocks after halt height")
	} else if timeoutCtx.Err() == nil {
		return fmt.Errorf("chain should not produce blocks after halt height")
	}

	height, err = c.Height(ctx)
	if err != nil {
		return err
	}

	// make sure that chain is halted; some chains may produce one more block after halt height
	if height-haltHeight > 1 {
		return fmt.Errorf("height %d is not within one block of halt height %d; chain isn't halted", height, haltHeight)
	}

	return c.ReplaceImagesAndRestart(ctx, version)
}

func (c *Chain) GetValidatorPower(ctx context.Context, hexaddr string) (int64, error) {
	var power int64
	err := CheckEndpoint(ctx, c.GetHostRPCAddress()+"/validators", func(b []byte) error {
		power = gjson.GetBytes(b, fmt.Sprintf("result.validators.#(address==\"%s\").voting_power", hexaddr)).Int()
		if power == 0 {
			return fmt.Errorf("validator %s power not found; validators are: %s", hexaddr, string(b))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return power, nil
}

func (c *Chain) GetValidatorHex(ctx context.Context, val int) (string, error) {
	json, err := c.Validators[val].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return "", err
	}
	providerHex := gjson.GetBytes(json, "address").String()
	return providerHex, nil
}

func getValidatorWallets(ctx context.Context, chain *Chain) ([]ValidatorWallet, error) {
	wallets := make([]ValidatorWallet, ValidatorCount)
	lock := new(sync.Mutex)
	eg := new(errgroup.Group)
	for i := 0; i < ValidatorCount; i++ {
		i := i
		eg.Go(func() error {
			// This moniker is hardcoded into the chain's genesis process.
			moniker := validatorMoniker
			address, err := chain.Validators[i].KeyBech32(ctx, moniker, "acc")
			if err != nil {
				return err
			}
			valoperAddress, err := chain.Validators[i].KeyBech32(ctx, moniker, "val")
			if err != nil {
				return err
			}
			valCons, _, err := chain.Validators[i].ExecBin(ctx, "comet", "show-address")
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			wallets[i] = ValidatorWallet{
				Moniker:        moniker,
				Address:        address,
				ValoperAddress: valoperAddress,
				ValConsAddress: strings.TrimSpace(string(valCons)),
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return wallets, nil
}

func (c *Chain) QueryJSON(ctx context.Context, jsonPath string, query ...string) (gjson.Result, error) {
	stdout, _, err := c.GetNode().ExecQuery(ctx, query...)
	if err != nil {
		return gjson.Result{}, err
	}
	retval := gjson.GetBytes(stdout, jsonPath)
	if !retval.Exists() {
		return gjson.Result{}, fmt.Errorf("json path %s not found in query result %s", jsonPath, stdout)
	}
	return retval, nil
}

// GetProposalID parses the proposal ID from the tx; necessary when the proposal type isn't accessible to interchaintest yet
func (c *Chain) GetProposalID(ctx context.Context, txhash string) (string, error) {
	stdout, _, err := c.GetNode().ExecQuery(ctx, "tx", txhash)
	if err != nil {
		return "", err
	}
	result := struct {
		Events []abcitypes.Event `json:"events"`
	}{}
	if err := json.Unmarshal(stdout, &result); err != nil {
		return "", err
	}
	for _, event := range result.Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "proposal_id" {
					return string(attr.Value), nil
				}
			}
		}
	}
	return "", fmt.Errorf("proposal ID not found in tx %s", txhash)
}
