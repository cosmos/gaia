use crate::type_urls::MSG_XFER_TYPE_URL;
use crate::utils::{
    bulk_get_user_keys, get_test_token_name, one_atom_128, send_funds_bulk, CosmosUser,
    ValidatorKeys, OPERATION_TIMEOUT, STAKING_TOKEN,
};
use althea_proto::cosmos_sdk_proto::cosmos::base::v1beta1::Coin as ProtoCoin;
use althea_proto::microtx::v1::MsgXfer;
use clarity::Uint256;
use deep_space::{Address, Coin, Contact, Msg};
use rand::distributions::Uniform;
use rand::{thread_rng, Rng};

pub const BASIS_POINTS_DIVISOR: u128 = 10_000;
pub const MICROTX_SUBSPACE: &str = "microtx";
/// This PARAM_KEY constant is defined in x/microtx/types/genesis.go and must match exactly
pub const XFER_FEE_BASIS_POINTS_PARAM_KEY: &str = "XferFeeBasisPoints";

/// Simulates activity of automated peer-to-peer transactions on Althea networks,
/// asserting that the correct fees are deducted and transfers succeed
pub async fn microtx_fees_test(contact: &Contact, validator_keys: Vec<ValidatorKeys>) {
    let num_users = 128;
    // Make users who will send tokens
    let senders = bulk_get_user_keys(None, num_users);

    // Make users who will receive tokens via MsgXfer
    let receivers = bulk_get_user_keys(None, num_users);

    // Send one footoken to each sender
    let foo_balance = one_atom_128();
    let foo_denom = get_test_token_name();
    let amount = Coin {
        amount: foo_balance.into(),
        denom: foo_denom.clone(),
    };
    send_funds_bulk(
        contact,
        validator_keys.get(0).unwrap().validator_key,
        &senders
            .clone()
            .iter()
            .map(|u| u.cosmos_address.clone())
            .collect::<Vec<Address>>(),
        amount,
        Some(OPERATION_TIMEOUT),
    )
    .await
    .expect("Unable to send funds to all senders!");

    let param = contact
        .get_param(MICROTX_SUBSPACE, XFER_FEE_BASIS_POINTS_PARAM_KEY)
        .await
        .expect("Unable to get XferFeeBasisPoints from microtx module");
    let param = param.param.unwrap().value;
    let xfer_fee_basis_points = param.trim_matches('"');
    info!("Got xfer_fee_basis_points: [{}]", xfer_fee_basis_points);

    let xfer_fee_basis_points: u128 = serde_json::from_str(&xfer_fee_basis_points).unwrap();
    let (xfers, amounts, fees) = generate_msg_xfers(
        &senders,
        &receivers,
        &foo_denom,
        foo_balance,
        xfer_fee_basis_points,
    );

    // Send the MsgXfers, check their execution, assert the balances have changed
    exec_and_check(contact, &senders, &xfers, &amounts, &fees, foo_balance).await;

    // Check that the senders and receivers have the expected balance
    assert_balance_changes(
        contact,
        &senders,
        &receivers,
        &amounts,
        &fees,
        foo_balance,
        &foo_denom,
    )
    .await;
}

/// Creates 3 Vec's: MsgXfer's, transfer amounts, and expected fees
/// The generated MsgXfer's will have a randomized transfer amount and a derived associated fee
/// Order is preserved so the i-th Msg corresponds to the i-th amount and the i-th fee
pub fn generate_msg_xfers(
    senders: &[CosmosUser],
    receivers: &[CosmosUser],
    denom: &str,
    sender_balance: u128,
    xfer_fee_basis_points: u128,
) -> (Vec<Msg>, Vec<Uint256>, Vec<Uint256>) {
    let mut msgs = Vec::with_capacity(senders.len());
    let mut amounts = Vec::with_capacity(senders.len());
    let mut fees = Vec::with_capacity(senders.len());

    let mut rng = thread_rng();
    let amount_range = Uniform::new(0u128, sender_balance);
    for (i, (sender, receiver)) in senders.into_iter().zip(receivers.into_iter()).enumerate() {
        let amount: u128 = if i == 0 {
            // Guarantee one MsgXfer failure
            sender_balance
        } else {
            rng.sample(amount_range)
        };
        let expected_fee: u128 = amount * xfer_fee_basis_points / BASIS_POINTS_DIVISOR;
        let amount_coin = ProtoCoin {
            denom: denom.to_string(),
            amount: amount.to_string(),
        };
        let amount_coins = vec![amount_coin];

        let msg_xfer = MsgXfer {
            receiver: receiver.cosmos_address.to_string(),
            sender: sender.cosmos_address.to_string(),
            amounts: amount_coins,
        };
        let msg = Msg::new(MSG_XFER_TYPE_URL, msg_xfer);

        msgs.push(msg);
        amounts.push(amount.into());
        fees.push(expected_fee.into());
        info!(
            "{}: {} (+ {}) -> {}",
            sender.cosmos_address,
            amount.to_string(),
            expected_fee.to_string(),
            receiver.cosmos_address,
        );
    }

    (msgs, amounts, fees)
}

