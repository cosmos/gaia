use std::time::SystemTime;

use crate::type_urls::{
    GENERIC_AUTHORIZATION_TYPE_URL, MSG_EXEC_TYPE_URL, MSG_GRANT_TYPE_URL, MSG_MULTI_SEND_TYPE_URL,
    MSG_SEND_TYPE_URL,
};
use crate::utils::{
    bulk_get_user_keys, create_parameter_change_proposal, encode_any, get_test_token_name,
    get_user_key, one_atom, send_funds_bulk, vote_yes_on_proposals, wait_for_proposals_to_execute,
    CosmosUser, ValidatorKeys, ADDRESS_PREFIX, OPERATION_TIMEOUT, STAKING_TOKEN,
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
pub const XFER_FEE_BASIS_POINTS_PARAM_KEY: &str = "XferFeeBasisPoints";

/// Simulates activity of automated peer-to-peer transactions on Althea networks,
/// asserting that the correct fees are deducted and transfers succeed
pub async fn microtx_fees(contact: &Contact, validator_keys: Vec<ValidatorKeys>) {
    let users = bulk_get_user_keys(None, 128);
    let amount = Coin {
        amount: one_atom(),
        denom: get_test_token_name(),
    };
    send_funds_bulk(contact, keys.get(0).unwrap(), users, amount, None).await;
}
