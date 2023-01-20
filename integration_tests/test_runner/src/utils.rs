use althea_proto::cosmos_sdk_proto::cosmos::{
    bank::v1beta1::Metadata,
    base::abci::v1beta1::TxResponse,
    gov::v1beta1::VoteOption,
    params::v1beta1::{ParamChange, ParameterChangeProposal},
    staking::v1beta1::{DelegationResponse, QueryValidatorsRequest},
    upgrade::v1beta1::{Plan, SoftwareUpgradeProposal},
};
use bytes::BytesMut;

use clarity::{Address as EthAddress, Uint256};
use deep_space::address::Address as CosmosAddress;
use deep_space::client::ChainStatus;
use deep_space::coin::Coin;
use deep_space::error::CosmosGrpcError;
use deep_space::private_key::{CosmosPrivateKey, PrivateKey};
use deep_space::{Address, Contact, EthermintPrivateKey};
use futures::future::join_all;
use prost::{DecodeError, Message};
use prost_types::Any;
use rand::{rngs::ThreadRng, Rng};
use std::time::{Duration, Instant};
use std::{convert::TryInto, env};
use tokio::time::sleep;

use crate::type_urls::{PARAMETER_CHANGE_PROPOSAL_TYPE_URL, SOFTWARE_UPGRADE_PROPOSAL_TYPE_URL};

/// the timeout for individual requests
pub const OPERATION_TIMEOUT: Duration = Duration::from_secs(30);
/// the timeout for the total system
pub const TOTAL_TIMEOUT: Duration = Duration::from_secs(300);
// The config file location for hermes
pub const HERMES_CONFIG: &str = "/althea/tests/assets/ibc-relayer-config.toml";

/// this value reflects the contents of /tests/container-scripts/setup-validator.sh
/// and is used to compute if a stake change is big enough to trigger a validator set
/// update since we want to make several such changes intentionally
pub const STAKE_SUPPLY_PER_VALIDATOR: u128 = 1000000000;
/// this is the amount each validator bonds at startup
pub const STARTING_STAKE_PER_VALIDATOR: u128 = STAKE_SUPPLY_PER_VALIDATOR / 2;
// Retrieve values from runtime ENV vars
lazy_static! {
    // GRAVITY CHAIN CONSTANTS
    // These constants all apply to the gravity instance running (gravity-test-1)
    pub static ref ADDRESS_PREFIX: String =
        env::var("ADDRESS_PREFIX").unwrap_or_else(|_| "althea".to_string());
    pub static ref STAKING_TOKEN: String =
        env::var("STAKING_TOKEN").unwrap_or_else(|_| "ualtg".to_owned());
    pub static ref COSMOS_NODE_GRPC: String =
        env::var("COSMOS_NODE_GRPC").unwrap_or_else(|_| "http://localhost:9090".to_owned());
    pub static ref COSMOS_NODE_ABCI: String =
        env::var("COSMOS_NODE_ABCI").unwrap_or_else(|_| "http://localhost:26657".to_owned());

    // IBC CHAIN CONSTANTS
    // These constants all apply to the gaiad instance running (ibc-test-1)
    pub static ref IBC_ADDRESS_PREFIX: String =
        env::var("IBC_ADDRESS_PREFIX").unwrap_or_else(|_| "cosmos".to_string());
    pub static ref IBC_STAKING_TOKEN: String =
        env::var("IBC_STAKING_TOKEN").unwrap_or_else(|_| "stake".to_owned());
    pub static ref IBC_NODE_GRPC: String =
        env::var("IBC_NODE_GRPC").unwrap_or_else(|_| "http://localhost:9190".to_owned());
    pub static ref IBC_NODE_ABCI: String =
        env::var("IBC_NODE_ABCI").unwrap_or_else(|_| "http://localhost:27657".to_owned());

    // this is the key the IBC relayer will use to send IBC messages and channel updates
    // it's a distinct address to prevent sequence collisions
    pub static ref RELAYER_MNEMONIC: String = "below great use captain upon ship tiger exhaust orient burger network uphold wink theory focus cloud energy flavor recall joy phone beach symptom hobby".to_string();
    pub static ref RELAYER_PRIVATE_KEY: CosmosPrivateKey = CosmosPrivateKey::from_phrase(&RELAYER_MNEMONIC, "").unwrap();
    pub static ref RELAYER_ADDRESS: CosmosAddress = RELAYER_PRIVATE_KEY.to_address(ADDRESS_PREFIX.as_str()).unwrap();
}

