//! This module contains the query message handlers

use cosmwasm_std::{to_json_binary, Binary, Deps, Env};
use ethereum_light_client::consensus_state::TrustedConsensusState;

use crate::{
    custom_query::{BlsVerifier, EthereumCustomQuery},
    msg::{
        CheckForMisbehaviourMsg, CheckForMisbehaviourResult, EthereumMisbehaviourMsg, StatusResult,
        TimestampAtHeightMsg, TimestampAtHeightResult, VerifyClientMessageMsg,
    },
    state::{get_eth_client_state, get_eth_consensus_state},
    ContractError,
};

/// Verifies the client message (header) that will be used for updating the state of the light client
/// The actual verification logic is done in the ethereum light client package
/// # Errors
/// Returns an error if the client message is invalid
/// # Returns
/// An empty response
#[allow(clippy::needless_pass_by_value)]
pub fn verify_client_message(
    deps: Deps<EthereumCustomQuery>,
    env: Env,
    verify_client_message_msg: VerifyClientMessageMsg,
) -> Result<Binary, ContractError> {
    let eth_client_state = get_eth_client_state(deps.storage)?;

    let bls_verifier = BlsVerifier {
        querier: deps.querier,
    };

    if let Ok(header) = serde_json::from_slice(&verify_client_message_msg.client_message) {
        let eth_consensus_state =
            get_eth_consensus_state(deps.storage, eth_client_state.latest_slot)?;

        ethereum_light_client::verify::verify_header(
            &eth_consensus_state,
            &eth_client_state,
            env.block.time.seconds(),
            &header,
            bls_verifier,
        )
        .map_err(ContractError::VerifyClientMessageFailed)?;

        return Ok(Binary::default());
    }

    if let Ok(misbehaviour) =
        serde_json::from_slice::<EthereumMisbehaviourMsg>(&verify_client_message_msg.client_message)
    {
        let eth_consensus_state = get_eth_consensus_state(deps.storage, misbehaviour.trusted_slot)?;

        ethereum_light_client::misbehaviour::verify_misbehaviour(
            &eth_client_state,
            &TrustedConsensusState {
                state: eth_consensus_state,
                sync_committee: misbehaviour.sync_committee,
            },
            &misbehaviour.update_1,
            &misbehaviour.update_2,
            env.block.time.seconds(),
            bls_verifier,
        )
        .map_err(ContractError::VerifyClientMessageFailed)?;

        return Ok(Binary::default());
    }

    Err(ContractError::InvalidClientMessage)
}

/// Checks for misbehaviour. Returning an error means no misbehaviour was found.
///
/// Note that we are replicating some of the logic of `verify_client_message` here, ideally we
/// would also check for misbehavior of the header in this function.
/// # Errors
/// Returns an error if the misbehaviour cannot be verified
#[allow(clippy::needless_pass_by_value)]
pub fn check_for_misbehaviour(
    deps: Deps<EthereumCustomQuery>,
    env: Env,
    check_for_misbehaviour_msg: CheckForMisbehaviourMsg,
) -> Result<Binary, ContractError> {
    let misbehaviour = serde_json::from_slice::<EthereumMisbehaviourMsg>(
        &check_for_misbehaviour_msg.client_message,
    )
    .map_err(ContractError::DeserializeEthMisbehaviourFailed)?;

    let eth_client_state = get_eth_client_state(deps.storage)?;
    let eth_consensus_state = get_eth_consensus_state(deps.storage, misbehaviour.trusted_slot)?;

    let bls_verifier = BlsVerifier {
        querier: deps.querier,
    };

    ethereum_light_client::misbehaviour::verify_misbehaviour(
        &eth_client_state,
        &TrustedConsensusState {
            state: eth_consensus_state,
            sync_committee: misbehaviour.sync_committee,
        },
        &misbehaviour.update_1,
        &misbehaviour.update_2,
        env.block.time.seconds(),
        bls_verifier,
    )
    .map_err(ContractError::VerifyClientMessageFailed)?;

    Ok(to_json_binary(&CheckForMisbehaviourResult {
        found_misbehaviour: true,
    })?)
}

