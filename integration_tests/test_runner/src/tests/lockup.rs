use crate::type_urls::{
    GENERIC_AUTHORIZATION_TYPE_URL, MSG_EXEC_TYPE_URL, MSG_GRANT_TYPE_URL, MSG_MULTI_SEND_TYPE_URL,
    MSG_SEND_TYPE_URL,
};
use crate::utils::{
    create_parameter_change_proposal, encode_any, get_user_key, one_atom, send_funds_bulk,
    vote_yes_on_proposals, wait_for_proposals_to_execute, CosmosUser, ValidatorKeys,
    ADDRESS_PREFIX, OPERATION_TIMEOUT, STAKING_TOKEN,
};
use althea_proto::cosmos_sdk_proto::cosmos::authz::v1beta1::{
    GenericAuthorization, Grant, MsgExec, MsgGrant,
};
use althea_proto::cosmos_sdk_proto::cosmos::bank::v1beta1::{Input, MsgMultiSend, MsgSend, Output};
use althea_proto::cosmos_sdk_proto::cosmos::base::v1beta1::Coin as ProtoCoin;
use althea_proto::cosmos_sdk_proto::cosmos::params::v1beta1::ParamChange;
use clarity::Uint256;
use deep_space::error::CosmosGrpcError;
use deep_space::{Address, Coin, Contact, Msg, PrivateKey};

/// These *_PARAM_KEY constants are defined in x/lockup/types/types.go and must match those values exactly
pub const LOCKED_PARAM_KEY: &str = "locked";
pub const LOCKED_MSG_TYPES_PARAM_KEY: &str = "lockedMessageTypes";
pub const LOCK_EXEMPT_PARAM_KEY: &str = "lockExempt";

/// Simulates the launch lockup process by setting the lockup module params via governance,
/// attempting to transfer tokens a variety of ways, and finally clearing the lockup module params
/// and asserting that balances can successfully be transferred
pub async fn lockup_test(contact: &Contact, validator_keys: Vec<ValidatorKeys>) {
    let lock_exempt = get_user_key(None);
    fund_lock_exempt_user(contact, &validator_keys, lock_exempt).await;
    lockup_the_chain(contact, &validator_keys, &lock_exempt).await;

    fail_to_send(contact, &validator_keys).await;
    send_from_lock_exempt(contact, lock_exempt).await;

    unlock_the_chain(contact, &validator_keys).await;
    successfully_send(contact, &validator_keys, lock_exempt).await;
}

async fn fund_lock_exempt_user(
    contact: &Contact,
    validator_keys: &[ValidatorKeys],
    lock_exempt: CosmosUser,
) {
    let sender = validator_keys.get(0).unwrap().validator_key;
    let amount = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: one_atom() * 100u16.into(),
    };
    contact
        .send_coins(
            amount.clone(),
            Some(amount),
            lock_exempt.cosmos_address,
            Some(OPERATION_TIMEOUT),
            sender,
        )
        .await
        .expect("Unable to send funds to lock exempt user!");
}

pub async fn lockup_the_chain(
    contact: &Contact,
    validator_keys: &[ValidatorKeys],
    lock_exempt: &CosmosUser,
) {
    let to_change = create_lockup_param_changes(lock_exempt.cosmos_address);
    let proposer = validator_keys.get(0).unwrap();
    let zero_fee = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: 0u8.into(),
    };
    create_parameter_change_proposal(contact, proposer.validator_key, to_change, zero_fee).await;

    vote_yes_on_proposals(contact, validator_keys, Some(OPERATION_TIMEOUT)).await;
    wait_for_proposals_to_execute(contact).await;
}

pub fn create_lockup_param_changes(exempt_user: Address) -> Vec<ParamChange> {
    // Params{lock_exempt:, locked: false, locked_message_types: Vec::new() };
    let lockup_param = ParamChange {
        subspace: "lockup".to_string(),
        key: String::new(),
        value: String::new(),
    };
    let mut locked = lockup_param.clone();
    locked.key = LOCKED_PARAM_KEY.to_string();
    locked.value = format!("{}", true);

    let mut lock_exempt = lockup_param.clone();
    lock_exempt.key = LOCK_EXEMPT_PARAM_KEY.to_string();
    lock_exempt.value = serde_json::to_string(&vec![exempt_user.to_string()]).unwrap();

    let locked_msgs = vec![
        MSG_SEND_TYPE_URL.to_string(),
        MSG_MULTI_SEND_TYPE_URL.to_string(),
    ];
    let mut locked_msg_types = lockup_param;
    locked_msg_types.key = LOCKED_MSG_TYPES_PARAM_KEY.to_string();
    locked_msg_types.value = serde_json::to_string(&locked_msgs).unwrap();

    vec![locked, lock_exempt, locked_msg_types]
}

