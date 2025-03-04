//! This module contains the sudo message handlers

use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut};
use ethereum_light_client::update::update_consensus_state;
use ibc_proto::ibc::{
    core::client::v1::Height as IbcProtoHeight,
    lightclients::wasm::v1::ConsensusState as WasmConsensusState,
};

use crate::{
    custom_query::EthereumCustomQuery,
    msg::{
        Height, UpdateStateMsg, UpdateStateOnMisbehaviourMsg, UpdateStateResult,
        VerifyMembershipMsg, VerifyNonMembershipMsg,
    },
    state::{
        get_eth_client_state, get_eth_consensus_state, get_wasm_client_state, store_client_state,
        store_consensus_state,
    },
    ContractError,
};

/// Verify the membership of a value at a given height
/// # Errors
/// Returns an error if the membership proof verification fails
/// # Returns
/// An empty response
pub fn verify_membership(
    deps: Deps<EthereumCustomQuery>,
    verify_membership_msg: VerifyMembershipMsg,
) -> Result<Binary, ContractError> {
    let eth_client_state = get_eth_client_state(deps.storage)?;
    let eth_consensus_state =
        get_eth_consensus_state(deps.storage, verify_membership_msg.height.revision_height)?;

    ethereum_light_client::membership::verify_membership(
        eth_consensus_state,
        eth_client_state,
        verify_membership_msg.proof.into(),
        verify_membership_msg
            .merkle_path
            .key_path
            .into_iter()
            .map(Into::into)
            .collect(),
        Some(verify_membership_msg.value.into()),
    )
    .map_err(ContractError::VerifyMembershipFailed)?;

    Ok(Binary::default())
}

/// Verify the non-membership of a value at a given height
/// # Errors
/// Returns an error if the non-membership proof verification fails
/// # Returns
/// An empty response
pub fn verify_non_membership(
    deps: Deps<EthereumCustomQuery>,
    verify_non_membership_msg: VerifyNonMembershipMsg,
) -> Result<Binary, ContractError> {
    let eth_client_state = get_eth_client_state(deps.storage)?;
    let eth_consensus_state = get_eth_consensus_state(
        deps.storage,
        verify_non_membership_msg.height.revision_height,
    )?;

    ethereum_light_client::membership::verify_membership(
        eth_consensus_state,
        eth_client_state,
        verify_non_membership_msg.proof.into(),
        verify_non_membership_msg
            .merkle_path
            .key_path
            .into_iter()
            .map(Into::into)
            .collect(),
        None,
    )
    .map_err(ContractError::VerifyNonMembershipFailed)?;

    Ok(Binary::default())
}

/// Update the state of the light client
/// This function is always called after the verify client message, so
/// we can assume the client message is valid and that the consensus state can be updated
/// # Errors
/// Returns an error if deserialization failes or if the light client update logic fails
/// # Returns
/// The updated slot (called height in regular IBC terms)
#[allow(clippy::needless_pass_by_value)]
pub fn update_state(
    deps: DepsMut<EthereumCustomQuery>,
    update_state_msg: UpdateStateMsg,
) -> Result<Binary, ContractError> {
    let header_bz: Vec<u8> = update_state_msg.client_message.into();
    let header = serde_json::from_slice(&header_bz)
        .map_err(ContractError::DeserializeClientMessageFailed)?;

    let eth_client_state = get_eth_client_state(deps.storage)?;
    let eth_consensus_state = get_eth_consensus_state(deps.storage, eth_client_state.latest_slot)?;

    let (updated_slot, updated_consensus_state, updated_client_state) =
        update_consensus_state(eth_consensus_state, eth_client_state, header)
            .map_err(ContractError::UpdateClientStateFailed)?;

    let consensus_state_bz: Vec<u8> = serde_json::to_vec(&updated_consensus_state)
        .map_err(ContractError::SerializeConsensusStateFailed)?;
    let wasm_consensus_state = WasmConsensusState {
        data: consensus_state_bz,
    };
    store_consensus_state(deps.storage, &wasm_consensus_state, updated_slot)?;

    if let Some(client_state) = updated_client_state {
        let client_state_bz: Vec<u8> =
            serde_json::to_vec(&client_state).map_err(ContractError::SerializeClientStateFailed)?;

        let mut wasm_client_state = get_wasm_client_state(deps.storage)?;
        wasm_client_state.data = client_state_bz;
        wasm_client_state.latest_height = Some(IbcProtoHeight {
            revision_number: 0,
            revision_height: updated_slot,
        });
        store_client_state(deps.storage, &wasm_client_state)?;
    }

    Ok(to_json_binary(&UpdateStateResult {
        heights: vec![Height {
            revision_number: 0,
            revision_height: updated_slot,
        }],
    })?)
}

