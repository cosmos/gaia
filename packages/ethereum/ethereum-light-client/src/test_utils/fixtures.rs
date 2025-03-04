//! Test fixtures types and ulitiies for the Ethereum light client

use std::path::PathBuf;

use alloy_primitives::Bytes;
use ethereum_types::execution::storage_proof::StorageProof;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::{client_state::ClientState, consensus_state::ConsensusState, header::Header};

/// A test fixture with an ordered list of light client operations from the e2e test
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct StepsFixture {
    /// steps is a list of light client operations
    pub steps: Vec<Step>,
}

/// Step is a light client operation such as an initial state, commitment proof, or update client
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct Step {
    /// name is the name of the operation, only used for documentation and easy of reading
    pub name: String,
    /// data is the operation data as a JSON object to be deserialized into the appropriate type
    pub data: Value,
}

/// The initial state of the light client in the e2e tests
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct InitialState {
    /// The client state at the initial state
    pub client_state: ClientState,
    /// The consensus state at the initial state
    pub consensus_state: ConsensusState,
}

/// The proof used to verify membership
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct CommitmentProof {
    /// The IBC path sent to verify membership
    #[schemars(with = "String")]
    pub path: Bytes,
    /// The storage proof used to verify membership
    pub storage_proof: StorageProof,
    /// The slot of the proof (ibc height)
    pub proof_slot: u64,
    /// The client state at the time of the proof
    pub client_state: ClientState,
    /// The consensus state at the time of the proof
    pub consensus_state: ConsensusState,
}

/// Operation to update the light client
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug)]
pub struct UpdateClient {
    /// The client state after the update
    pub client_state: ClientState,
    /// The consensus state after the update
    pub consensus_state: ConsensusState,
    /// The headers used to update the light client, in order
    pub updates: Vec<Header>,
}

impl StepsFixture {
    /// Deserializes the data at the given step into the given type
    /// # Panics
    /// Panics if the data cannot be deserialized into the given type
    #[must_use]
    pub fn get_data_at_step<T>(&self, step: usize) -> T
    where
        T: serde::de::DeserializeOwned,
    {
        serde_json::from_value(self.steps[step].data.clone()).unwrap()
    }
}

/// load loads a test fixture from a JSON file
/// # Panics
/// Panics if the file cannot be opened or the contents cannot be deserialized
#[must_use]
pub fn load<T>(name: &str) -> T
where
    T: serde::de::DeserializeOwned,
{
    // Construct the path relative to the Cargo manifest directory
    let mut path = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    path.push("src/test_utils/fixtures");
    path.push(format!("{name}.json"));

    // Open the file and deserialize its contents
    let file = std::fs::File::open(path).unwrap();
    serde_json::from_reader(file).unwrap()
}