pub async fn fail_to_send(contact: &Contact, validator_keys: &[ValidatorKeys]) {
    let sender = validator_keys.get(0).unwrap().validator_key;
    let receiver = get_user_key(None);
    let amount = ProtoCoin {
        denom: STAKING_TOKEN.clone(),
        amount: one_atom().to_string(),
    };

    let msg_send = create_bank_msg_send(sender, receiver.cosmos_address, amount.clone());
    let res = contact
        .send_message(&[msg_send], None, &[], Some(OPERATION_TIMEOUT), sender)
        .await;
    info!("Tried to send via bank MsgSend: {:?}", res);
    res.expect_err("Successfully sent? Should not be possible!");
    let msg_multi_send =
        create_bank_msg_multi_send(sender, receiver.cosmos_address, amount.clone());
    let res = contact
        .send_message(
            &[msg_multi_send],
            None,
            &[],
            Some(OPERATION_TIMEOUT),
            sender,
        )
        .await;
    info!("Tried to send via bank MsgMultiSend: {:?}", res);
    res.expect_err("Successfully sent? Should not be possible!");
    let (authz_send, authorized) =
        create_authz_bank_msg_send(contact, sender, receiver.cosmos_address, amount.clone())
            .await
            .unwrap();
    let res = contact
        .send_message(
            &[authz_send],
            None,
            &[],
            Some(OPERATION_TIMEOUT),
            authorized.cosmos_key,
        )
        .await;
    info!("Tried to send via authz MsgSend: {:?}", res);
    res.expect_err("Successfully sent? Should not be possible!");
    let (authz_multi_send, authorized) =
        create_authz_bank_msg_multi_send(contact, sender, receiver.cosmos_address, amount.clone())
            .await
            .unwrap();
    let res = contact
        .send_message(
            &[authz_multi_send],
            None,
            &[],
            Some(OPERATION_TIMEOUT),
            authorized.cosmos_key,
        )
        .await;
    info!("Tried to send via authz MsgSend: {:?}", res);
    res.expect_err("Successfully sent? Should not be possible!");
}

/// Creates a x/bank MsgSend to transfer `amount` from `sender` to `receiver`
pub fn create_bank_msg_send(sender: impl PrivateKey, receiver: Address, amount: ProtoCoin) -> Msg {
    let send = MsgSend {
        from_address: sender.to_address(&ADDRESS_PREFIX).unwrap().to_string(),
        to_address: receiver.to_string(),
        amount: vec![amount],
    };
    Msg::new(MSG_SEND_TYPE_URL, send)
}

/// Creates a x/bank MsgMultiSend to transfer `amount` from `sender` to `receiver`
pub fn create_bank_msg_multi_send(
    sender: impl PrivateKey,
    receiver: Address,
    amount: ProtoCoin,
) -> Msg {
    let input = Input {
        address: sender.to_address(&ADDRESS_PREFIX).unwrap().to_string(),
        coins: vec![amount.clone()],
    };
    let output = Output {
        address: receiver.to_string(),
        coins: vec![amount],
    };
    let multi_send = MsgMultiSend {
        inputs: vec![input],
        outputs: vec![output],
    };

    Msg::new(MSG_MULTI_SEND_TYPE_URL, multi_send)
}

/// Submits an Authorization using x/authz to give the returned private key control over `sender`'s tokens, then crafts
/// an authz MsgExec-wrapped bank MsgSend and returns that as well
pub async fn create_authz_bank_msg_send(
    contact: &Contact,
    sender: impl PrivateKey,
    receiver: Address,
    amount: ProtoCoin,
) -> Result<(Msg, CosmosUser), CosmosGrpcError> {
    let authorizee = get_user_key(None);
    let grant_msg_send = create_authorization(
        sender.clone(),
        authorizee.cosmos_address,
        MSG_SEND_TYPE_URL.to_string(),
    );

    let res = contact
        .send_message(
            &[grant_msg_send],
            None,
            &[],
            Some(OPERATION_TIMEOUT),
            sender.clone(),
        )
        .await;
    info!("Granted MsgSend authorization with response {:?}", res);
    res?;

    let send = create_bank_msg_send(sender.clone(), receiver, amount);
    let send_any: prost_types::Any = send.into();
    let exec = MsgExec {
        grantee: authorizee.cosmos_address.to_string(),
        msgs: vec![send_any],
    };
    let exec_msg = Msg::new(MSG_EXEC_TYPE_URL, exec);

    Ok((exec_msg, authorizee))
}