/// Update the state of the light client on misbehaviour
/// # Errors
/// Returns an error if the misbehaviour verification fails
#[allow(clippy::needless_pass_by_value)]
pub fn misbehaviour(
    deps: DepsMut<EthereumCustomQuery>,
    _msg: UpdateStateOnMisbehaviourMsg,
) -> Result<Binary, ContractError> {
    let mut eth_client_state = get_eth_client_state(deps.storage)?;
    eth_client_state.is_frozen = true;

    let client_state_bz: Vec<u8> =
        serde_json::to_vec(&eth_client_state).map_err(ContractError::SerializeClientStateFailed)?;

    let mut wasm_client_state = get_wasm_client_state(deps.storage)?;
    wasm_client_state.data = client_state_bz;

    store_client_state(deps.storage, &wasm_client_state)?;

    Ok(Binary::default())
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{
        coins, from_json,
        testing::{message_info, mock_env},
        Binary,
    };
    use ethereum_light_client::test_utils::fixtures::{
        self, CommitmentProof, InitialState, StepsFixture, UpdateClient,
    };

    use crate::{
        contract::instantiate,
        msg::{Height, MerklePath, UpdateStateMsg, UpdateStateResult, VerifyMembershipMsg},
        sudo::{update_state, verify_membership},
        test::mk_deps,
    };

    #[test]
    fn test_verify_membership() {
        let mut deps = mk_deps();
        let creator = deps.api.addr_make("creator");
        let info = message_info(&creator, &coins(1, "uatom"));

        let fixture: StepsFixture =
            fixtures::load("TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Groth16");

        let commitment_proof_fixture: CommitmentProof = fixture.get_data_at_step(2);

        let client_state = commitment_proof_fixture.client_state;
        let client_state_bz: Vec<u8> = serde_json::to_vec(&client_state).unwrap();
        let consensus_state = commitment_proof_fixture.consensus_state;
        let consensus_state_bz: Vec<u8> = serde_json::to_vec(&consensus_state).unwrap();

        let msg = crate::msg::InstantiateMsg {
            client_state: Binary::from(client_state_bz),
            consensus_state: Binary::from(consensus_state_bz),
            checksum: b"checksum".into(),
        };
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let proof = commitment_proof_fixture.storage_proof;
        let proof_bz = serde_json::to_vec(&proof).unwrap();
        let path = commitment_proof_fixture.path;
        let value = proof.value;
        let value_bz = value.to_be_bytes_vec();

        let msg = VerifyMembershipMsg {
            height: Height {
                revision_number: 0,
                revision_height: commitment_proof_fixture.proof_slot,
            },
            delay_time_period: 0,
            delay_block_period: 0,
            proof: Binary::from(proof_bz),
            merkle_path: MerklePath {
                key_path: vec![Binary::from(path.to_vec())],
            },
            value: Binary::from(value_bz),
        };
        verify_membership(deps.as_ref(), msg).unwrap();
    }

    #[test]
    fn test_update_state() {
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

        let msg = UpdateStateMsg {
            client_message: Binary::from(header_bz),
        };
        let res = update_state(deps.as_mut(), msg).unwrap();
        let update_state_result: UpdateStateResult = from_json(res).unwrap();
        assert_eq!(1, update_state_result.heights.len());
        assert_eq!(
            header.consensus_update.attested_header.beacon.slot,
            update_state_result.heights[0].revision_height
        );
    }
}
