//! State management for the Ethereum light client

use cosmwasm_std::Storage;
use ethereum_light_client::client_state::ClientState as EthClientState;
use ethereum_light_client::consensus_state::ConsensusState as EthConsensusState;
use ibc_proto::{
    google::protobuf::Any,
    ibc::lightclients::wasm::v1::{
        ClientState as WasmClientState, ConsensusState as WasmConsensusState,
    },
};
use prost::Message;

use crate::ContractError;

/// The store key used by `ibc-go` to store the client state
pub const HOST_CLIENT_STATE_KEY: &str = "clientState";
/// The store key used by `ibc-go` to store the consensus states
pub const HOST_CONSENSUS_STATES_KEY: &str = "consensusStates";

/// The key used to store the consensus states by height
#[must_use]
pub fn consensus_db_key(slot: u64) -> String {
    format!("{}/{}-{}", HOST_CONSENSUS_STATES_KEY, 0, slot)
}

/// Get the Wasm client state
/// # Errors
/// Returns an error if the client state is not found or cannot be deserialized
/// # Returns
/// The Wasm client state
#[allow(clippy::module_name_repetitions)]
pub fn get_wasm_client_state(storage: &dyn Storage) -> Result<WasmClientState, ContractError> {
    let wasm_client_state_any_bz = storage
        .get(HOST_CLIENT_STATE_KEY.as_bytes())
        .ok_or(ContractError::ClientStateNotFound)?;
    let wasm_client_state_any = Any::decode(wasm_client_state_any_bz.as_slice())?;

    Ok(WasmClientState::decode(
        wasm_client_state_any.value.as_slice(),
    )?)
}

/// Get the Ethereum client state
/// # Errors
/// Returns an error if the client state is not found or cannot be deserialized
/// # Returns
/// The Ethereum client state
#[allow(clippy::module_name_repetitions)]
pub fn get_eth_client_state(storage: &dyn Storage) -> Result<EthClientState, ContractError> {
    let wasm_client_state = get_wasm_client_state(storage)?;
    Ok(serde_json::from_slice(&wasm_client_state.data)?)
}

/// Get the Ethereum consensus state at a given height
/// # Errors
/// Returns an error if the consensus state is not found or cannot be deserialized
/// # Returns
/// The Ethereum consensus state
#[allow(clippy::module_name_repetitions)]
pub fn get_eth_consensus_state(
    storage: &dyn Storage,
    slot: u64,
) -> Result<EthConsensusState, ContractError> {
    let wasm_consensus_state_any_bz = storage
        .get(consensus_db_key(slot).as_bytes())
        .ok_or(ContractError::ConsensusStateNotFound)?;
    let wasm_consensus_state_any = Any::decode(wasm_consensus_state_any_bz.as_slice())?;
    let wasm_consensus_state =
        WasmConsensusState::decode(wasm_consensus_state_any.value.as_slice())?;

    Ok(serde_json::from_slice(&wasm_consensus_state.data)?)
}

/// Store the consensus state
/// # Errors
/// Returns an error if the consensus state cannot be serialized into an Any
#[allow(clippy::module_name_repetitions)]
pub fn store_consensus_state(
    storage: &mut dyn Storage,
    wasm_consensus_state: &WasmConsensusState,
    slot: u64,
) -> Result<(), ContractError> {
    let wasm_consensus_state_any = Any::from_msg(wasm_consensus_state)?;
    storage.set(
        consensus_db_key(slot).as_bytes(),
        wasm_consensus_state_any.encode_to_vec().as_slice(),
    );

    Ok(())
}

/// Store the client state
/// # Errors
/// Returns an error if the client state cannot be serialized into an Any
#[allow(clippy::module_name_repetitions)]
pub fn store_client_state(
    storage: &mut dyn Storage,
    wasm_client_state: &WasmClientState,
) -> Result<(), ContractError> {
    let wasm_client_state_any = Any::from_msg(wasm_client_state)?;
    storage.set(
        HOST_CLIENT_STATE_KEY.as_bytes(),
        wasm_client_state_any.encode_to_vec().as_slice(),
    );

    Ok(())
}
