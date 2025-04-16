package chainsuite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.uber.org/multierr"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ccvclient "github.com/cosmos/interchain-security/v7/x/ccv/provider/client"

	sdkmath "cosmossdk.io/math"

	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
)

type ConsumerBootstrapCb func(ctx context.Context, consumer *cosmos.CosmosChain)

type ConsumerConfig struct {
	ChainName                       string
	Version                         string
	Denom                           string
	ShouldCopyProviderKey           []bool
	TopN                            int
	ValidatorSetCap                 int
	ValidatorPowerCap               int
	AllowInactiveVals               bool
	MinStake                        uint64
	Allowlist                       []string
	Denylist                        []string
	InitialHeight                   uint64
	DistributionTransmissionChannel string
	Spec                            *interchaintest.ChainSpec

	DuringDepositPeriod ConsumerBootstrapCb
	DuringVotingPeriod  ConsumerBootstrapCb
	BeforeSpawnTime     ConsumerBootstrapCb
	AfterSpawnTime      ConsumerBootstrapCb
}

type proposalWaiter struct {
	canDeposit chan struct{}
	isInVoting chan struct{}
	canVote    chan struct{}
	isPassed   chan struct{}
}

func (pw *proposalWaiter) waitForDepositAllowed() {
	<-pw.canDeposit
}

func (pw *proposalWaiter) startVotingPeriod() {
	close(pw.isInVoting)
}

func (pw *proposalWaiter) waitForVoteAllowed() {
	<-pw.canVote
}

func (pw *proposalWaiter) pass() {
	close(pw.isPassed)
}

func (pw *proposalWaiter) AllowVote() {
	close(pw.canVote)
}

func (pw *proposalWaiter) WaitForPassed() {
	<-pw.isPassed
}

func (pw *proposalWaiter) AllowDeposit() {
	close(pw.canDeposit)
}

func (pw *proposalWaiter) WaitForVotingPeriod() {
	<-pw.isInVoting
}

func newProposalWaiter() *proposalWaiter {
	return &proposalWaiter{
		canDeposit: make(chan struct{}),
		isInVoting: make(chan struct{}),
		canVote:    make(chan struct{}),
		isPassed:   make(chan struct{}),
	}
}

