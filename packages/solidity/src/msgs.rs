//! This module contains all the message types used in Solidity IBC Eureka.
//! In case some message types are not found in the `ics26` module nor the `sp1_ics07` module,
//! they are defined here.

use super::sp1_ics07;
use alloy_sol_types::SolValue;
use ibc_client_tendermint_types::ConsensusState as ICS07TendermintConsensusState;
use ibc_core_commitment_types::{commitment::CommitmentRoot, merkle::MerklePath};
use tendermint::{hash::Algorithm, Time};
use tendermint_light_client_verifier::types::{Hash, TrustThreshold as TendermintTrustThreshold};
use time::OffsetDateTime;

alloy_sol_types::sol!("../../contracts/msgs/IICS26RouterMsgs.sol");
alloy_sol_types::sol!("../../contracts/msgs/IICS02ClientMsgs.sol");
alloy_sol_types::sol!("../../contracts/msgs/ILightClientMsgs.sol");
alloy_sol_types::sol!("../../contracts/msgs/IICS20TransferMsgs.sol");
alloy_sol_types::sol!("../../contracts/msgs/IIBCAppCallbacks.sol");

alloy_sol_types::sol!("../../contracts/light-clients/msgs/IICS07TendermintMsgs.sol");
alloy_sol_types::sol!("../../contracts/light-clients/msgs/ISP1Msgs.sol");
alloy_sol_types::sol!("../../contracts/light-clients/msgs/IMembershipMsgs.sol");
alloy_sol_types::sol!("../../contracts/light-clients/msgs/IMisbehaviourMsgs.sol");
alloy_sol_types::sol!("../../contracts/light-clients/msgs/IUpdateClientMsgs.sol");
alloy_sol_types::sol!("../../contracts/light-clients/msgs/IUcAndMembershipMsgs.sol");

#[cfg(feature = "rpc")]
impl ISP1Msgs::SP1Proof {
    /// Create a new [`SP1Proof`] instance.
    ///
    /// # Panics
    /// Panics if the vkey is not a valid hex string, or if the bytes cannot be decoded.
    #[must_use]
    pub fn new(vkey: &str, proof: Vec<u8>, public_values: Vec<u8>) -> Self {
        let stripped = vkey.strip_prefix("0x").expect("failed to strip prefix");
        let vkey_bytes: [u8; 32] = hex::decode(stripped)
            .expect("failed to decode vkey")
            .try_into()
            .expect("invalid vkey length");
        Self {
            vKey: vkey_bytes.into(),
            proof: proof.into(),
            publicValues: public_values.into(),
        }
    }
}

#[allow(clippy::fallible_impl_from)]
impl From<ICS07TendermintConsensusState> for IICS07TendermintMsgs::ConsensusState {
    fn from(ics07_tendermint_consensus_state: ICS07TendermintConsensusState) -> Self {
        let root: [u8; 32] = ics07_tendermint_consensus_state
            .root
            .into_vec()
            .try_into()
            .unwrap();
        let next_validators_hash: [u8; 32] = ics07_tendermint_consensus_state
            .next_validators_hash
            .as_bytes()
            .try_into()
            .unwrap();
        Self {
            #[allow(clippy::cast_possible_truncation, clippy::cast_sign_loss)]
            timestamp: ics07_tendermint_consensus_state.timestamp.unix_timestamp() as u64,
            root: root.into(),
            nextValidatorsHash: next_validators_hash.into(),
        }
    }
}

#[allow(clippy::fallible_impl_from)]
impl From<IICS07TendermintMsgs::ConsensusState> for ICS07TendermintConsensusState {
    fn from(consensus_state: IICS07TendermintMsgs::ConsensusState) -> Self {
        let time =
            OffsetDateTime::from_unix_timestamp(consensus_state.timestamp.try_into().unwrap())
                .unwrap();
        let seconds = time.unix_timestamp();
        let nanos = time.nanosecond();
        Self {
            timestamp: Time::from_unix_timestamp(seconds, nanos).unwrap(),
            root: CommitmentRoot::from_bytes(&consensus_state.root.0),
            next_validators_hash: Hash::from_bytes(
                Algorithm::Sha256,
                &consensus_state.nextValidatorsHash.0,
            )
            .unwrap(),
        }
    }
}

