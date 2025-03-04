//! This module defines the chain listener for 'ibc-go-eureka'.

use futures::future;
use tendermint::Hash;
use tendermint_rpc::{Client, HttpClient};

use anyhow::Result;

use crate::{chain::CosmosSdk, events::EurekaEvent};

use super::ChainListenerService;

/// The `ChainListener` listens for events on the Cosmos SDK chain.
pub struct ChainListener(HttpClient);

impl ChainListener {
    /// Create a new [`Self`] instance.
    #[must_use]
    pub const fn new(tm_client: HttpClient) -> Self {
        Self(tm_client)
    }

    /// Get the HTTP client for tendermint.
    #[must_use]
    pub const fn client(&self) -> &HttpClient {
        &self.0
    }

    /// Get the chain ID.
    ///
    /// # Errors
    /// Returns an error if the chain ID cannot be fetched.
    pub async fn chain_id(&self) -> Result<String> {
        Ok(self
            .client()
            .latest_block()
            .await?
            .block
            .header
            .chain_id
            .into())
    }
}

#[async_trait::async_trait]
impl ChainListenerService<CosmosSdk> for ChainListener {
    async fn fetch_tx_events(&self, tx_ids: Vec<Hash>) -> Result<Vec<EurekaEvent>> {
        Ok(
            future::try_join_all(tx_ids.into_iter().map(|tx_id| async move {
                Ok::<_, tendermint_rpc::Error>(
                    self.client()
                        .tx(tx_id, false)
                        .await?
                        .tx_result
                        .events
                        .into_iter()
                        .filter_map(|e| EurekaEvent::try_from(e).ok()),
                )
            }))
            .await?
            .into_iter()
            .flatten()
            .collect(),
        )
    }

    async fn fetch_events(&self, start_height: u32, end_height: u32) -> Result<Vec<EurekaEvent>> {
        Ok(
            future::try_join_all((start_height..=end_height).map(|h| async move {
                let resp = self.client().block_results(h).await?;
                Ok::<_, tendermint_rpc::Error>(
                    resp.txs_results
                        .unwrap_or_default()
                        .into_iter()
                        .flat_map(|tx| tx.events)
                        .chain(resp.begin_block_events.unwrap_or_default())
                        .chain(resp.end_block_events.unwrap_or_default())
                        .chain(resp.finalize_block_events)
                        .filter_map(|e| EurekaEvent::try_from(e).ok()),
                )
            }))
            .await?
            .into_iter()
            .flatten()
            .collect(),
        )
    }
}