/// Gets the consensus timestamp at a given height
/// # Errors
/// Returns an error if the conensus state is not found
/// # Returns
/// The timestamp at the given height
#[allow(clippy::needless_pass_by_value)]
pub fn timestamp_at_height(
    deps: Deps<EthereumCustomQuery>,
    timestamp_at_height_msg: TimestampAtHeightMsg,
) -> Result<Binary, ContractError> {
    let eth_consensus_state =
        get_eth_consensus_state(deps.storage, timestamp_at_height_msg.height.revision_height)?;

    let nano_timestamp = eth_consensus_state.timestamp * 1_000_000_000; // ibc-go expects nanoseconds

    Ok(to_json_binary(&TimestampAtHeightResult {
        timestamp: nano_timestamp,
    })?)
}

/// Gets the status of the light client
/// # Returns
/// Active status, because no other state is currently implemented
/// # Errors
/// It won't error at this point
// TODO: Implement a proper status once freezing/misbehaviour is implemented #164
pub fn status() -> Result<Binary, ContractError> {
    Ok(to_json_binary(&StatusResult {
        status: "Active".to_string(),
    })?)
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{
        coins, from_json,
        testing::{message_info, mock_env},
        Binary, Timestamp,
    };
    use ethereum_light_client::test_utils::fixtures::{
        self, InitialState, StepsFixture, UpdateClient,
    };

    use crate::{
        contract::{instantiate, query},
        msg::{
            Height, QueryMsg, StatusMsg, StatusResult, TimestampAtHeightMsg,
            TimestampAtHeightResult, VerifyClientMessageMsg,
        },
        query::timestamp_at_height,
        test::mk_deps,
    };

    use super::verify_client_message;

    #[test]
    fn test_verify_client_message() {
        let mut deps = mk_deps();
        let creator = deps.api.addr_make("creator");
        let info = message_info(&creator, &coins(1, "uatom"));

        let fixture: StepsFixture =
            fixtures::load("TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Groth16");

        let initial_state: InitialState = fixture.get_data_at_step(0);

        let client_state = initial_state.client_state;

        let consensus_state = initial_state.consensus_state;

        let client_state_bz: Vec<u8> = serde_json::to_vec(&client_state).unwrap();
        let consensus_state_bz: Vec<u8> = serde_json::to_vec(&consensus_state).unwrap();

        let msg = crate::msg::InstantiateMsg {
            client_state: Binary::from(client_state_bz),
            consensus_state: Binary::from(consensus_state_bz),
            checksum: b"checksum".into(),
        };

        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let update_client: UpdateClient = fixture.get_data_at_step(1);
        let header = update_client.updates[0].clone();
        let header_bz: Vec<u8> = serde_json::to_vec(&header).unwrap();

        let mut env = mock_env();
        env.block.time = Timestamp::from_seconds(
            header.consensus_update.attested_header.execution.timestamp + 1000,
        );

        verify_client_message(
            deps.as_ref(),
            env,
            VerifyClientMessageMsg {
                client_message: Binary::from(header_bz),
            },
        )
        .unwrap();
    }

    #[test]
    fn test_timestamp_at_height() {
        let mut deps = mk_deps();
        let creator = deps.api.addr_make("creator");
        let info = message_info(&creator, &coins(1, "uatom"));

        let fixture: StepsFixture =
            fixtures::load("TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Groth16");

        let initial_state: InitialState = fixture.get_data_at_step(0);

        let client_state = initial_state.client_state;
        let consensus_state = initial_state.consensus_state;

        let client_state_bz: Vec<u8> = serde_json::to_vec(&client_state).unwrap();
        let consensus_state_bz: Vec<u8> = serde_json::to_vec(&consensus_state).unwrap();

        let msg = crate::msg::InstantiateMsg {
            client_state: Binary::from(client_state_bz),
            consensus_state: Binary::from(consensus_state_bz),
            checksum: b"checksum".into(),
        };

        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let res = timestamp_at_height(
            deps.as_ref(),
            TimestampAtHeightMsg {
                height: Height {
                    revision_number: 0,
                    revision_height: consensus_state.slot,
                },
            },
        )
        .unwrap();
        let timestamp_at_height_result: TimestampAtHeightResult = from_json(&res).unwrap();
        assert_eq!(
            consensus_state.timestamp * 1_000_000_000, // ibc-go expects nanoseconds
            timestamp_at_height_result.timestamp
        );
    }

    #[test]
    fn test_status() {
        let deps = mk_deps();
        let res = query(deps.as_ref(), mock_env(), QueryMsg::Status(StatusMsg {})).unwrap();
        let status_response: StatusResult = from_json(&res).unwrap();
        assert_eq!("Active", status_response.status);
    }
}