/// Gets the standard non-token fee for the testnet. We deploy the test chain with STAKE
/// and FOOTOKEN balances by default, one footoken is sufficient for any Cosmos tx fee except
/// fees for send_to_eth messages which have to be of the same bridged denom so that the relayers
/// on the Ethereum side can be paid in that token.
pub fn get_fee(denom: Option<String>) -> Coin {
    match denom {
        None => Coin {
            denom: get_test_token_name(),
            amount: 1u32.into(),
        },
        Some(denom) => Coin {
            denom,
            amount: 1u32.into(),
        },
    }
}

pub fn get_deposit() -> Coin {
    Coin {
        denom: STAKING_TOKEN.to_string(),
        amount: 1_000_000_000u64.into(),
    }
}

pub fn get_test_token_name() -> String {
    "footoken".to_string()
}

/// Returns the chain-id of the gravity instance running, see GRAVITY CHAIN CONSTANTS above
pub fn get_chain_id() -> String {
    "althea-test-1".to_string()
}

/// Returns the chain-id of the gaiad instance running, see IBC CHAIN CONSTANTS above
pub fn get_ibc_chain_id() -> String {
    "ibc-test-1".to_string()
}

pub fn one_atom() -> Uint256 {
    one_atom_128().into()
}

pub fn one_atom_128() -> u128 {
    1000000u128
}

pub fn one_eth() -> Uint256 {
    one_eth_128().into()
}

pub fn one_eth_128() -> u128 {
    1000000000000000000u128
}

pub fn one_hundred_eth() -> Uint256 {
    (1000000000000000000u128 * 100).into()
}

/// returns the required denom metadata for deployed the Footoken
/// token defined in our test environment
pub async fn footoken_metadata(contact: &Contact) -> Metadata {
    let metadata = contact.get_all_denoms_metadata().await.unwrap();
    for m in metadata {
        if m.base == "footoken" {
            return m;
        }
    }
    panic!("Footoken metadata not set?");
}

pub fn get_decimals(meta: &Metadata) -> u32 {
    for m in meta.denom_units.iter() {
        if m.denom == meta.display {
            return m.exponent;
        }
    }
    panic!("Invalid metadata!")
}

pub fn get_coins(denom: &str, balances: &[Coin]) -> Option<Coin> {
    for coin in balances {
        if coin.denom.starts_with(denom) {
            return Some(coin.clone());
        }
    }
    None
}

/// This is a hardcoded very high gas value used in transaction stress test to counteract rollercoaster
/// gas prices due to the way that test fills blocks
pub const HIGH_GAS_PRICE: u64 = 1_000_000_000u64;

// Generates a new BridgeUserKey through randomly generated secrets
// cosmos_prefix allows for generation of a cosmos_address with a different prefix than "gravity"
pub fn get_user_key(cosmos_prefix: Option<&str>) -> CosmosUser {
    *bulk_get_user_keys(cosmos_prefix, 1).get(0).unwrap()
}

// Generates many CosmosUser keys + addresses
pub fn bulk_get_user_keys(cosmos_prefix: Option<&str>, num_users: i64) -> Vec<CosmosUser> {
    let cosmos_prefix = cosmos_prefix.unwrap_or(ADDRESS_PREFIX.as_str());

    let mut rng = rand::thread_rng();
    let mut users = Vec::with_capacity(num_users.try_into().unwrap());
    for _ in 0..num_users {
        let secret: [u8; 32] = rng.gen();
        let cosmos_key = CosmosPrivateKey::from_secret(&secret);
        let cosmos_address = cosmos_key.to_address(cosmos_prefix).unwrap();
        let user = CosmosUser {
            cosmos_address,
            cosmos_key,
        };

        users.push(user)
    }

    users
}

