use core::str::FromStr;
use std::thread;

use crate::utils::ValidatorKeys;
use crate::utils::OPERATION_TIMEOUT;
use crate::utils::{
    get_deposit, get_gravity_chain_id, get_ibc_chain_id, HERMES_CONFIG, RELAYER_ADDRESS,
    RELAYER_MNEMONIC,
};
use deep_space::private_key::{CosmosPrivateKey, PrivateKey, DEFAULT_COSMOS_HD_PATH};
use deep_space::Contact;
use ibc::core::ics24_host::identifier::ChainId;
use ibc_relayer::config::AddressType;
use ibc_relayer::keyring::{HDPath, KeyRing, Store};
use std::fs::File;
use std::io::{BufRead, BufReader};
use std::os::unix::io::{FromRawFd, IntoRawFd};
use std::process::{Command, Stdio};

/// Parses the output of the cosmoscli keys add command to import the private key
fn parse_phrases(filename: &str) -> (Vec<CosmosPrivateKey>, Vec<String>) {
    let file = File::open(filename).expect("Failed to find phrases");
    let reader = BufReader::new(file);
    let mut ret_keys = Vec::new();
    let mut ret_phrases = Vec::new();

    for line in reader.lines() {
        let phrase = line.expect("Error reading phrase file!");
        if phrase.is_empty()
            || phrase.contains("write this mnemonic phrase")
            || phrase.contains("recover your account if")
        {
            continue;
        }
        let key = CosmosPrivateKey::from_phrase(&phrase, "").expect("Bad phrase!");
        ret_keys.push(key);
        ret_phrases.push(phrase);
    }
    (ret_keys, ret_phrases)
}

/// Validator private keys are generated via the gravity key add
/// command, from there they are used to create gentx's and start the
/// chain, these keys change every time the container is restarted.
/// The mnemonic phrases are dumped into a text file /validator-phrases
/// the phrases are in increasing order, so validator 1 is the first key
/// and so on. While validators may later fail to start it is guaranteed
/// that we have one key for each validator in this file.
pub fn parse_validator_keys() -> (Vec<CosmosPrivateKey>, Vec<String>) {
    let filename = "/validator-phrases";
    info!("Reading mnemonics from {}", filename);
    parse_phrases(filename)
}

/// The same as parse_validator_keys() except for a second chain accessed
/// over IBC for testing purposes
pub fn parse_ibc_validator_keys() -> (Vec<CosmosPrivateKey>, Vec<String>) {
    let filename = "/ibc-validator-phrases";
    info!("Reading mnemonics from {}", filename);
    parse_phrases(filename)
}

pub fn get_keys() -> Vec<ValidatorKeys> {
    let (cosmos_keys, cosmos_phrases) = parse_validator_keys();
    let mut ret = Vec::new();
    for (c_key, c_phrase) in cosmos_keys.into_iter().zip(cosmos_phrases) {
        ret.push(ValidatorKeys {
            validator_key: c_key,
            validator_phrase: c_phrase,
        })
    }
    ret
}

// Creates a key in the relayer's test keyring, which the relayer should use
// Hermes stores its keys in hermes_home/ gravity_phrase is for the main chain
/// ibc phrase is for the test chain
pub fn setup_relayer_keys(
    gravity_phrase: &str,
    ibc_phrase: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    let mut keyring = KeyRing::new(
        Store::Test,
        "gravity",
        &ChainId::from_string(&get_gravity_chain_id()),
    )?;

    let key = keyring.key_from_mnemonic(
        gravity_phrase,
        &HDPath::from_str(DEFAULT_COSMOS_HD_PATH).unwrap(),
        &AddressType::Cosmos,
    )?;
    keyring.add_key("gravitykey", key)?;

    keyring = KeyRing::new(
        Store::Test,
        "cosmos",
        &ChainId::from_string(&get_ibc_chain_id()),
    )?;
    let key = keyring.key_from_mnemonic(
        ibc_phrase,
        &HDPath::from_str(DEFAULT_COSMOS_HD_PATH).unwrap(),
        &AddressType::Cosmos,
    )?;
    keyring.add_key("ibckey", key)?;

    Ok(())
}

// Create a channel between gravity chain and the ibc test chain over the "transfer" port
// Writes the output to /ibc-relayer-logs/channel-creation
pub fn create_ibc_channel(hermes_base: &mut Command) {
    // hermes -c config.toml create channel gravity-test-1 ibc-test-1 --port-a transfer --port-b transfer
    let create_channel = hermes_base.args([
        "create",
        "channel",
        &get_gravity_chain_id(),
        &get_ibc_chain_id(),
        "--port-a",
        "transfer",
        "--port-b",
        "transfer",
    ]);

    let out_file = File::options()
        .write(true)
        .open("/ibc-relayer-logs/channel-creation")
        .unwrap()
        .into_raw_fd();
    unsafe {
        // unsafe needed for stdout + stderr redirect to file
        let create_channel = create_channel
            .stdout(Stdio::from_raw_fd(out_file))
            .stderr(Stdio::from_raw_fd(out_file));
        info!("Create channel command: {:?}", create_channel);
        create_channel.spawn().expect("Could not create channel");
    }
}

// Start an IBC relayer locally and run until it terminates
// full_scan Force a full scan of the chains for clients, connections and channels
// Writes the output to /ibc-relayer-logs/hermes-logs
pub fn run_ibc_relayer(hermes_base: &mut Command, full_scan: bool) {
    let mut start = hermes_base.arg("start");
    if full_scan {
        start = start.arg("-f");
    }
    let out_file = File::options()
        .write(true)
        .open("/ibc-relayer-logs/hermes-logs")
        .unwrap()
        .into_raw_fd();
    unsafe {
        // unsafe needed for stdout + stderr redirect to file
        start
            .stdout(Stdio::from_raw_fd(out_file))
            .stderr(Stdio::from_raw_fd(out_file))
            .spawn()
            .expect("Could not run hermes");
    }
}

// starts up the IBC relayer (hermes) in a background thread
pub async fn start_ibc_relayer(contact: &Contact, keys: &[ValidatorKeys], ibc_phrases: &[String]) {
    contact
        .send_coins(
            get_deposit(),
            None,
            *RELAYER_ADDRESS,
            Some(OPERATION_TIMEOUT),
            keys[0].validator_key,
        )
        .await
        .unwrap();
    info!("test-runner starting IBC relayer mode: init hermes, create ibc channel, start hermes");
    let mut hermes_base = Command::new("hermes");
    let hermes_base = hermes_base.arg("-c").arg(HERMES_CONFIG);
    setup_relayer_keys(&RELAYER_MNEMONIC, &ibc_phrases[0]).unwrap();
    create_ibc_channel(hermes_base);
    thread::spawn(|| {
        let mut hermes_base = Command::new("hermes");
        let hermes_base = hermes_base.arg("-c").arg(HERMES_CONFIG);
        run_ibc_relayer(hermes_base, true); // likely will not return from here, just keep running
    });
    info!("Running ibc relayer in the background, directing output to /ibc-relayer-logs");
}