/// Submits an Authorization using x/authz to give the returned private key control over `sender`'s tokens, then crafts
/// an authz MsgExec-wrapped bank MsgMultiSend and returns that as well
pub async fn create_authz_bank_msg_multi_send(
    contact: &Contact,
    sender: impl PrivateKey,
    receiver: Address,
    amount: ProtoCoin,
) -> Result<(Msg, CosmosUser), CosmosGrpcError> {
    let authorizee = get_user_key(None);
    let grant_msg_multi_send = create_authorization(
        sender.clone(),
        authorizee.cosmos_address,
        MSG_MULTI_SEND_TYPE_URL.to_string(),
    );

    let res = contact
        .send_message(
            &[grant_msg_multi_send],
            None,
            &[],
            Some(OPERATION_TIMEOUT),
            sender.clone(),
        )
        .await;
    info!("Granted MsgSend authorization with response {:?}", res);
    res?;

    let multi_send = create_bank_msg_multi_send(sender.clone(), receiver, amount);
    let multi_send_any: prost_types::Any = multi_send.into();
    let exec = MsgExec {
        grantee: authorizee.cosmos_address.to_string(),
        msgs: vec![multi_send_any],
    };
    let exec_msg = Msg::new(MSG_EXEC_TYPE_URL, exec);

    Ok((exec_msg, authorizee))
}

/// Creates a MsgGrant to give a GenericAuthorization for `authorizee` to submit any Msg with the given `msg_type_url`
/// on behalf of `authorizer`
pub fn create_authorization(
    authorizer: impl PrivateKey,
    authorizee: Address,
    msg_type_url: String,
) -> Msg {
    let granter = authorizer.to_address(&ADDRESS_PREFIX).unwrap().to_string();

    // The authorization we want to store
    let auth = GenericAuthorization { msg: msg_type_url };
    let auth_any = encode_any(auth, GENERIC_AUTHORIZATION_TYPE_URL.to_string());

    // The authorization and any associated auth expiration
    let grant = Grant {
        authorization: Some(auth_any),
        expiration: None,
    };

    // The msg which must be submitted by the granter to give the grantee the specific authorization (with expiration)
    let msg_grant = MsgGrant {
        granter,
        grantee: authorizee.to_string(),
        grant: Some(grant),
    };

    Msg::new(MSG_GRANT_TYPE_URL, msg_grant)
}

async fn send_from_lock_exempt(contact: &Contact, lock_exempt: CosmosUser) {
    let amount = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: one_atom(),
    };

    send_from_and_assert_balance_changes(contact, lock_exempt.cosmos_key, amount).await;
}

pub async fn send_from_and_assert_balance_changes(
    contact: &Contact,
    from: impl PrivateKey,
    amount: Coin,
) {
    let receiver = get_user_key(None);
    let pre_balance = contact
        .get_balance(receiver.cosmos_address, STAKING_TOKEN.clone())
        .await
        .unwrap();
    send_funds_bulk(
        contact,
        from.clone(),
        &[receiver.cosmos_address],
        amount.clone(),
        Some(OPERATION_TIMEOUT),
    )
    .await
    .unwrap();
    let post_balance = contact
        .get_balance(receiver.cosmos_address, STAKING_TOKEN.clone())
        .await
        .unwrap();
    assert_balance_changes(pre_balance, post_balance, amount.amount);
}

pub fn assert_balance_changes(
    pre_balance: Option<Coin>,
    post_balance: Option<Coin>,
    expected_amount: Uint256,
) {
    let diff: Uint256 = match (pre_balance, post_balance) {
        (Some(pre), Some(post)) => {
            if post.amount < pre.amount {
                panic!("Unexpected lesser balance!");
            }
            post.amount - pre.amount
        }
        (None, Some(post)) => post.amount,
        (_, _) => {
            panic!("Unexpected balance change!");
        }
    };
    if diff != expected_amount {
        panic!("Unexpected diff: {}, expected {}", diff, expected_amount);
    }
}

async fn unlock_the_chain(contact: &Contact, validator_keys: &[ValidatorKeys]) {
    let unlock = ParamChange {
        subspace: "lockup".to_string(),
        key: LOCKED_PARAM_KEY.to_string(),
        value: format!("{}", false),
    };
    let proposer = validator_keys.get(0).unwrap();
    let zero_fee = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: 0u8.into(),
    };
    create_parameter_change_proposal(contact, proposer.validator_key, vec![unlock], zero_fee).await;

    vote_yes_on_proposals(contact, validator_keys, Some(OPERATION_TIMEOUT)).await;
    wait_for_proposals_to_execute(contact).await;
}

async fn successfully_send(
    contact: &Contact,
    validator_keys: &[ValidatorKeys],
    lock_exempt: CosmosUser,
) {
    let val0 = validator_keys.get(0).unwrap().validator_key;
    let amount = Coin {
        denom: STAKING_TOKEN.clone(),
        amount: one_atom(),
    };
    send_from_and_assert_balance_changes(contact, val0, amount.clone()).await;
    send_from_and_assert_balance_changes(contact, lock_exempt.cosmos_key, amount.clone()).await;
}