func (p *Chain) AddConsumerChain(ctx context.Context, relayer *Relayer, config ConsumerConfig) (*Chain, error) {
	dockerClient, dockerNetwork := GetDockerContext(ctx)

	if len(config.ShouldCopyProviderKey) < len(p.Validators) {
		return nil, fmt.Errorf("shouldCopyProviderKey should have at least %d elements", len(p.Validators))
	}

	spawnTime := time.Now().Add(ChainSpawnWait)
	// We need -test- in there because certain consumer IDs are hardcoded into the binary and we can't re-launch them
	chainID := fmt.Sprintf("%s-test-%d", config.ChainName, len(p.Consumers)+1)

	var proposalWaiter *proposalWaiter
	var errCh chan error
	if p.GetNode().HasCommand(ctx, "tx", "provider", "create-consumer") {
		errCh = make(chan error, 1)
		close(errCh)
		err := p.CreateConsumerPermissionless(ctx, chainID, config, spawnTime)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		proposalWaiter, errCh, err = p.SubmitConsumerAdditionProposal(ctx, chainID, config, spawnTime)
		if err != nil {
			return nil, err
		}
	}

	defaultSpec := p.DefaultConsumerChainSpec(ctx, chainID, config, spawnTime, proposalWaiter)
	config.Spec = MergeChainSpecs(defaultSpec, config.Spec)
	if config.Spec.InterchainSecurityConfig.ICSImageRepo == "" {
		config.Spec.InterchainSecurityConfig.ICSImageRepo = "ghcr.io/hyphacoop/ics"
	}
	// providerICS := p.GetNode().ICSVersion(ctx)
	// if config.Spec.InterchainSecurityConfig.ConsumerVerOverride == "" {
	// 	// This will disable the genesis transform
	// 	config.Spec.InterchainSecurityConfig.ConsumerVerOverride = providerICS
	// }
	cf := interchaintest.NewBuiltinChainFactory(
		GetLogger(ctx),
		[]*interchaintest.ChainSpec{config.Spec},
	)
	chains, err := cf.Chains(p.GetNode().TestName)
	if err != nil {
		return nil, err
	}
	cosmosConsumer := chains[0].(*cosmos.CosmosChain)

	// We can't use AddProviderConsumerLink here because the provider chain is already built; we'll have to do everything by hand.
	p.Consumers = append(p.Consumers, cosmosConsumer)
	cosmosConsumer.Provider = p.CosmosChain

	relayerWallet, err := cosmosConsumer.BuildRelayerWallet(ctx, "relayer-"+cosmosConsumer.Config().ChainID)
	if err != nil {
		return nil, err
	}
	wallets := make([]ibc.Wallet, len(p.Validators)+1)
	wallets[0] = relayerWallet
	// This is a hack, but we need to create wallets for the validators that have the right moniker.
	for i := 1; i <= len(p.Validators); i++ {
		wallets[i], err = cosmosConsumer.BuildRelayerWallet(ctx, validatorMoniker)
		if err != nil {
			return nil, err
		}
	}
	walletAmounts := make([]ibc.WalletAmount, len(wallets))
	for i, wallet := range wallets {
		walletAmounts[i] = ibc.WalletAmount{
			Address: wallet.FormattedAddress(),
			Denom:   cosmosConsumer.Config().Denom,
			Amount:  sdkmath.NewInt(ValidatorFunds),
		}
	}
	ic := interchaintest.NewInterchain().
		AddChain(cosmosConsumer, walletAmounts...).
		AddRelayer(relayer, "relayer")

	if err := ic.Build(ctx, GetRelayerExecReporter(ctx), interchaintest.InterchainBuildOptions{
		Client:    dockerClient,
		NetworkID: dockerNetwork,
		TestName:  p.GetNode().TestName,
	}); err != nil {
		return nil, err
	}

	// The chain should be built now, so we gotta check for errors in passing the proposal.
	if err := <-errCh; err != nil {
		return nil, err
	}

	for i, val := range cosmosConsumer.Validators {
		if err := val.RecoverKey(ctx, validatorMoniker, wallets[i+1].Mnemonic()); err != nil {
			return nil, err
		}
	}
	consumer, err := chainFromCosmosChain(cosmosConsumer, relayerWallet)
	if err != nil {
		return nil, err
	}

	err = relayer.SetupChainKeys(ctx, consumer)
	if err != nil {
		return nil, err
	}
	rep := GetRelayerExecReporter(ctx)
	if err := relayer.StopRelayer(ctx, rep); err != nil {
		return nil, err
	}
	if err := relayer.StartRelayer(ctx, rep); err != nil {
		return nil, err
	}
	err = relayer.ConnectProviderConsumer(ctx, p, consumer)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (p *Chain) CreateConsumerPermissionless(ctx context.Context, chainID string, config ConsumerConfig, spawnTime time.Time) error {
	revisionHeight := config.InitialHeight
	if revisionHeight == 0 {
		revisionHeight = 1
	}
	initParams := &providertypes.ConsumerInitializationParameters{
		InitialHeight:                     clienttypes.Height{RevisionNumber: clienttypes.ParseChainID(chainID), RevisionHeight: revisionHeight},
		SpawnTime:                         spawnTime,
		BlocksPerDistributionTransmission: BlocksPerDistribution,
		CcvTimeoutPeriod:                  2419200000000000,
		TransferTimeoutPeriod:             3600000000000,
		ConsumerRedistributionFraction:    "0.75",
		HistoricalEntries:                 10000,
		UnbondingPeriod:                   1728000000000000,
		GenesisHash:                       []byte("Z2VuX2hhc2g="),
		BinaryHash:                        []byte("YmluX2hhc2g="),
		DistributionTransmissionChannel:   config.DistributionTransmissionChannel,
	}
	powerShapingParams := &providertypes.PowerShapingParameters{
		Top_N:              0,
		ValidatorSetCap:    uint32(config.ValidatorSetCap),
		ValidatorsPowerCap: uint32(config.ValidatorPowerCap),
		AllowInactiveVals:  config.AllowInactiveVals,
		MinStake:           config.MinStake,
		Allowlist:          config.Allowlist,
		Denylist:           config.Denylist,
	}
	params := providertypes.MsgCreateConsumer{
		ChainId: chainID,
		Metadata: providertypes.ConsumerMetadata{
			Name:        config.ChainName,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
		InitializationParameters: initParams,
		PowerShapingParameters:   powerShapingParams,
	}

	paramsBz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	err = p.GetNode().WriteFile(ctx, paramsBz, "consumer-addition.json")
	if err != nil {
		return err
	}
	_, err = p.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName, "provider", "create-consumer", path.Join(p.GetNode().HomeDir(), "consumer-addition.json"))
	if err != nil {
		return err
	}
	if config.TopN > 0 {
		govAddress, err := p.GetGovernanceAddress(ctx)
		if err != nil {
			return err
		}
		consumerID, err := p.QueryJSON(ctx, fmt.Sprintf("chains.#(chain_id=%q).consumer_id", chainID), "provider", "list-consumer-chains")
		if err != nil {
			return err
		}
		update := &providertypes.MsgUpdateConsumer{
			ConsumerId:      consumerID.String(),
			NewOwnerAddress: govAddress,
			Metadata: &providertypes.ConsumerMetadata{
				Name:        config.ChainName,
				Description: "Consumer chain",
				Metadata:    "ipfs://",
			},
			InitializationParameters: initParams,
			PowerShapingParameters:   powerShapingParams,
		}
		updateBz, err := json.Marshal(update)
		if err != nil {
			return err
		}
		err = p.GetNode().WriteFile(ctx, updateBz, "consumer-update.json")
		if err != nil {
			return err
		}
		_, err = p.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName, "provider", "update-consumer", path.Join(p.GetNode().HomeDir(), "consumer-update.json"))
		if err != nil {
			return err
		}
		powerShapingParams.Top_N = uint32(config.TopN)
		update = &providertypes.MsgUpdateConsumer{
			Owner:      govAddress,
			ConsumerId: consumerID.String(),
			Metadata: &providertypes.ConsumerMetadata{
				Name:        config.ChainName,
				Description: "Consumer chain",
				Metadata:    "ipfs://",
			},
			InitializationParameters: initParams,
			PowerShapingParameters:   powerShapingParams,
		}
		prop, err := p.BuildProposal([]cosmos.ProtoMessage{update}, "update consumer", "update consumer", "", GovDepositAmount, "", false)
		if err != nil {
			return err
		}
		txhash, err := p.GetNode().SubmitProposal(ctx, p.ValidatorWallets[0].Moniker, prop)
		if err != nil {
			return err
		}
		propID, err := p.GetProposalID(ctx, txhash)
		if err != nil {
			return err
		}
		if err := p.PassProposal(ctx, propID); err != nil {
			return err
		}

	}
	return nil
}