/// Executes the given `msgs`, checking that the associated `msg_amounts` and
/// `msg_exp_fees` have been deducted from the accounts except in the situation
/// where an amount and fee total higher than the account's balance
pub async fn exec_and_check(
    contact: &Contact,
    senders: &[CosmosUser],
    msgs: &[Msg],
    msg_amounts: &[Uint256],
    msg_exp_fees: &[Uint256],
    token_balance: u128,
) {
    let zero_fee = Coin {
        amount: 0u8.into(),
        denom: STAKING_TOKEN.clone(),
    };
    let token_balance: Uint256 = token_balance.into();
    for (((sender, msg), amt), exp_fee) in senders
        .into_iter()
        .zip(msgs.into_iter())
        .zip(msg_amounts.into_iter())
        .zip(msg_exp_fees.into_iter())
    {
        let res = contact
            .send_message(
                &[msg.clone()],
                None,
                &[zero_fee.clone()],
                Some(OPERATION_TIMEOUT),
                sender.cosmos_key,
            )
            .await;
        if token_balance < amt.clone() + exp_fee.clone() {
            // FAILURE CASE
            assert!(
                res.is_err(),
                "Unexpected success when sending more than {}: address {}, amt {}, fee {}",
                token_balance,
                sender.cosmos_address,
                amt,
                exp_fee
            );
        } else {
            // SUCCESS CASE
            assert!(
                res.is_ok(),
                "Unexpected failure when sending <= {}: address {}, amt {}, fee {}: res {:?}",
                token_balance,
                sender.cosmos_address,
                amt,
                exp_fee,
                res,
            );
        }
        debug!("Sent MsgXfer with response {:?}", res);
    }
}

/// Asserts that the senders have appropriate reduced balances and receivers
/// have increased balances, accounting for expected failures
pub async fn assert_balance_changes(
    contact: &Contact,
    senders: &[CosmosUser],
    receivers: &[CosmosUser],
    msg_amounts: &[Uint256],
    msg_exp_fees: &[Uint256],
    token_balance: u128,
    token_denom: &str,
) {
    let token_balance: Uint256 = token_balance.into();
    for (((sender, receiver), amt), exp_fee) in senders
        .into_iter()
        .zip(receivers.into_iter())
        .zip(msg_amounts.into_iter())
        .zip(msg_exp_fees.into_iter())
    {
        let sender_bal = contact
            .get_balance(sender.cosmos_address, token_denom.to_string())
            .await
            .unwrap();
        let receiver_bal = contact
            .get_balance(receiver.cosmos_address, token_denom.to_string())
            .await
            .unwrap();
        let sender_bal = match sender_bal {
            Some(v) => v.amount,
            None => 0u8.into(),
        };
        let receiver_bal = match receiver_bal {
            Some(v) => v.amount,
            None => 0u8.into(),
        };

        if token_balance < amt.clone() + exp_fee.clone() {
            // FAILURE CASE
            let exp_send_bal: Uint256 = token_balance.clone();
            let exp_recv_bal: Uint256 = 0u8.into();

            assert!(
                sender_bal == exp_send_bal && receiver_bal == exp_recv_bal,
                "Expected unchanged balances, found sender {} balance ({}), receiver {} balance ({})",
                sender.cosmos_address.to_string(),
                sender_bal,
                receiver.cosmos_address.to_string(),
                receiver_bal,
            );
        } else {
            // SUCCESS CASE
            let exp_send_bal: Uint256 = token_balance.clone() - amt.clone() - exp_fee.clone();
            let exp_recv_bal: Uint256 = amt.clone();

            assert!(
                sender_bal == exp_send_bal && receiver_bal == exp_recv_bal,
                "Expected balance transfer less fee, found sender {} balance ({}), receiver {} balance ({})",
                sender.cosmos_address.to_string(),
                sender_bal,
                receiver.cosmos_address.to_string(),
                receiver_bal,
            );
        }
    }
}