#[derive(Debug, Eq, PartialEq, Clone, Copy, Hash)]
pub struct CosmosUser {
    pub cosmos_address: CosmosAddress,
    pub cosmos_key: CosmosPrivateKey,
}

// Generates a new EthermintUserKey through a randomly generated secret
// cosmos_prefix allows for generation of a cosmos_address with a different prefix than "gravity"
pub fn get_ethermint_key(cosmos_prefix: Option<&str>) -> EthermintUserKey {
    let cosmos_prefix = cosmos_prefix.unwrap_or(ADDRESS_PREFIX.as_str());

    let mut rng = rand::thread_rng();
    let secret: [u8; 32] = rng.gen();
    // the starting location of the funds
    // the destination on cosmos that sends along to the final ethereum destination
    let ethermint_key = EthermintPrivateKey::from_secret(&secret);
    let ethermint_address = ethermint_key.to_address(cosmos_prefix).unwrap();
    // TODO: Verify that this conversion works like `evmosd debug addr`
    let eth_address = EthAddress::from_slice(ethermint_address.get_bytes()).unwrap();

    EthermintUserKey {
        ethermint_address,
        ethermint_key,
        eth_address,
    }
}

// Represents an Ethermint account, with address represented in the cosmos-sdk and Ethereum styles
#[derive(Debug, Eq, PartialEq, Clone, Copy)]
pub struct EthermintUserKey {
    pub ethermint_address: CosmosAddress, // the user's address according to ethsecp256k1
    pub ethermint_key: EthermintPrivateKey, // the user's private key
    pub eth_address: EthAddress,          // the ethermint_address treated as an EthAddress
}

#[derive(Debug, Clone)]
pub struct ValidatorKeys {
    /// The validator key used by this validator to actually sign and produce blocks
    pub validator_key: CosmosPrivateKey,
    // The mnemonic phrase used to generate validator_key
    pub validator_phrase: String,
}

/// Creates a proposal to change the params of our test chain
pub async fn create_parameter_change_proposal(
    contact: &Contact,
    key: impl PrivateKey,
    params_to_change: Vec<ParamChange>,
    fee_coin: Coin,
) {
    let proposal = ParameterChangeProposal {
        title: "Set althea settings!".to_string(),
        description: "test proposal".to_string(),
        changes: params_to_change,
    };
    let res = submit_parameter_change_proposal(
        proposal,
        get_deposit(),
        fee_coin,
        contact,
        key,
        Some(TOTAL_TIMEOUT),
    )
    .await
    .unwrap();
    trace!("Gov proposal executed with {:?}", res);
}

// Prints out current stake to the console
pub async fn print_validator_stake(contact: &Contact) {
    let validators = contact
        .get_validators_list(QueryValidatorsRequest::default())
        .await
        .unwrap();
    for validator in validators {
        info!(
            "Validator {} has {} tokens",
            validator.operator_address, validator.tokens
        );
    }
}

// Simple arguments to create a proposal with
pub struct UpgradeProposalParams {
    pub upgrade_height: i64,
    pub plan_name: String,
    pub plan_info: String,
    pub proposal_title: String,
    pub proposal_desc: String,
}

// Creates and submits a SoftwareUpgradeProposal to the chain, then votes yes with all validators
pub async fn execute_upgrade_proposal(
    contact: &Contact,
    keys: &[ValidatorKeys],
    timeout: Option<Duration>,
    upgrade_params: UpgradeProposalParams,
) {
    let duration = match timeout {
        Some(dur) => dur,
        None => OPERATION_TIMEOUT,
    };

    let plan = Plan {
        name: upgrade_params.plan_name,
        time: None,
        height: upgrade_params.upgrade_height,
        info: upgrade_params.plan_info,
    };
    let proposal = SoftwareUpgradeProposal {
        title: upgrade_params.proposal_title,
        description: upgrade_params.proposal_desc,
        plan: Some(plan),
    };
    let res = submit_upgrade_proposal(
        proposal,
        get_deposit(),
        get_fee(None),
        contact,
        keys[0].validator_key,
        Some(duration),
    )
    .await
    .unwrap();
    info!("Gov proposal executed with {:?}", res);

    vote_yes_on_proposals(contact, keys, None).await;
    wait_for_proposals_to_execute(contact).await;
}

