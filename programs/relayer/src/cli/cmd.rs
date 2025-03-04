//! Defines the client interface for the relayer server.

use clap::{command, Parser};

/// The command line interface for the relayer.
#[derive(Clone, Debug, Parser)]
#[command(version, about, long_about = None)]
pub struct RelayerCli {
    /// The subcommand to run.
    #[command(subcommand)]
    pub command: Commands,
}

/// The subcommands for the relayer.
#[derive(Clone, Debug, Parser)]
pub enum Commands {
    /// The subcommand to run the relayer.
    Start(start::Args),
}

/// The arguments for the start subcommand.
pub mod start {
    use super::Parser;

    /// The arguments for the start subcommand.
    #[derive(Clone, Debug, Parser)]
    pub struct Args {
        /// The configuration file for the relayer.
        #[clap(short = 'c', long)]
        pub config: String,
    }
}
