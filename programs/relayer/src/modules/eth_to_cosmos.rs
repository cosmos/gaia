//! Defines Ethereum to Cosmos relayer module.

use std::str::FromStr;

use alloy::{
    primitives::{Address, TxHash},
    providers::{Provider, RootProvider},
};
use ibc_eureka_relayer_lib::{
    events::EurekaEvent,
    listener::{cosmos_sdk, eth_eureka, ChainListenerService},
    tx_builder::{eth_to_cosmos, TxBuilderService},
};
use tendermint::Hash;
use tendermint_rpc::{HttpClient, Url};
use tonic::{Request, Response};

use crate::{
    api::{self, relayer_service_server::RelayerService},
    core::modules::RelayerModule,
};

/// The `CosmosToCosmosRelayerModule` struct defines the Cosmos to Cosmos relayer module.
#[derive(Clone, Copy, Debug)]
#[allow(clippy::module_name_repetitions)]
pub struct EthToCosmosRelayerModule;

/// The `CosmosToCosmosRelayerModuleService` defines the relayer service from Cosmos to Cosmos.
struct EthToCosmosRelayerModuleService {
    /// The chain listener for `EthEureka`.
    pub eth_listener: eth_eureka::ChainListener<RootProvider>,
    /// The chain listener for Cosmos SDK.
    pub tm_listener: cosmos_sdk::ChainListener,
    /// The transaction builder for Ethereum to Cosmos.
    pub tx_builder: EthToCosmosTxBuilder,
}

enum EthToCosmosTxBuilder {
    Real(eth_to_cosmos::TxBuilder<RootProvider>),
    Mock(eth_to_cosmos::MockTxBuilder<RootProvider>),
}

/// The configuration for the Cosmos to Cosmos relayer module.
#[derive(Clone, Debug, serde::Deserialize, serde::Serialize)]
#[allow(clippy::module_name_repetitions)]
pub struct EthToCosmosConfig {
    /// The ICS26 address.
    pub ics26_address: Address,
    /// The tendermint RPC URL.
    pub tm_rpc_url: String,
    /// The EVM RPC URL.
    pub eth_rpc_url: String,
    /// The Ethereum Beacon API URL
    pub eth_beacon_api_url: String,
    /// The address of the submitter.
    /// Required since cosmos messages require a signer address.
    pub signer_address: String,
    /// Whether to run in mock mode.
    #[serde(default)]
    pub mock: bool,
}

impl EthToCosmosRelayerModuleService {
    async fn new(config: EthToCosmosConfig) -> Self {
        let provider = RootProvider::builder()
            .on_builtin(&config.eth_rpc_url)
            .await
            .unwrap_or_else(|e| panic!("failed to create provider: {e}"));
        let eth_listener = eth_eureka::ChainListener::new(config.ics26_address, provider.clone());

        let tm_client = HttpClient::new(
            Url::from_str(&config.tm_rpc_url)
                .unwrap_or_else(|_| panic!("invalid tendermint RPC URL: {}", config.tm_rpc_url)),
        )
        .expect("Failed to create tendermint HTTP client");
        let tm_listener = cosmos_sdk::ChainListener::new(tm_client.clone());

        let tx_builder = if config.mock {
            EthToCosmosTxBuilder::Mock(eth_to_cosmos::MockTxBuilder::new(
                config.ics26_address,
                provider,
                config.signer_address,
            ))
        } else {
            EthToCosmosTxBuilder::Real(eth_to_cosmos::TxBuilder::new(
                config.ics26_address,
                provider,
                config.eth_beacon_api_url,
                tm_client,
                config.signer_address,
            ))
        };

        Self {
            eth_listener,
            tm_listener,
            tx_builder,
        }
    }
}

