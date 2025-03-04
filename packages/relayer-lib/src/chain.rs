//! Defines the [`Chain`] interface and some of its implementations.

use alloy::primitives::TxHash;
use serde::{de::DeserializeOwned, Serialize};
use std::fmt::Debug;

use crate::events::EurekaEvent;

/// The `Chain` trait defines the interface for a chain.
pub trait Chain {
    /// The event type that the listener will return.
    /// These should be the events that the relayer is interested in.
    type Event: Clone + Debug;
    /// The transaction identifier type that the listener will ask for.
    /// This is often a hash of the transaction.
    type TxId: Clone + Serialize + DeserializeOwned + Debug;
    /// The block height type that the listener will ask for.
    /// This is often a u64.
    type Height: Clone + Serialize + DeserializeOwned + Debug + std::cmp::PartialOrd;
}

/// The `CosmosSdk` is a concrete implementation of the `Chain` trait for the Cosmos SDK.
pub struct CosmosSdk;

impl Chain for CosmosSdk {
    type Event = EurekaEvent;
    type TxId = tendermint::Hash;
    type Height = u32;
}

/// The `EthEureka` is an implementation of the `Chain` trait for `solidity-ibc-eureka` contracts.
pub struct EthEureka;

impl Chain for EthEureka {
    type Event = EurekaEvent;
    type TxId = TxHash;
    type Height = u64;
}