func (p *Chain) DefaultConsumerChainSpec(ctx context.Context, chainID string, config ConsumerConfig, spawnTime time.Time, proposalWaiter *proposalWaiter) *interchaintest.ChainSpec {
	const (
		strideChain  = "stride"
		icsConsumer  = "ics-consumer"
		neutronChain = "neutron"
	)
	fullNodes := len(p.FullNodes)
	validators := len(p.Validators)

	chainType := config.ChainName
	version := config.Version
	denom := config.Denom
	shouldCopyProviderKey := config.ShouldCopyProviderKey

	bechPrefix := ""
	if chainType == icsConsumer {
		majorVersion, err := strconv.Atoi(version[1:2])
		if err != nil {
			// this really shouldn't happen unless someone misconfigured something
			panic(fmt.Sprintf("failed to parse major version from %s: %v", version, err))
		}
		if majorVersion >= 4 {
			bechPrefix = "consumer"
		}
	} else if chainType == strideChain {
		bechPrefix = "stride"
	}
	genesisOverrides := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.slashing.params.signed_blocks_window", strconv.Itoa(SlashingWindowConsumer)),
		cosmos.NewGenesisKV("app_state.ccvconsumer.params.reward_denoms", []string{denom}),
		cosmos.NewGenesisKV("app_state.ccvconsumer.params.provider_reward_denoms", []string{p.Config().Denom}),
		cosmos.NewGenesisKV("app_state.ccvconsumer.params.blocks_per_distribution_transmission", BlocksPerDistribution),
	}
	if config.TopN >= 0 {
		genesisOverrides = append(genesisOverrides, cosmos.NewGenesisKV("app_state.ccvconsumer.params.soft_opt_out_threshold", "0.0"))
	}
	if chainType == neutronChain {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV("app_state.globalfee.params.minimum_gas_prices", []interface{}{
				map[string]interface{}{
					"amount": "0.005",
					"denom":  denom,
				},
			}),
		)
	}

	if chainType == strideChain {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV("app_state.gov.params.voting_period", GovVotingPeriod.String()),
		)
	}
	modifyGenesis := func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		b, err := cosmos.ModifyGenesis(genesisOverrides)(cc, b)
		if err != nil {
			return nil, err
		}
		if chainType == neutronChain || chainType == strideChain {
			// Stride and Neutron aren't updated yet to use the consumer ID
			b, err = sjson.DeleteBytes(b, "app_state.ccvconsumer.params.consumer_id")
			if err != nil {
				return nil, err
			}
		}
		if chainType == strideChain {
			b, err = sjson.SetBytes(b, "app_state.epochs.epochs.#(identifier==\"day\").duration", "120s")
			if err != nil {
				return nil, err
			}
			b, err = sjson.SetBytes(b, "app_state.epochs.epochs.#(identifier==\"stride_epoch\").duration", "30s")
			if err != nil {
				return nil, err
			}
		}
		if gjson.GetBytes(b, "consensus").Exists() {
			return sjson.SetBytes(b, "consensus.block.max_gas", "50000000")
		}
		return sjson.SetBytes(b, "consensus_params.block.max_gas", "50000000")
	}

	return &interchaintest.ChainSpec{
		Name:          chainType,
		Version:       version,
		ChainName:     chainID,
		NumFullNodes:  &fullNodes,
		NumValidators: &validators,
		ChainConfig: ibc.ChainConfig{
			Denom:         denom,
			GasPrices:     "0.005" + denom,
			GasAdjustment: 2.0,
			Gas:           "auto",
			ChainID:       chainID,
			ConfigFileOverrides: map[string]any{
				"config/config.toml": DefaultConfigToml(),
			},
			PreGenesis: func(consumer ibc.Chain) error {
				if config.DuringDepositPeriod != nil {
					config.DuringDepositPeriod(ctx, consumer.(*cosmos.CosmosChain))
				}
				if proposalWaiter != nil {
					proposalWaiter.AllowDeposit()
					proposalWaiter.WaitForVotingPeriod()
				}
				if config.DuringVotingPeriod != nil {
					config.DuringVotingPeriod(ctx, consumer.(*cosmos.CosmosChain))
				}
				if proposalWaiter != nil {
					proposalWaiter.AllowVote()
					proposalWaiter.WaitForPassed()
				}
				tCtx, tCancel := context.WithDeadline(ctx, spawnTime)
				defer tCancel()
				if config.BeforeSpawnTime != nil {
					config.BeforeSpawnTime(tCtx, consumer.(*cosmos.CosmosChain))
				}
				// interchaintest will set up the validator keys right before PreGenesis.
				// Now we just need to wait for the chain to spawn before interchaintest can get the ccv file.
				// This wait is here and not there because of changes we've made to interchaintest that need to be upstreamed in an orderly way.
				GetLogger(ctx).Sugar().Infof("waiting for chain %s to spawn at %s", chainID, spawnTime)
				<-tCtx.Done()
				if err := testutil.WaitForBlocks(ctx, 2, p); err != nil {
					return err
				}
				if config.AfterSpawnTime != nil {
					config.AfterSpawnTime(ctx, consumer.(*cosmos.CosmosChain))
				}
				return nil
			},
			Bech32Prefix:         bechPrefix,
			ModifyGenesisAmounts: DefaultGenesisAmounts(denom),
			ModifyGenesis:        modifyGenesis,
			InterchainSecurityConfig: ibc.ICSConfig{
				ConsumerCopyProviderKey: func(i int) bool {
					return shouldCopyProviderKey[i]
				},
			},
		},
	}
}