// votes yes on every proposal available
pub async fn vote_yes_on_proposals(
    contact: &Contact,
    keys: &[ValidatorKeys],
    timeout: Option<Duration>,
) {
    let duration = match timeout {
        Some(dur) => dur,
        None => OPERATION_TIMEOUT,
    };
    // Vote yes on all proposals with all validators
    let proposals = contact
        .get_governance_proposals_in_voting_period()
        .await
        .unwrap();
    trace!("Found proposals: {:?}", proposals.proposals);
    let mut futs = Vec::new();
    for proposal in proposals.proposals {
        for key in keys.iter() {
            let res =
                vote_yes_with_retry(contact, proposal.proposal_id, key.validator_key, duration);
            futs.push(res);
        }
    }
    // vote on the proposal in parallel, reducing the number of blocks we wait for all
    // the tx's to get in.
    join_all(futs).await;
}

/// this utility function repeatedly attempts to vote yes on a governance
/// proposal up to MAX_VOTES times before failing
pub async fn vote_yes_with_retry(
    contact: &Contact,
    proposal_id: u64,
    key: impl PrivateKey,
    timeout: Duration,
) {
    const MAX_VOTES: u64 = 5;
    let mut counter = 0;
    let mut res = contact
        .vote_on_gov_proposal(
            proposal_id,
            VoteOption::Yes,
            get_fee(None),
            key.clone(),
            Some(timeout),
        )
        .await;
    while let Err(e) = res {
        contact.wait_for_next_block(TOTAL_TIMEOUT).await.unwrap();
        res = contact
            .vote_on_gov_proposal(
                proposal_id,
                VoteOption::Yes,
                get_fee(None),
                key.clone(),
                Some(timeout),
            )
            .await;
        counter += 1;
        if counter > MAX_VOTES {
            error!(
                "Vote for proposal has failed more than {} times, error {:?}",
                MAX_VOTES, e
            );
            panic!("failed to vote{}", e);
        }
    }
    let res = res.unwrap();
    info!(
        "Voting yes on governance proposal costing {} gas",
        res.gas_used
    );
}

// Checks that cosmos_account has each balance specified in expected_cosmos_coins.
// Note: ignores balances not in expected_cosmos_coins
pub async fn check_cosmos_balances(
    contact: &Contact,
    cosmos_account: CosmosAddress,
    expected_cosmos_coins: &[Coin],
) {
    let mut num_found = 0;

    let start = Instant::now();

    while Instant::now() - start < TOTAL_TIMEOUT {
        let mut good = true;
        let curr_balances = contact.get_balances(cosmos_account).await.unwrap();
        // These loops use loop labels, see the documentation on loop labels here for more information
        // https://doc.rust-lang.org/reference/expressions/loop-expr.html#loop-labels
        'outer: for bal in curr_balances.iter() {
            if num_found == expected_cosmos_coins.len() {
                break 'outer; // done searching entirely
            }
            'inner: for j in 0..expected_cosmos_coins.len() {
                if num_found == expected_cosmos_coins.len() {
                    break 'outer; // done searching entirely
                }
                if expected_cosmos_coins[j].denom != bal.denom {
                    continue;
                }
                let check = expected_cosmos_coins[j].amount == bal.amount;
                good = check;
                if !check {
                    warn!(
                        "found balance {}! expected {} trying again",
                        bal, expected_cosmos_coins[j].amount
                    );
                }
                num_found += 1;
                break 'inner; // done searching for this particular balance
            }
        }

        let check = num_found == curr_balances.len();
        // if it's already false don't set to true
        good = check || good;
        if !check {
            warn!(
                "did not find the correct balance for each expected coin! found {} of {}, trying again",
                num_found,
                curr_balances.len()
            );
        }
        if good {
            return;
        } else {
            sleep(Duration::from_secs(1)).await;
        }
    }
    panic!("Failed to find correct balances in check_cosmos_balances")
}

