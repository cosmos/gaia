export const currentParams = {
  lastRelease: {
    releaseName: "v13",
    releaseDate: "",
    blockHeight: "",
    governanceProposalLink: "",
  },
  currentRelease: {
    releaseName: 'v14.1.0',
    releaseDate: '',
    blockHeight: '',
    governanceProposalLink: '',
    gaidExecutionOutput : 'name: gaia\nserver_name: gaiad\nversion: v14.1.0\ncommit: 0d9408e9169488707f1ad423e87d0df84a30431f\nbuild_tags: netgo,ledger\ngo: go version go1.20.10 darwin/arm64\nbuild_deps:\n- cosmossdk.io/api@v0.2.6\n- cosmossdk.io/core@v0.5.1',
    golangVersion : '1.20.x',
  },
  nextRelease: {
    releaseName: "v15",
    releaseDate: "",
    blockHeight: "",
    governanceProposalLink: "",
  },
  proposals: {
    numberOfValidatorsProp: "https://www.mintscan.io/cosmos/proposals/797",
  },
  auth: {
    max_memo_characters: "512",
    tx_sig_limit: "7",
    sig_verify_cost_ed25519: "590",
    tx_size_cost_per_byte: "10",
    sig_verify_cost_secp256k1: "1000",
  },
  baseapp: {
    BlockParams: {
      max_bytes: "200000",
      max_gas: "40000000",
    },
    EvidenceParams: {
      max_age_duration: "172800000000000",
      max_age_num_blocks: "1000000",
      max_bytes: "50000",
    },
    ValidatorParams: {
      pub_key_types: ["ed25519"],
    },
  },
  crisis: {
    ConstantFee: {
      amount: "1333000000",
      denom: "uatom",
    },
  },
  distribution: {
    base_proposer_reward: "0.010000000000000000",
    bonus_proposer_reward: "0.040000000000000000",
    community_tax: "0.020000000000000000",
    withdraw_addr_enabled: true,
  },
  gov: {
    deposit_params: {
      max_deposit_period: "1209600000000000",
      min_deposit: [
        {
          amount: "250000000",
          denom: "uatom",
        },
      ],
    },
    tally_params: {
      quorum: "0.400000000000000000",
      threshold: "0.500000000000000000",
      veto_threshold: "0.334000000000000000",
    },
    voting_params: {
      voting_period: "1209600000000000",
    },
    params: {
      min_deposit: [
        {
          denom: "stake",
          amount: "10000000",
        },
      ],
      max_deposit_period: "172800s",
      voting_period: "60s",
      quorum: "0.334000000000000000",
      threshold: "0.500000000000000000",
      veto_threshold: "0.334000000000000000",
      min_initial_deposit_ratio: "0.000000000000000000",
      burn_vote_quorum: false,
      burn_proposal_deposit_prevote: false,
      burn_vote_veto: true,
      min_deposit_ratio: "0.010000000000000000",
    },
  },
  mint: {
    blocks_per_year: "4360000",
    goal_bonded: "0.670000000000000000",
    inflation_max: "0.200000000000000000",
    inflation_min: "0.070000000000000000",
    inflation_rate_change: "1.000000000000000000",
    mint_denom: "uatom",
  },
  slashing: {
    downtime_jail_duration: "600000000000",
    min_signed_per_window: "0.050000000000000000",
    signed_blocks_window: "10000",
    slash_fraction_double_sign: "0.050000000000000000",
    slash_fraction_downtime: "0.000100000000000000",
  },
  staking: {
    unbonding_time: "86400s",
    max_validators: 100,
    max_entries: 7,
    historical_entries: 10000,
    bond_denom: "stake",
    min_commission_rate: "0.000000000000000000",
    validator_bond_factor: "-1.000000000000000000",
    global_liquid_staking_cap: "1.000000000000000000",
    validator_liquid_staking_cap: "1.000000000000000000",
  },
  transfer: {
    ReceiveEnabled: true,
    SendEnabled: true,
  },
};
