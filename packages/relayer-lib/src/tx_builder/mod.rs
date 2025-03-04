//! This module defines the [`TxBuilderService`] trait and some of its implementations.
//! This interface is used to generate proofs and submit transactions to a chain.

pub mod cosmos_to_cosmos;
#[cfg(feature = "sp1-toolchain")]
pub mod cosmos_to_eth;
pub mod eth_to_cosmos;
mod r#trait;

#[allow(clippy::module_name_repetitions)]
pub use r#trait::TxBuilderService;