/// waits for the cosmos chain to start producing blocks, used to prevent race conditions
/// where our tests try to start running before the Cosmos chain is ready
pub async fn wait_for_cosmos_online(contact: &Contact, timeout: Duration) {
    let start = Instant::now();
    while let Err(CosmosGrpcError::NodeNotSynced) | Err(CosmosGrpcError::ChainNotRunning) =
        contact.wait_for_next_block(timeout).await
    {
        sleep(Duration::from_secs(1)).await;
        if Instant::now() - start > timeout {
            panic!("Cosmos node has not come online during timeout!")
        }
    }
    contact.wait_for_next_block(timeout).await.unwrap();
}

/// This function returns the valoper address of a validator
/// to whom delegating the returned amount of staking token will
/// create a 5% or greater change in voting power, triggering the
/// creation of a validator set update.
pub async fn get_validator_to_delegate_to(contact: &Contact) -> (CosmosAddress, Coin) {
    let validators = contact.get_active_validators().await.unwrap();
    let mut total_bonded_stake: Uint256 = 0u8.into();
    let mut has_the_least = None;
    let mut lowest = 0u8.into();
    for v in validators {
        let amount: Uint256 = v.tokens.parse().unwrap();
        total_bonded_stake += amount.clone();

        if lowest == 0u8.into() || amount < lowest {
            lowest = amount;
            has_the_least = Some(v.operator_address.parse().unwrap());
        }
    }

    // since this is five percent of the total bonded stake
    // delegating this to the validator who has the least should
    // do the trick
    let five_percent = total_bonded_stake / 20u8.into();
    let five_percent = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: five_percent,
    };

    (has_the_least.unwrap(), five_percent)
}

/// Waits for a particular block to be created
/// Returns an error if the chain fails to progress in a timely manner or the chain is not running
/// Panics if the block has already been surpassed
pub async fn wait_for_block(contact: &Contact, height: u64) -> Result<(), CosmosGrpcError> {
    let status = contact.get_chain_status().await?;
    let mut curr_height = match status {
        // Check the current height
        ChainStatus::Syncing => return Err(CosmosGrpcError::NodeNotSynced),
        ChainStatus::WaitingToStart => return Err(CosmosGrpcError::ChainNotRunning),
        ChainStatus::Moving { block_height } => {
            if block_height > height {
                panic!(
                    "Block height {} surpassed, current height is {}",
                    height, block_height
                );
            }
            block_height
        }
    };
    while curr_height < height {
        // Wait for the desired height
        contact.wait_for_next_block(OPERATION_TIMEOUT).await?; // Err if any block takes 30s+
        let new_status = contact.get_chain_status().await?;
        if let ChainStatus::Moving { block_height } = new_status {
            curr_height = block_height
        } else {
            // wait_for_next_block checks every second, so it's not likely the chain could halt for
            // an upgrade before we find the desired height
            return Err(CosmosGrpcError::BadResponse(
                "Wait for block: Chain was running and now it's not?".to_string(),
            ));
        }
    }
    Ok(())
}

/// Delegates `delegate_amount` to `delegate_to` and queries for confirmation of that delegation
/// Returns an error if the delegation or the query fail, returns the result of the delegation query
pub async fn delegate_and_confirm(
    contact: &Contact,
    user_key: impl PrivateKey,
    user_address: Address,
    delegate_to: Address,
    delegate_amount: Coin,
    fee_coin: Coin,
) -> Result<Option<DelegationResponse>, CosmosGrpcError> {
    let deleg_result = contact
        .delegate_to_validator(
            delegate_to,
            delegate_amount.clone(),
            fee_coin,
            user_key,
            Some(TOTAL_TIMEOUT),
        )
        .await;
    if deleg_result.is_err() {
        let err_str = format!(
            "Failed to delegate {} to validator {}, error {:?}",
            delegate_amount,
            delegate_to,
            deleg_result.unwrap_err()
        );
        error!("{}", err_str);
        return Err(CosmosGrpcError::BadResponse(err_str));
    }
    let deleg_confirm = contact.get_delegation(delegate_to, user_address).await;
    if deleg_confirm.is_err() {
        let err_str = format!(
            "Failed to query for delegation of {} to validator {}, error {:?}",
            delegate_amount,
            delegate_to,
            deleg_confirm.unwrap_err()
        );
        error!("{}", err_str);
        return Err(CosmosGrpcError::BadResponse(err_str));
    }
    Ok(deleg_confirm.unwrap())
}