func (p *Chain) SubmitConsumerAdditionProposal(ctx context.Context, chainID string, config ConsumerConfig, spawnTime time.Time) (*proposalWaiter, chan error, error) {
	propWaiter := newProposalWaiter()
	prop := p.buildConsumerAdditionJSON(chainID, config, spawnTime)
	propTx, err := p.ConsumerAdditionProposal(ctx, interchaintest.FaucetAccountKeyName, prop)
	if err != nil {
		return nil, nil, err
	}
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := p.WaitForProposalStatus(ctx, propTx.ProposalID, govv1.StatusDepositPeriod); err != nil {
			errCh <- err
			panic(err)
		}
		propWaiter.waitForDepositAllowed()

		if _, err := p.GetNode().ExecTx(ctx, interchaintest.FaucetAccountKeyName, "gov", "deposit", propTx.ProposalID, prop.Deposit); err != nil {
			errCh <- err
			panic(err)
		}

		if err := p.WaitForProposalStatus(ctx, propTx.ProposalID, govv1.StatusVotingPeriod); err != nil {
			errCh <- err
			panic(err)
		}
		propWaiter.startVotingPeriod()
		propWaiter.waitForVoteAllowed()

		if err := p.PassProposal(ctx, propTx.ProposalID); err != nil {
			errCh <- err
			panic(err)
		}
		propWaiter.pass()
	}()
	return propWaiter, errCh, nil
}

