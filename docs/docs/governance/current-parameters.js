export const currentParams = {
  lastRelease: {
    releaseName: "v11",
    releaseDate: "",
    blockHeight: "",
    governanceProposalLink: "",
  },
  currentRelease: {
    releaseName: "v12",
    releaseDate: "",
    blockHeight: "",
    governanceProposalLink: "",
    gaidExecutionOutput:
      "❯ gaiad version --long\nname: gaia\nserver_name: gaiad\nversion: v12.0.0\ncommit: 6f8067d76ce30996f83645862153ccfaf5f13dd1\nbuild_tags: netgo,ledger\ngo: go version go1.20.4 darwin/arm64\nbuild_deps:\n- cosmossdk.io/api@v0.2.6\n- cosmossdk.io/core@v0.5.1\n- cosmossdk.io/depinject@v1.0.0-alpha.3\n- cosmossdk.io/errors@v1.0.0\n- filippo.io/edwards25519@v1.0.0-rc.1\n- github.com/99designs/go-keychain@v0.0.0-20191008050251-8e49817e8af4\n- github.com/99designs/keyring@v1.2.1 => github.com/cosmos/keyring@v1.2.0\n- github.com/ChainSafe/go-schnorrkel@v1.0.0\n- github.com/Workiva/go-datastructures@v1.0.53\n- github.com/armon/go-metrics@v0.4.1\n- github.com/beorn7/perks@v1.0.1\n- github.com/bgentry/speakeasy@v0.1.1-0.20220910012023-760eaf8b6816\n- github.com/btcsuite/btcd/btcec/v2@v2.3.2\n- github.com/cenkalti/backoff/v4@v4.1.3\n- github.com/cespare/xxhash/v2@v2.1.2\n- github.com/coinbase/rosetta-sdk-go@v0.7.9\n- github.com/cometbft/cometbft-db@v0.7.0\n- github.com/confio/ics23/go@v0.9.0\n- github.com/cosmos/btcutil@v1.0.4\n- github.com/cosmos/cosmos-db@v0.0.0-20221226095112-f3c38ecb5e32\n- github.com/cosmos/cosmos-proto@v1.0.0-beta.1\n- github.com/cosmos/cosmos-sdk@v0.45.16-ics => github.com/cosmos/cosmos-sdk@v0.45.16-ics-lsm\n- github.com/cosmos/go-bip39@v1.0.0\n- github.com/cosmos/iavl@v0.19.5\n- github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4@v4.1.0\n- github.com/cosmos/ibc-go/v4@v4.4.2\n- github.com/cosmos/interchain-security/v2@v2.0.0 => github.com/cosmos/interchain-security/v2@v2.0.0-lsm\n- github.com/cosmos/ledger-cosmos-go@v0.12.2\n- github.com/creachadair/taskgroup@v0.3.2\n- github.com/davecgh/go-spew@v1.1.1\n- github.com/decred/dcrd/dcrec/secp256k1/v4@v4.0.1\n- github.com/desertbit/timer@v0.0.0-20180107155436-c41aec40b27f\n- github.com/dvsekhvalnov/jose2go@v1.5.0\n- github.com/felixge/httpsnoop@v1.0.2\n- github.com/fsnotify/fsnotify@v1.6.0\n- github.com/go-kit/kit@v0.12.0\n- github.com/go-kit/log@v0.2.1\n- github.com/go-logfmt/logfmt@v0.5.1\n- github.com/gogo/gateway@v1.1.0\n- github.com/gogo/protobuf@v1.3.3 => github.com/regen-network/protobuf@v1.3.3-alpha.regen.1\n- github.com/golang/protobuf@v1.5.3\n- github.com/golang/snappy@v0.0.4\n- github.com/google/btree@v1.1.2\n- github.com/google/orderedcode@v0.0.1\n- github.com/gorilla/handlers@v1.5.1\n- github.com/gorilla/mux@v1.8.0\n- github.com/gorilla/websocket@v1.5.0\n- github.com/gravity-devs/liquidity@v1.6.0 => github.com/gravity-devs/liquidity@v1.6.0-forced-withdrawal\n- github.com/grpc-ecosystem/go-grpc-middleware@v1.3.0\n- github.com/grpc-ecosystem/grpc-gateway@v1.16.0\n- github.com/grpc-ecosystem/grpc-gateway/v2@v2.10.2\n- github.com/gtank/merlin@v0.1.1\n- github.com/gtank/ristretto255@v0.1.2\n- github.com/hashicorp/go-immutable-radix@v1.3.1\n- github.com/hashicorp/golang-lru@v0.5.5-0.20210104140557-80c98217689d\n- github.com/hashicorp/hcl@v1.0.0\n- github.com/hdevalence/ed25519consensus@v0.0.0-20220222234857-c00d1f31bab3\n- github.com/iancoleman/orderedmap@v0.2.0\n- github.com/improbable-eng/grpc-web@v0.15.0\n- github.com/klauspost/compress@v1.15.11\n- github.com/lib/pq@v1.10.7\n- github.com/libp2p/go-buffer-pool@v0.1.0\n- github.com/magiconair/properties@v1.8.7\n- github.com/mattn/go-colorable@v0.1.13\n- github.com/mattn/go-isatty@v0.0.17\n- github.com/matttproud/golang_protobuf_extensions@v1.0.2-0.20181231171920-c182affec369\n- github.com/mimoo/StrobeGo@v0.0.0-20210601165009-122bf33a46e0\n- github.com/minio/highwayhash@v1.0.2\n- github.com/mitchellh/mapstructure@v1.5.0\n- github.com/mtibben/percent@v0.2.1\n- github.com/pelletier/go-toml/v2@v2.0.8\n- github.com/pkg/errors@v0.9.1\n- github.com/pmezard/go-difflib@v1.0.0\n- github.com/prometheus/client_golang@v1.14.0\n- github.com/prometheus/client_model@v0.3.0\n- github.com/prometheus/common@v0.37.0\n- github.com/prometheus/procfs@v0.8.0\n- github.com/rakyll/statik@v0.1.7\n- github.com/rcrowley/go-metrics@v0.0.0-20201227073835-cf1acfcdf475\n- github.com/regen-network/cosmos-proto@v0.3.1\n- github.com/rs/cors@v1.8.2\n- github.com/rs/zerolog@v1.27.0\n- github.com/spf13/afero@v1.9.5\n- github.com/spf13/cast@v1.5.1\n- github.com/spf13/cobra@v1.7.0\n- github.com/spf13/jwalterweatherman@v1.1.0\n- github.com/spf13/pflag@v1.0.5\n- github.com/spf13/viper@v1.16.0\n- github.com/stretchr/testify@v1.8.4\n- github.com/subosito/gotenv@v1.4.2\n- github.com/syndtr/goleveldb@v1.0.1-0.20210819022825-2ae1ddf74ef7\n- github.com/tendermint/go-amino@v0.16.0\n- github.com/tendermint/tendermint@v0.34.27 => github.com/cometbft/cometbft@v0.34.29\n- github.com/tendermint/tm-db@v0.6.7\n- github.com/tidwall/btree@v1.5.0\n- github.com/zondax/hid@v0.9.1\n- github.com/zondax/ledger-go@v0.14.1\n- golang.org/x/crypto@v0.11.0\n- golang.org/x/exp@v0.0.0-20221205204356-47842c84f3db\n- golang.org/x/net@v0.12.0\n- golang.org/x/sys@v0.10.0\n- golang.org/x/term@v0.10.0\n- golang.org/x/text@v0.11.0\n- google.golang.org/genproto@v0.0.0-20230410155749-daa745c078e1\n- google.golang.org/grpc@v1.56.2 => google.golang.org/grpc@v1.33.2\n- google.golang.org/protobuf@v1.31.0\n- gopkg.in/ini.v1@v1.67.0\n- gopkg.in/yaml.v2@v2.4.0\n- gopkg.in/yaml.v3@v3.0.1\n- nhooyr.io/websocket@v1.8.6\ncosmos_sdk_version: v0.45.16-ics",
    golangVersion: "1.20.x",
  },
  nextRelease: {
    releaseName: "v12",
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