/// Sends the given `amount` to each of `receivers` coming from `sender`
pub async fn send_funds_bulk(
    contact: &Contact,
    sender: impl PrivateKey,
    receivers: &[Address],
    amount: Coin,
    timeout: Option<Duration>,
) -> Result<(), CosmosGrpcError> {
    let fee = Some(Coin {
        denom: STAKING_TOKEN.clone(),
        amount: 0u8.into(),
    });
    for dest in receivers {
        contact
            .send_coins(amount.clone(), fee.clone(), *dest, timeout, sender.clone())
            .await?;
    }

    Ok(())
}

/// Waits up to TOTAL_TIMEOUT or provided timeout for the `user_address` account to gain at least `balance`
pub async fn wait_for_balance(
    contact: &Contact,
    user_address: Address,
    balance: Coin,
    timeout: Option<Duration>,
) {
    let duration = timeout.unwrap_or(TOTAL_TIMEOUT);
    let start = Instant::now();
    while Instant::now() - start < duration {
        let actual_balance = contact
            .get_balance(user_address, balance.denom.clone())
            .await;
        if let Ok(Some(bal)) = actual_balance {
            if bal.denom == balance.denom && bal.amount >= balance.amount {
                return;
            }
        }

        contact.wait_for_next_block(duration).await.unwrap();
    }

    panic!("User did not attain >= expected balance");
}

/// Encodes and submits a proposal change bridge parameters, should maybe be in deep_space
pub async fn submit_parameter_change_proposal(
    proposal: ParameterChangeProposal,
    deposit: Coin,
    fee: Coin,
    contact: &Contact,
    key: impl PrivateKey,
    wait_timeout: Option<Duration>,
) -> Result<TxResponse, CosmosGrpcError> {
    // encode as a generic proposal
    let any = encode_any(proposal, PARAMETER_CHANGE_PROPOSAL_TYPE_URL.to_string());
    contact
        .create_gov_proposal(any, deposit, fee, key, wait_timeout)
        .await
}

/// Encodes and submits a proposal to upgrade chain software, should maybe be in deep_space (sorry)
pub async fn submit_upgrade_proposal(
    proposal: SoftwareUpgradeProposal,
    deposit: Coin,
    fee: Coin,
    contact: &Contact,
    key: impl PrivateKey,
    wait_timeout: Option<Duration>,
) -> Result<TxResponse, CosmosGrpcError> {
    // encode as a generic proposal
    let any = encode_any(proposal, SOFTWARE_UPGRADE_PROPOSAL_TYPE_URL.to_string());
    contact
        .create_gov_proposal(any, deposit, fee, key, wait_timeout)
        .await
}

/// waits for the governance proposal to execute by waiting for it to leave
/// the 'voting' status
pub async fn wait_for_proposals_to_execute(contact: &Contact) {
    let start = Instant::now();
    loop {
        let proposals = contact
            .get_governance_proposals_in_voting_period()
            .await
            .unwrap();
        if Instant::now() - start > TOTAL_TIMEOUT {
            panic!("Gov proposal did not execute")
        } else if proposals.proposals.is_empty() {
            return;
        }
        sleep(Duration::from_secs(5)).await;
    }
}

/// Helper function for encoding the the proto any type
pub fn encode_any(input: impl prost::Message, type_url: impl Into<String>) -> Any {
    let mut value = Vec::new();
    input.encode(&mut value).unwrap();
    Any {
        type_url: type_url.into(),
        value,
    }
}

pub fn decode_any<T: Message + Default>(any: Any) -> Result<T, DecodeError> {
    let bytes = any.value;

    decode_bytes(bytes)
}

pub fn decode_bytes<T: Message + Default>(bytes: Vec<u8>) -> Result<T, DecodeError> {
    let mut buf = BytesMut::with_capacity(bytes.len());
    buf.extend_from_slice(&bytes);

    // Here we use the `T` type to decode whatever type of message this attestation holds
    // for use in the `f` function
    T::decode(buf)
}