func (p *Chain) buildConsumerAdditionJSON(chainID string, config ConsumerConfig, spawnTime time.Time) ccvclient.ConsumerAdditionProposalJSON {
	prop := ccvclient.ConsumerAdditionProposalJSON{
		Title:         fmt.Sprintf("Addition of %s consumer chain", chainID),
		Summary:       "Proposal to add new consumer chain",
		ChainId:       chainID,
		InitialHeight: clienttypes.Height{RevisionNumber: clienttypes.ParseChainID(chainID), RevisionHeight: 1},
		GenesisHash:   []byte("gen_hash"),
		BinaryHash:    []byte("bin_hash"),
		SpawnTime:     spawnTime,

		BlocksPerDistributionTransmission: BlocksPerDistribution,
		CcvTimeoutPeriod:                  2419200000000000,
		TransferTimeoutPeriod:             3600000000000,
		ConsumerRedistributionFraction:    "0.75",
		HistoricalEntries:                 10000,
		UnbondingPeriod:                   1728000000000000,
		Deposit:                           strconv.Itoa(GovMinDepositAmount/2) + p.Config().Denom,
	}
	if config.TopN >= 0 {
		prop.TopN = uint32(config.TopN)
	}
	if config.ValidatorSetCap > 0 {
		prop.ValidatorSetCap = uint32(config.ValidatorSetCap)
	}
	if config.ValidatorPowerCap > 0 {
		prop.ValidatorsPowerCap = uint32(config.ValidatorPowerCap)
	}
	if config.AllowInactiveVals {
		prop.AllowInactiveVals = true
	}
	if config.MinStake > 0 {
		prop.MinStake = config.MinStake
	}
	return prop
}

