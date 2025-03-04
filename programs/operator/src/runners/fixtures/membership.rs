//! Runner for generating `membership` fixtures

use crate::{
    cli::command::{fixtures::MembershipCmd, OutputPath},
    runners::genesis::SP1ICS07TendermintGenesis,
};
use alloy::sol_types::SolValue;
use core::str;
use ibc_client_tendermint_types::ConsensusState;
use ibc_eureka_solidity_types::msgs::{
    IICS07TendermintMsgs::{
        ClientState, ConsensusState as SolConsensusState, SupportedZkAlgorithm,
    },
    IMembershipMsgs::{KVPair, MembershipOutput, MembershipProof, SP1MembershipProof},
    ISP1Msgs::SP1Proof,
};
use serde::{Deserialize, Serialize};
use sp1_ics07_tendermint_prover::{programs::MembershipProgram, prover::SP1ICS07TendermintProver};
use sp1_ics07_tendermint_utils::rpc::TendermintRpcExt;
use sp1_sdk::{HashableKey, ProverClient};
use std::path::PathBuf;
use tendermint_rpc::HttpClient;

/// The fixture data to be used in [`MembershipProgram`] tests.
#[serde_with::serde_as]
#[derive(Debug, Clone, Deserialize, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SP1ICS07MembershipFixture {
    /// The genesis data.
    #[serde(flatten)]
    pub genesis: SP1ICS07TendermintGenesis,
    /// The height of the proof.
    #[serde_as(as = "serde_with::hex::Hex")]
    pub proof_height: Vec<u8>,
    /// The encoded public values.
    #[serde_as(as = "serde_with::hex::Hex")]
    pub membership_proof: Vec<u8>,
}

/// Writes the proof data for the given trusted and target blocks to the given fixture path.
#[allow(clippy::missing_errors_doc, clippy::missing_panics_doc)]
pub async fn run(args: MembershipCmd) -> anyhow::Result<()> {
    assert!(!args.membership.key_paths.is_empty());

    let tm_rpc_client = HttpClient::from_env();

    let trusted_light_block = tm_rpc_client
        .get_light_block(Some(args.membership.trusted_block))
        .await?;

    let genesis = SP1ICS07TendermintGenesis::from_env(
        &trusted_light_block,
        args.membership.trust_options.trusting_period,
        args.membership.trust_options.trust_level,
        args.proof_type,
    )
    .await?;

    let trusted_client_state = ClientState::abi_decode(&genesis.trusted_client_state, false)?;
    let trusted_consensus_state =
        SolConsensusState::abi_decode(&genesis.trusted_consensus_state, false)?;

    let membership_proof = run_sp1_membership(
        &tm_rpc_client,
        args.membership.base64,
        args.membership.key_paths,
        args.membership.trusted_block,
        trusted_consensus_state,
        args.proof_type,
    )
    .await?;

    let fixture = SP1ICS07MembershipFixture {
        genesis,
        proof_height: trusted_client_state.latestHeight.abi_encode(),
        membership_proof: membership_proof.abi_encode(),
    };

    match args.membership.output_path {
        OutputPath::File(path) => {
            // Save the proof data to the file path.
            std::fs::write(PathBuf::from(path), serde_json::to_string_pretty(&fixture)?).unwrap();
        }
        OutputPath::Stdout => {
            println!("{}", serde_json::to_string_pretty(&fixture)?);
        }
    }

    Ok(())
}

/// Generates an sp1 membership proof for the given args
#[allow(
    clippy::missing_errors_doc,
    clippy::missing_panics_doc,
    clippy::module_name_repetitions
)]
pub async fn run_sp1_membership(
    tm_rpc_client: &HttpClient,
    is_base64: bool,
    key_paths: Vec<String>,
    trusted_block: u32,
    trusted_consensus_state: SolConsensusState,
    proof_type: SupportedZkAlgorithm,
) -> anyhow::Result<MembershipProof> {
    let sp1_prover = ProverClient::from_env();
    let verify_mem_prover =
        SP1ICS07TendermintProver::<MembershipProgram, _>::new(proof_type, &sp1_prover);

    let commitment_root_bytes = ConsensusState::from(trusted_consensus_state.clone())
        .root
        .as_bytes()
        .to_vec();

    let kv_proofs: Vec<(_, _)> =
        futures::future::try_join_all(key_paths.into_iter().map(|path| async {
            let path: Vec<Vec<u8>> = if is_base64 {
                path.split('\\')
                    .map(subtle_encoding::base64::decode)
                    .collect::<Result<_, _>>()?
            } else {
                vec![b"ibc".into(), path.into_bytes()]
            };
            assert_eq!(path.len(), 2);

            let (value, proof) = tm_rpc_client.prove_path(&path, trusted_block).await?;
            let kv_pair = KVPair {
                path: path.into_iter().map(Into::into).collect(),
                value: value.into(),
            };

            anyhow::Ok((kv_pair, proof))
        }))
        .await?;

    // Generate a header update proof for the specified blocks.
    let proof_data = verify_mem_prover.generate_proof(&commitment_root_bytes, kv_proofs);

    let bytes = proof_data.public_values.as_slice();
    let output = MembershipOutput::abi_decode(bytes, true)?;
    assert_eq!(output.commitmentRoot.as_slice(), &commitment_root_bytes);

    let sp1_membership_proof = SP1MembershipProof {
        sp1Proof: SP1Proof::new(
            &verify_mem_prover.vkey.bytes32(),
            proof_data.bytes(),
            proof_data.public_values.to_vec(),
        ),
        trustedConsensusState: trusted_consensus_state,
    };

    Ok(MembershipProof::from(sp1_membership_proof))
}
