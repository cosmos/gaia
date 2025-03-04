//! This module defines the chain listener for 'solidity-ibc-eureka' contracts.

use alloy::{
    primitives::{Address, TxHash},
    providers::Provider,
    rpc::types::Filter,
    sol_types::SolEventInterface,
};
use anyhow::{anyhow, Result};
use futures::future;
use ibc_eureka_solidity_types::ics26::router::{routerEvents, routerInstance};

use crate::{chain::EthEureka, events::EurekaEvent};

use super::ChainListenerService;

/// The `ChainListenerService` listens for events on the Ethereum chain.
pub struct ChainListener<P: Provider> {
    /// The IBC Eureka router instance.
    ics26_router: routerInstance<(), P>,
}

impl<P: Provider> ChainListener<P> {
    /// Create a new `ChainListenerService` instance.
    pub const fn new(ics26_address: Address, provider: P) -> Self {
        Self {
            ics26_router: routerInstance::new(ics26_address, provider),
        }
    }
}

impl<P> ChainListener<P>
where
    P: Provider,
{
    /// Get the chain ID.
    /// # Errors
    /// Returns an error if the chain ID cannot be fetched.
    pub async fn chain_id(&self) -> Result<String> {
        Ok(self
            .ics26_router
            .provider()
            .get_chain_id()
            .await?
            .to_string())
    }
}

#[async_trait::async_trait]
impl<P> ChainListenerService<EthEureka> for ChainListener<P>
where
    P: Provider,
{
    async fn fetch_tx_events(&self, tx_ids: Vec<TxHash>) -> Result<Vec<EurekaEvent>> {
        Ok(
            future::try_join_all(tx_ids.into_iter().map(|tx_id| async move {
                let block_hash = self
                    .ics26_router
                    .provider()
                    .get_transaction_by_hash(tx_id)
                    .await?
                    .ok_or_else(|| anyhow!("Transaction {} not found", tx_id))?
                    .block_hash
                    .ok_or_else(|| anyhow!("Transaction {} has not been mined", tx_id))?;

                let event_filter = Filter::new()
                    .events(EurekaEvent::evm_signatures())
                    .address(*self.ics26_router.address())
                    .at_block_hash(block_hash);

                Ok::<_, anyhow::Error>(
                    self.ics26_router
                        .provider()
                        .get_logs(&event_filter)
                        .await?
                        .iter()
                        .filter(|log| log.transaction_hash.unwrap_or_default() == tx_id)
                        .filter_map(|log| {
                            let sol_event = routerEvents::decode_log(&log.inner, true).ok()?.data;
                            EurekaEvent::try_from(sol_event).ok()
                        })
                        .collect::<Vec<_>>(),
                )
            }))
            .await?
            .into_iter()
            .flatten()
            .collect(),
        )
    }

    async fn fetch_events(&self, start_height: u64, end_height: u64) -> Result<Vec<EurekaEvent>> {
        let event_filter = Filter::new()
            .events(EurekaEvent::evm_signatures())
            .address(*self.ics26_router.address())
            .from_block(start_height)
            .to_block(end_height);

        Ok(self
            .ics26_router
            .provider()
            .get_logs(&event_filter)
            .await?
            .iter()
            .filter_map(|log| {
                let sol_event = routerEvents::decode_log(&log.inner, true).ok()?.data;
                EurekaEvent::try_from(sol_event).ok()
            })
            .collect())
    }
}
