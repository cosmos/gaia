//! This module defines [`AccountProof`].

use alloy_primitives::{Bytes, B256};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// The account proof
#[derive(Serialize, Deserialize, JsonSchema, PartialEq, Eq, Clone, Debug, Default)]
pub struct AccountProof {
    /// The account storage root
    #[schemars(with = "String")]
    pub storage_root: B256,
    /// The account proof
    #[schemars(with = "Vec<String>")]
    pub proof: Vec<Bytes>,
}
