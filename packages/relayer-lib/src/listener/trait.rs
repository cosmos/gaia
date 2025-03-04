//! Defines the [`ChainListenerService`] trait.

use anyhow::Result;

use crate::chain::Chain;

/// The `ChainListenerService` trait defines the interface for a service that listens to a chain
#[async_trait::async_trait]
pub trait ChainListenerService<C: Chain> {
    /// Fetch events from a transaction.
    async fn fetch_tx_events(&self, tx_ids: Vec<C::TxId>) -> Result<Vec<C::Event>>;

    /// Fetch events from a block range.
    /// Both the start and end heights are inclusive.
    async fn fetch_events(
        &self,
        start_height: C::Height,
        end_height: C::Height,
    ) -> Result<Vec<C::Event>>;
}
