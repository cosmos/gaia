//! RPC client for interacting with a Tendermint node.

use core::str::FromStr;
use std::{collections::HashMap, env};

use anyhow::Result;

use cosmos_sdk_proto::{
    cosmos::staking::v1beta1::{Params, QueryParamsRequest, QueryParamsResponse},
    prost::Message,
    traits::MessageExt,
    Any,
};
use ibc_core_client_types::proto::v1::{
    QueryClientStateRequest, QueryClientStateResponse, QueryConsensusStateRequest,
    QueryConsensusStateResponse,
};
use ibc_core_commitment_types::merkle::MerkleProof;
use tendermint::{block::signed_header::SignedHeader, validator::Set};
use tendermint_light_client_verifier::types::{LightBlock, ValidatorSet};
use tendermint_rpc::{Client, HttpClient, Paging, Url};

use crate::merkle::convert_tm_to_ics_merkle_proof;

/// An extension trait for [`HttpClient`] that provides additional methods for
/// obtaining light blocks.
#[async_trait::async_trait]
pub trait TendermintRpcExt {
    /// Creates a new instance of the Tendermint RPC client from the environment variables.
    ///
    /// # Panics
    /// Panics if the `TENDERMINT_RPC_URL` environment variable is not set or if the URL is
    /// invalid.
    #[must_use]
    fn from_env() -> Self;
    /// Gets a light block for a specific block height.
    /// If `block_height` is `None`, the latest block is fetched.
    ///
    /// # Errors
    /// Returns an error if the RPC request fails or if the response cannot be parsed.
    async fn get_light_block(&self, block_height: Option<u32>) -> Result<LightBlock>;
    /// Queries the Cosmos SDK for staking parameters.
    async fn sdk_staking_params(&self) -> Result<Params>;
    /// Fetches the client state from the Tendermint node.
    async fn client_state(&self, client_id: String) -> Result<Any>;
    /// Fetches the Ethereum consensus state from the light client on cosmos.
    /// If the revision height is 0, the latest height is fetched.
    async fn consensus_state(&self, client_id: String, revision_height: u64) -> Result<Any>;
    /// Proves a path in the chain's Merkle tree and returns the value at the path and the proof.
    /// If the value is empty, then this is a non-inclusion proof.
    async fn prove_path(&self, path: &[Vec<u8>], height: u32) -> Result<(Vec<u8>, MerkleProof)>;
}

#[async_trait::async_trait]
impl TendermintRpcExt for HttpClient {
    fn from_env() -> Self {
        Self::new(
            Url::from_str(&env::var("TENDERMINT_RPC_URL").expect("TENDERMINT_RPC_URL not set"))
                .expect("Failed to parse URL"),
        )
        .expect("Failed to create HTTP client")
    }

    async fn get_light_block(&self, block_height: Option<u32>) -> Result<LightBlock> {
        let peer_id = self.status().await?.node_info.id;
        let commit_response;
        let height;
        if let Some(block_height) = block_height {
            commit_response = self.commit(block_height).await?;
            height = block_height;
        } else {
            commit_response = self.latest_commit().await?;
            height = commit_response
                .signed_header
                .header
                .height
                .value()
                .try_into()?;
        }
        let mut signed_header = commit_response.signed_header;

        let validator_response = self.validators(height, Paging::All).await?;
        let validators = Set::with_proposer(
            validator_response.validators,
            signed_header.header().proposer_address,
        )?;

        let next_validator_response = self.validators(height + 1, Paging::All).await?;
        let next_validators = Set::with_proposer(
            next_validator_response.validators,
            // WARN: This proposer is likely to be incorrect,
            // but it is not used in the light block verification,
            // and required by ibc-go's validate basic.
            signed_header.header().proposer_address,
        )?;

        sort_signatures_by_validators_power_desc(&mut signed_header, &validators);
        Ok(LightBlock::new(
            signed_header,
            validators,
            next_validators,
            peer_id,
        ))
    }

    async fn sdk_staking_params(&self) -> Result<Params> {
        let abci_resp = self
            .abci_query(
                Some("/cosmos.staking.v1beta1.Query/Params".to_string()),
                QueryParamsRequest::default().to_bytes()?,
                None,
                false,
            )
            .await?;
        QueryParamsResponse::decode(abci_resp.value.as_slice())?
            .params
            .ok_or_else(|| anyhow::anyhow!("No staking params found"))
    }

    async fn client_state(&self, client_id: String) -> Result<Any> {
        let abci_resp = self
            .abci_query(
                Some("/ibc.core.client.v1.Query/ClientState".to_string()),
                QueryClientStateRequest { client_id }.to_bytes()?,
                None,
                false,
            )
            .await?;

        QueryClientStateResponse::decode(abci_resp.value.as_slice())?
            .client_state
            .ok_or_else(|| anyhow::anyhow!("No client state found"))
    }

    async fn consensus_state(&self, client_id: String, revision_height: u64) -> Result<Any> {
        let abci_resp = self
            .abci_query(
                Some("/ibc.core.client.v1.Query/ConsensusState".to_string()),
                QueryConsensusStateRequest {
                    client_id,
                    revision_number: 0,
                    revision_height,
                    latest_height: revision_height == 0,
                }
                .encode_to_vec(),
                None,
                false,
            )
            .await?;

        QueryConsensusStateResponse::decode(abci_resp.value.as_slice())?
            .consensus_state
            .ok_or_else(|| anyhow::anyhow!("No consensus state found"))
    }

    async fn prove_path(&self, path: &[Vec<u8>], height: u32) -> Result<(Vec<u8>, MerkleProof)> {
        let res = self
            .abci_query(
                Some(format!("store/{}/key", std::str::from_utf8(&path[0])?)),
                path[1..].concat(),
                // Proof height should be the block before the target block.
                Some((height - 1).into()),
                true,
            )
            .await?;

        if u32::try_from(res.height.value())? + 1 != height {
            anyhow::bail!("Proof height mismatch");
        }

        if res.key.as_slice() != path[1].as_slice() {
            anyhow::bail!("Key mismatch");
        }

        let vm_proof = convert_tm_to_ics_merkle_proof(
            &res.proof
                .ok_or_else(|| anyhow::anyhow!("Proof could not be retrieved"))?,
        )?;
        if vm_proof.proofs.is_empty() {
            anyhow::bail!("Empty proof");
        }

        anyhow::Ok((res.value, vm_proof))
    }
}

/// Sorts the signatures in the signed header based on the descending order of validators' power.
fn sort_signatures_by_validators_power_desc(
    signed_header: &mut SignedHeader,
    validators_set: &ValidatorSet,
) {
    let validator_powers: HashMap<_, _> = validators_set
        .validators()
        .iter()
        .map(|v| (v.address, v.power()))
        .collect();

    signed_header.commit.signatures.sort_by(|a, b| {
        let power_a = a
            .validator_address()
            .and_then(|addr| validator_powers.get(&addr))
            .unwrap_or(&0);
        let power_b = b
            .validator_address()
            .and_then(|addr| validator_powers.get(&addr))
            .unwrap_or(&0);
        power_b.cmp(power_a)
    });
}