impl From<IMembershipMsgs::KVPair> for (MerklePath, Vec<u8>) {
    fn from(kv_pair: IMembershipMsgs::KVPair) -> Self {
        (
            MerklePath {
                key_path: kv_pair
                    .path
                    .into_iter()
                    .map(Vec::from)
                    .map(Into::into)
                    .collect(),
            },
            kv_pair.value.into(),
        )
    }
}

impl From<IMembershipMsgs::SP1MembershipProof> for IMembershipMsgs::MembershipProof {
    fn from(proof: IMembershipMsgs::SP1MembershipProof) -> Self {
        Self {
            proofType: IMembershipMsgs::MembershipProofType::SP1MembershipProof,
            proof: proof.abi_encode().into(),
        }
    }
}

impl From<IMembershipMsgs::SP1MembershipAndUpdateClientProof> for IMembershipMsgs::MembershipProof {
    fn from(proof: IMembershipMsgs::SP1MembershipAndUpdateClientProof) -> Self {
        Self {
            proofType: IMembershipMsgs::MembershipProofType::SP1MembershipAndUpdateClientProof,
            proof: proof.abi_encode().into(),
        }
    }
}

impl From<sp1_ics07::IICS07TendermintMsgs::TrustThreshold>
    for IICS07TendermintMsgs::TrustThreshold
{
    fn from(trust_threshold: sp1_ics07::IICS07TendermintMsgs::TrustThreshold) -> Self {
        Self {
            numerator: trust_threshold.numerator,
            denominator: trust_threshold.denominator,
        }
    }
}

impl From<sp1_ics07::IICS02ClientMsgs::Height> for IICS02ClientMsgs::Height {
    fn from(height: sp1_ics07::IICS02ClientMsgs::Height) -> Self {
        Self {
            revisionNumber: height.revisionNumber,
            revisionHeight: height.revisionHeight,
        }
    }
}

#[allow(clippy::fallible_impl_from)]
impl From<IICS07TendermintMsgs::TrustThreshold> for TendermintTrustThreshold {
    fn from(trust_threshold: IICS07TendermintMsgs::TrustThreshold) -> Self {
        Self::new(
            trust_threshold.numerator.into(),
            trust_threshold.denominator.into(),
        )
        .unwrap()
    }
}

impl TryFrom<TendermintTrustThreshold> for IICS07TendermintMsgs::TrustThreshold {
    type Error = <u64 as TryInto<u32>>::Error;

    fn try_from(trust_threshold: TendermintTrustThreshold) -> Result<Self, Self::Error> {
        Ok(Self {
            numerator: trust_threshold.numerator().try_into()?,
            denominator: trust_threshold.denominator().try_into()?,
        })
    }
}

impl TryFrom<ibc_core_client_types::Height> for IICS02ClientMsgs::Height {
    type Error = <u64 as TryInto<u32>>::Error;

    fn try_from(height: ibc_core_client_types::Height) -> Result<Self, Self::Error> {
        Ok(Self {
            revisionNumber: height.revision_number().try_into()?,
            revisionHeight: height.revision_height().try_into()?,
        })
    }
}

#[cfg(feature = "rpc")]
#[allow(clippy::fallible_impl_from)]
impl From<sp1_ics07::sp1_ics07_tendermint::clientStateReturn>
    for IICS07TendermintMsgs::ClientState
{
    fn from(client_state: sp1_ics07::sp1_ics07_tendermint::clientStateReturn) -> Self {
        Self {
            chainId: client_state.chainId,
            trustLevel: client_state.trustLevel.into(),
            trustingPeriod: client_state.trustingPeriod,
            unbondingPeriod: client_state.unbondingPeriod,
            latestHeight: client_state.latestHeight.into(),
            isFrozen: client_state.isFrozen,
            zkAlgorithm: IICS07TendermintMsgs::SupportedZkAlgorithm::try_from(
                client_state.zkAlgorithm,
            )
            .unwrap(),
        }
    }
}
