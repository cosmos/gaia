use crate::chain::Chain;

use anyhow::Result;

/// The `TxBuilderService` trait defines the interface for a service that submits transactions
/// to a chain based on events from two chains.
#[async_trait::async_trait]
pub trait TxBuilderService<A: Chain, B: Chain> {
    /// Generate a transaction to chain A based on the events from chain A and chain B.
    /// Events from chain A are often used for timeout purposes and can be left empty.
    ///
    /// # Returns
    /// The relay transaction bytes.
    async fn relay_events(
        &self,
        src_events: Vec<A::Event>,
        target_events: Vec<B::Event>,
        target_client_id: String,
    ) -> Result<Vec<u8>>;
}
