//! This module defines [`TxBuilder`] which is responsible for building transactions to be sent to
//! the Cosmos SDK chain from events received from another Cosmos SDK chain.

use anyhow::Result;
use ibc_proto_eureka::{
    cosmos::tx::v1beta1::TxBody,
    google::protobuf::Any,
    ibc::{
        core::client::v1::{Height, MsgUpdateClient},
        lightclients::tendermint::v1::ClientState,
    },
};
use prost::Message;
use sp1_ics07_tendermint_utils::{light_block::LightBlockExt, rpc::TendermintRpcExt};
use tendermint_rpc::HttpClient;

use crate::{
    chain::CosmosSdk,
    events::EurekaEvent,
    utils::cosmos::{self},
};

use super::r#trait::TxBuilderService;

/// The `TxBuilder` produces txs to [`CosmosSdk`] based on events from [`CosmosSdk`].
#[allow(dead_code)]
pub struct TxBuilder {
    /// The HTTP client for the source chain.
    pub source_tm_client: HttpClient,
    /// The HTTP client for the target chain.
    pub target_tm_client: HttpClient,
    /// The signer address for the Cosmos messages.
    pub signer_address: String,
}

impl TxBuilder {
    /// Creates a new `TxBuilder`.
    #[must_use]
    pub const fn new(
        source_tm_client: HttpClient,
        target_tm_client: HttpClient,
        signer_address: String,
    ) -> Self {
        Self {
            source_tm_client,
            target_tm_client,
            signer_address,
        }
    }
}

#[async_trait::async_trait]
impl TxBuilderService<CosmosSdk, CosmosSdk> for TxBuilder {
    #[tracing::instrument(skip_all)]
    async fn relay_events(
        &self,
        src_events: Vec<EurekaEvent>,
        target_events: Vec<EurekaEvent>,
        target_client_id: String,
    ) -> Result<Vec<u8>> {
        let client_state = ClientState::decode(
            self.target_tm_client
                .client_state(target_client_id.clone())
                .await?
                .value
                .as_slice(),
        )?;

        let target_light_block = self.source_tm_client.get_light_block(None).await?;
        let revision_height = target_light_block.height().value();
        let revision_number = client_state
            .latest_height
            .ok_or_else(|| anyhow::anyhow!("No latest height found"))?
            .revision_number;

        let target_height = Height {
            revision_number,
            revision_height,
        };

        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)?
            .as_secs();

        let mut timeout_msgs = cosmos::target_events_to_timeout_msgs(
            target_events,
            &target_client_id,
            &target_height,
            &self.signer_address,
            now,
        );

        let (mut recv_msgs, mut ack_msgs) = cosmos::src_events_to_recv_and_ack_msgs(
            src_events,
            &target_client_id,
            &target_height,
            &self.signer_address,
            now,
        );

        cosmos::inject_tendermint_proofs(
            &mut recv_msgs,
            &mut ack_msgs,
            &mut timeout_msgs,
            &self.source_tm_client,
            &target_height,
        )
        .await?;

        let trusted_light_block = self
            .source_tm_client
            .get_light_block(Some(
                client_state
                    .latest_height
                    .ok_or_else(|| anyhow::anyhow!("No latest height found"))?
                    .revision_height
                    .try_into()?,
            ))
            .await?;
        let proposed_header = target_light_block.into_header(&trusted_light_block);
        let update_msg = MsgUpdateClient {
            client_id: target_client_id,
            client_message: Some(Any::from_msg(&proposed_header)?),
            signer: self.signer_address.clone(),
        };

        let all_msgs = std::iter::once(Any::from_msg(&update_msg))
            .chain(timeout_msgs.into_iter().map(|m| Any::from_msg(&m)))
            .chain(recv_msgs.into_iter().map(|m| Any::from_msg(&m)))
            .chain(ack_msgs.into_iter().map(|m| Any::from_msg(&m)))
            .collect::<Result<Vec<_>, _>>()?;
        if all_msgs.len() == 1 {
            // The update message is the only message.
            anyhow::bail!("No messages to relay to Cosmos");
        }

        tracing::debug!(
            "Messages to be relayed to Cosmos: {:?}",
            all_msgs[1..].to_vec()
        );

        let tx_body = TxBody {
            messages: all_msgs,
            ..Default::default()
        };
        Ok(tx_body.encode_to_vec())
    }
}
