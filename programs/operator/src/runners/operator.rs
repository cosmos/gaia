//! Contains the runner for the `operator run` command.

use std::env;

use crate::cli::command::operator::Args;
use alloy::{providers::ProviderBuilder, sol_types::SolValue};
use ibc_eureka_solidity_types::{
    msgs::{ISP1Msgs::SP1Proof, IUpdateClientMsgs::MsgUpdateClient},
    sp1_ics07::sp1_ics07_tendermint,
};
use sp1_ics07_tendermint_prover::{
    programs::UpdateClientProgram, prover::SP1ICS07TendermintProver,
};
use sp1_ics07_tendermint_utils::{eth, light_block::LightBlockExt, rpc::TendermintRpcExt};
use sp1_sdk::{utils::setup_logger, HashableKey, ProverClient};
use tendermint_rpc::HttpClient;

/// Runs the update client program in a loop.
/// If the `only_once` flag is set, the program will only run once.
#[allow(clippy::missing_errors_doc, clippy::missing_panics_doc)]
pub async fn run(args: Args) -> anyhow::Result<()> {
    setup_logger();
    if dotenv::dotenv().is_err() {
        tracing::warn!("No .env file found");
    }

    let rpc_url = env::var("RPC_URL").expect("RPC_URL not set");
    let contract_address = env::var("CONTRACT_ADDRESS").expect("CONTRACT_ADDRESS not set");

    // Instantiate a Tendermint prover based on the environment variable.
    let wallet = eth::wallet_from_env();
    let provider = ProviderBuilder::new()
        .wallet(wallet)
        .on_http(rpc_url.parse()?);

    let contract = sp1_ics07_tendermint::new(contract_address.parse()?, provider);
    let contract_client_state = contract.clientState().call().await?;
    let tendermint_rpc_client = HttpClient::from_env();

    let sp1_prover = ProverClient::from_env();
    let prover = SP1ICS07TendermintProver::<UpdateClientProgram, _>::new(
        contract_client_state.zkAlgorithm.try_into()?,
        &sp1_prover,
    );

    loop {
        let contract_client_state = contract.clientState().call().await?;

        // Read the existing trusted header hash from the contract.
        let trusted_block_height = contract_client_state.latestHeight.revisionHeight;
        assert!(
            trusted_block_height != 0,
            "No trusted height found on the contract. Something is wrong with the contract."
        );

        let trusted_light_block = tendermint_rpc_client
            .get_light_block(Some(trusted_block_height))
            .await?;

        // Get trusted consensus state from the trusted light block.
        let trusted_consensus_state = trusted_light_block.to_consensus_state().into();

        let target_light_block = tendermint_rpc_client.get_light_block(None).await?;
        let target_height = target_light_block.height().value();

        // Get the proposed header from the target light block.
        let proposed_header = target_light_block.into_header(&trusted_light_block);

        let now = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)?
            .as_secs();

        // Generate a proof of the transition from the trusted block to the target block.
        let proof_data = prover.generate_proof(
            &contract_client_state.into(),
            &trusted_consensus_state,
            &proposed_header,
            now,
        );

        let update_msg = MsgUpdateClient {
            sp1Proof: SP1Proof::new(
                &prover.vkey.bytes32(),
                proof_data.bytes(),
                proof_data.public_values.to_vec(),
            ),
        };

        contract
            .updateClient(update_msg.abi_encode().into())
            .send()
            .await?
            .watch()
            .await?;

        tracing::info!(
            "Updated the ICS-07 Tendermint light client at address {} from block {} to block {}.",
            contract_address,
            trusted_block_height,
            target_height
        );

        if args.only_once {
            tracing::info!("Exiting because '--only-once' flag is set.");
            return Ok(());
        }

        // Sleep for 60 seconds.
        tracing::debug!("sleeping for 60 seconds");
        tokio::time::sleep(std::time::Duration::from_secs(60)).await;
    }
}
