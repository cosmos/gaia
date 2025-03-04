//! This module defines the conversion functions between Tendermint and ICS Merkle proofs.

use ibc_core_commitment_types::{merkle::MerkleProof, proto::ics23::CommitmentProof};
use tendermint::merkle::proof::ProofOps;

/// Convert a Tendermint proof to an ICS Merkle proof.
///
/// # Errors
/// Returns a decoding error if the prost merge.
pub fn convert_tm_to_ics_merkle_proof(
    tm_proof: &ProofOps,
) -> Result<MerkleProof, prost::DecodeError> {
    let mut proofs = Vec::with_capacity(tm_proof.ops.len());

    for op in &tm_proof.ops {
        let mut parsed = CommitmentProof { proof: None };
        prost::Message::merge(&mut parsed, op.data.as_slice())?;
        proofs.push(parsed);
    }

    Ok(MerkleProof { proofs })
}
