use clap::Parser;
use sp1_ics07_tendermint_operator::{
    cli::command::{fixtures, Commands, OperatorCli},
    runners::{
        self,
        fixtures::{membership, misbehaviour, uc_and_mem, update_client},
    },
};
use sp1_sdk::utils::setup_logger;

/// An implementation of a Tendermint Light Client operator that will poll an onchain Tendermint
/// light client and generate a proof of the transition from the latest block in the contract to the
/// latest block on the chain. Then, submits the proof to the contract and updates the contract with
/// the latest block hash and height.
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    setup_logger();

    if dotenv::dotenv().is_err() {
        tracing::warn!("No .env file found");
    }

    let cli = OperatorCli::parse();
    match cli.command {
        Commands::Start(args) => runners::operator::run(args).await,
        Commands::Genesis(args) => runners::genesis::run(args).await,
        Commands::Fixtures(cmd) => match cmd.command {
            fixtures::Cmds::UpdateClient(args) => update_client::run(args).await,
            fixtures::Cmds::Membership(args) => membership::run(args).await,
            fixtures::Cmds::UpdateClientAndMembership(args) => uc_and_mem::run(args).await,
            fixtures::Cmds::Misbehaviour(args) => misbehaviour::run(args).await,
        },
    }
}