func (p *Chain) CheckCCV(ctx context.Context, consumer *Chain, relayer *Relayer, amount, valIdx, blocksPerEpoch int) error {
	providerAddress := p.ValidatorWallets[valIdx]

	json, err := p.Validators[valIdx].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return err
	}
	providerHex := gjson.GetBytes(json, "address").String()
	json, err = consumer.Validators[valIdx].ReadFile(ctx, "config/priv_validator_key.json")
	if err != nil {
		return err
	}
	consumerHex := gjson.GetBytes(json, "address").String()

	providerPowerBefore, err := p.GetValidatorPower(ctx, providerHex)
	if err != nil {
		return err
	}

	if err := p.Validators[valIdx].StakingDelegate(ctx, providerAddress.Moniker, providerAddress.ValoperAddress, fmt.Sprintf("%d%s", amount, p.Config().Denom)); err != nil {
		return err
	}

	if blocksPerEpoch > 1 {
		providerPower, err := p.GetValidatorPower(ctx, providerHex)
		if err != nil {
			return err
		}
		if providerPowerBefore >= providerPower {
			return errors.New("provider power did not increase after delegation")
		}
		consumerPower, err := consumer.GetValidatorPower(ctx, consumerHex)
		if err != nil {
			return err
		}
		if providerPower == consumerPower {
			return fmt.Errorf("consumer power updated too soon")
		}
		if err := testutil.WaitForBlocks(ctx, blocksPerEpoch, p); err != nil {
			return err
		}
	}

	if err := relayer.ClearCCVChannel(ctx, p, consumer); err != nil {
		return err
	}
	if err := testutil.WaitForBlocks(ctx, 2, p, consumer); err != nil {
		return err
	}

	tCtx, tCancel := context.WithTimeout(ctx, 15*time.Minute)
	defer tCancel()
	var retErr error
	for tCtx.Err() == nil {
		retErr = nil
		providerPower, err := p.GetValidatorPower(ctx, providerHex)
		if err != nil {
			return err
		}
		consumerPower, err := consumer.GetValidatorPower(ctx, consumerHex)
		if err != nil {
			return err
		}
		if providerPowerBefore >= providerPower {
			retErr = fmt.Errorf("provider power did not increase after delegation")
		} else if providerPower != consumerPower {
			retErr = fmt.Errorf("consumer power did not update after provider delegation")
		}
		if retErr == nil {
			break
		}
		time.Sleep(CommitTimeout)
	}
	return retErr
}

func (p *Chain) IsValoperJailed(ctx context.Context, valoper string) (bool, error) {
	out, _, err := p.Validators[0].ExecQuery(ctx, "staking", "validator", valoper)
	if err != nil {
		return false, err
	}
	if gjson.GetBytes(out, "jailed").Exists() {
		return gjson.GetBytes(out, "jailed").Bool(), nil
	}
	return gjson.GetBytes(out, "validator.jailed").Bool(), nil
}

func (p *Chain) IsValidatorJailedForConsumerDowntime(ctx context.Context, relayer *Relayer, consumer *Chain, validatorIdx int) (jailed bool, err error) {
	if err = consumer.Validators[validatorIdx].PauseContainer(ctx); err != nil {
		return
	}
	defer func() {
		sErr := consumer.Validators[validatorIdx].UnpauseContainer(ctx)
		if sErr != nil {
			err = multierr.Append(err, sErr)
			return
		}
		time.Sleep(10 * CommitTimeout)
		if jailed && err == nil {
			if _, err = p.Validators[validatorIdx].ExecTx(ctx, p.ValidatorWallets[validatorIdx].Moniker, "slashing", "unjail"); err != nil {
				return
			}
			var stillJailed bool
			if stillJailed, err = p.IsValoperJailed(ctx, p.ValidatorWallets[validatorIdx].ValoperAddress); stillJailed {
				err = fmt.Errorf("validator %d is still jailed after unjailing", validatorIdx)
			}
		}
	}()
	if p.Config().ChainID != consumer.Config().ChainID {
		if err := relayer.ClearCCVChannel(ctx, p, consumer); err != nil {
			return false, err
		}
		tCtx, tCancel := context.WithTimeout(ctx, SlashingWindowConsumer*2*CommitTimeout)
		defer tCancel()
		if err = testutil.WaitForBlocks(tCtx, SlashingWindowConsumer+1, consumer); err != nil {
			if tCtx.Err() != nil {
				err = fmt.Errorf("chain %s is stopped: %w", consumer.Config().ChainID, err)
			}
			return
		}
		if err := relayer.ClearCCVChannel(ctx, p, consumer); err != nil {
			return false, err
		}
	}
	tCtx, tCancel := context.WithTimeout(ctx, 30*CommitTimeout)
	defer tCancel()
	for tCtx.Err() == nil {
		jailed, err = p.IsValoperJailed(ctx, p.ValidatorWallets[validatorIdx].ValoperAddress)
		if err != nil || jailed {
			return
		}
		time.Sleep(CommitTimeout)
	}
	return false, nil
}

func (c *Chain) GetConsumerID(ctx context.Context, consumerID string) (string, error) {
	consumerIDJSON, err := c.QueryJSON(ctx, fmt.Sprintf("chains.#(chain_id=%q).consumer_id", consumerID), "provider", "list-consumer-chains")
	if err != nil {
		return "", err
	}
	return consumerIDJSON.String(), nil
}
