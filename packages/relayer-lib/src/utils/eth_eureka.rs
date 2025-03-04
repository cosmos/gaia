//! Relayer utilities for `solidity-ibc-eureka` chains.

use alloy::{primitives::Bytes, sol_types::SolValue};
use anyhow::Result;
use futures::future;
use ibc_eureka_solidity_types::{
    ics26::{
        router::{ackPacketCall, recvPacketCall, routerCalls},
        IICS02ClientMsgs::Height,
        IICS26RouterMsgs::{MsgAckPacket, MsgRecvPacket, MsgTimeoutPacket},
    },
    msgs::{
        IICS07TendermintMsgs::ClientState,
        IMembershipMsgs::{KVPair, MembershipProof, SP1MembershipAndUpdateClientProof},
        ISP1Msgs::SP1Proof,
    },
};
use sp1_ics07_tendermint_prover::{
    programs::UpdateClientAndMembershipProgram, prover::SP1ICS07TendermintProver,
};
use sp1_ics07_tendermint_utils::{light_block::LightBlockExt, rpc::TendermintRpcExt};
use sp1_prover::components::SP1ProverComponents;
use sp1_sdk::{HashableKey, Prover};
use tendermint_light_client_verifier::types::LightBlock;
use tendermint_rpc::HttpClient;

use crate::events::EurekaEvent;

/// Converts a list of [`EurekaEvent`]s to a list of [`routerCalls::timeoutPacket`]s with empty
/// proofs.
pub fn target_events_to_timeout_msgs(
    target_events: Vec<EurekaEvent>,
    target_client_id: &str,
    target_height: &Height,
    now: u64,
) -> Vec<routerCalls> {
    target_events
        .into_iter()
        .filter_map(|e| match e {
            EurekaEvent::SendPacket(packet) => {
                if now >= packet.timeoutTimestamp && packet.sourceClient == target_client_id {
                    Some(routerCalls::timeoutPacket(
                        ibc_eureka_solidity_types::ics26::router::timeoutPacketCall {
                            msg_: MsgTimeoutPacket {
                                packet,
                                proofHeight: target_height.clone(),
                                proofTimeout: Bytes::default(),
                            },
                        },
                    ))
                } else {
                    None
                }
            }
            EurekaEvent::WriteAcknowledgement(..) => None,
        })
        .collect()
}

/// Converts a list of [`EurekaEvent`]s to a list of [`routerCalls::recvPacket`]s and
/// [`routerCalls::ackPacket`]s with empty proofs.
pub fn src_events_to_recv_and_ack_msgs(
    src_events: Vec<EurekaEvent>,
    target_client_id: &str,
    target_height: &Height,
    now: u64,
) -> Vec<routerCalls> {
    src_events
        .into_iter()
        .filter_map(|e| match e {
            EurekaEvent::SendPacket(packet) => {
                if packet.timeoutTimestamp > now && packet.destClient == target_client_id {
                    Some(routerCalls::recvPacket(recvPacketCall {
                        msg_: MsgRecvPacket {
                            packet,
                            proofHeight: target_height.clone(),
                            proofCommitment: Bytes::default(),
                        },
                    }))
                } else {
                    None
                }
            }
            EurekaEvent::WriteAcknowledgement(packet, acks) => {
                if packet.sourceClient == target_client_id {
                    Some(routerCalls::ackPacket(ackPacketCall {
                        msg_: MsgAckPacket {
                            packet,
                            acknowledgement: acks[0].clone(), // TODO: handle multiple acks (#93)
                            proofHeight: target_height.clone(),
                            proofAcked: Bytes::default(),
                        },
                    }))
                } else {
                    None
                }
            }
        })
        .collect()
}

/// Generates and injects an SP1 proof into the first message in `msgs`.
/// # Errors
/// Returns an error if the sp1 proof cannot be generated.
pub async fn inject_sp1_proof<C: SP1ProverComponents>(
    sp1_prover: &dyn Prover<C>,
    msgs: &mut [routerCalls],
    tm_client: &HttpClient,
    target_light_block: LightBlock,
    client_state: ClientState,
    now: u64,
) -> Result<()> {
    let target_height = u32::try_from(target_light_block.height().value())?;

    let ibc_paths = msgs
        .iter()
        .map(|msg| match msg {
            routerCalls::timeoutPacket(call) => call.msg_.packet.receipt_commitment_path(),
            routerCalls::recvPacket(call) => call.msg_.packet.commitment_path(),
            routerCalls::ackPacket(call) => call.msg_.packet.ack_commitment_path(),
            _ => unreachable!(),
        })
        .map(|path| vec![b"ibc".into(), path]);

    let kv_proofs: Vec<(_, _)> = future::try_join_all(ibc_paths.into_iter().map(|path| async {
        let (value, proof) = tm_client.prove_path(&path, target_height).await?;
        let kv_pair = KVPair {
            path: path.into_iter().map(Into::into).collect(),
            value: value.into(),
        };
        anyhow::Ok((kv_pair, proof))
    }))
    .await?;

    let trusted_light_block = tm_client
        .get_light_block(Some(client_state.latestHeight.revisionHeight))
        .await?;

    // Get the proposed header from the target light block.
    let proposed_header = target_light_block.into_header(&trusted_light_block);

    let uc_and_mem_prover = SP1ICS07TendermintProver::<UpdateClientAndMembershipProgram, _>::new(
        client_state.zkAlgorithm,
        sp1_prover,
    );

    let uc_and_mem_proof = uc_and_mem_prover.generate_proof(
        &client_state,
        &trusted_light_block.to_consensus_state().into(),
        &proposed_header,
        now,
        kv_proofs,
    );

    let sp1_proof = MembershipProof::from(SP1MembershipAndUpdateClientProof {
        sp1Proof: SP1Proof::new(
            &uc_and_mem_prover.vkey.bytes32(),
            uc_and_mem_proof.bytes(),
            uc_and_mem_proof.public_values.to_vec(),
        ),
    });

    // inject proof
    match msgs.first_mut() {
        Some(routerCalls::timeoutPacket(ref mut call)) => {
            *call.msg_.proofTimeout = sp1_proof.abi_encode().into();
        }
        Some(routerCalls::recvPacket(ref mut call)) => {
            *call.msg_.proofCommitment = sp1_proof.abi_encode().into();
        }
        Some(routerCalls::ackPacket(ref mut call)) => {
            *call.msg_.proofAcked = sp1_proof.abi_encode().into();
        }
        _ => unreachable!(),
    }

    Ok(())
}