#[tonic::async_trait]
impl RelayerService for EthToCosmosRelayerModuleService {
    #[tracing::instrument(skip_all)]
    async fn info(
        &self,
        _request: Request<api::InfoRequest>,
    ) -> Result<Response<api::InfoResponse>, tonic::Status> {
        tracing::info!("Received info request.");
        Ok(Response::new(api::InfoResponse {
            target_chain: Some(api::Chain {
                chain_id: self
                    .tm_listener
                    .chain_id()
                    .await
                    .map_err(|e| tonic::Status::from_error(e.to_string().into()))?,
                ibc_version: "2".to_string(),
                ibc_contract: String::new(),
            }),
            source_chain: Some(api::Chain {
                chain_id: self
                    .eth_listener
                    .chain_id()
                    .await
                    .map_err(|e| tonic::Status::from_error(e.to_string().into()))?,
                ibc_version: "2".to_string(),
                ibc_contract: self.tx_builder.ics26_router_address().to_string(),
            }),
        }))
    }

    #[tracing::instrument(skip_all)]
    async fn relay_by_tx(
        &self,
        request: Request<api::RelayByTxRequest>,
    ) -> Result<Response<api::RelayByTxResponse>, tonic::Status> {
        tracing::info!("Handling relay by tx request for eth to cosmos...");

        let inner_req = request.into_inner();
        tracing::info!("Got {} source tx IDs", inner_req.source_tx_ids.len());
        tracing::info!("Got {} timeout tx IDs", inner_req.timeout_tx_ids.len());
        let eth_txs = inner_req
            .source_tx_ids
            .into_iter()
            .map(TryInto::<[u8; 32]>::try_into)
            .map(|tx_hash| tx_hash.map(TxHash::from))
            .collect::<Result<Vec<_>, _>>()
            .map_err(|tx| tonic::Status::from_error(format!("invalid tx hash: {tx:?}").into()))?;

        let cosmos_txs = inner_req
            .timeout_tx_ids
            .into_iter()
            .map(Hash::try_from)
            .collect::<Result<Vec<_>, _>>()
            .map_err(|e| tonic::Status::from_error(e.to_string().into()))?;

        let eth_events = self
            .eth_listener
            .fetch_tx_events(eth_txs)
            .await
            .map_err(|e| tonic::Status::from_error(e.to_string().into()))?;

        tracing::debug!(eth_events = ?eth_events, "Fetched EVM events.");
        tracing::info!("Fetched {} eureka events from EVM.", eth_events.len());

        let cosmos_events = self
            .tm_listener
            .fetch_tx_events(cosmos_txs)
            .await
            .map_err(|e| tonic::Status::from_error(e.to_string().into()))?;

        tracing::debug!(cosmos_events = ?cosmos_events, "Fetched cosmos events.");
        tracing::info!(
            "Fetched {} eureka events from CosmosSDK.",
            cosmos_events.len()
        );

        let tx = self
            .tx_builder
            .relay_events(eth_events, cosmos_events, inner_req.target_client_id)
            .await
            .map_err(|e| tonic::Status::from_error(e.to_string().into()))?;

        tracing::info!("Relay by tx request completed.");

        Ok(Response::new(api::RelayByTxResponse {
            tx,
            address: String::new(),
        }))
    }
}

#[tonic::async_trait]
impl RelayerModule for EthToCosmosRelayerModule {
    fn name(&self) -> &'static str {
        "eth_to_cosmos"
    }

    #[tracing::instrument(skip_all)]
    async fn create_service(
        &self,
        config: serde_json::Value,
    ) -> anyhow::Result<Box<dyn RelayerService>> {
        let config = serde_json::from_value::<EthToCosmosConfig>(config)
            .map_err(|e| anyhow::anyhow!("failed to parse config: {e}"))?;

        tracing::info!("Starting Ethereum to Cosmos relayer server.");
        Ok(Box::new(EthToCosmosRelayerModuleService::new(config).await))
    }
}

impl EthToCosmosTxBuilder {
    async fn relay_events(
        &self,
        src_events: Vec<EurekaEvent>,
        target_events: Vec<EurekaEvent>,
        target_client_id: String,
    ) -> anyhow::Result<Vec<u8>> {
        match self {
            Self::Real(tb) => {
                tb.relay_events(src_events, target_events, target_client_id)
                    .await
            }
            Self::Mock(tb) => {
                tb.relay_events(src_events, target_events, target_client_id)
                    .await
            }
        }
    }

    const fn ics26_router_address(&self) -> &Address {
        match self {
            Self::Real(tb) => tb.ics26_router.address(),
            Self::Mock(tb) => tb.ics26_router.address(),
        }
    }
}
