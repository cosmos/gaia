//! This module defines [`StorageProof`].

use alloy_primitives::{Bytes, B256, U256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// The key-value storage proof for a smart contract account
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
pub struct StorageProof {
    /// The key of the storage
    #[schemars(with = "String")]
    pub key: B256,
    /// The value of the storage
    #[schemars(with = "String")]
    pub value: U256,
    /// The proof of the storage
    #[schemars(with = "Vec<String>")]
    pub proof: Vec<Bytes>,
}
