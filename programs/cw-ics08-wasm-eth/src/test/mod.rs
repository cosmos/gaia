use std::marker::PhantomData;

use alloy_primitives::B256;
use cosmwasm_std::{
    testing::{
        mock_dependencies, MockApi, MockQuerier, MockQuerierCustomHandlerResult, MockStorage,
    },
    Binary, OwnedDeps, SystemResult,
};
use ethereum_light_client::test_utils::bls_verifier::{aggreagate, fast_aggregate_verify};
use ethereum_types::consensus::bls::{BlsPublicKey, BlsSignature};

use crate::custom_query::EthereumCustomQuery;

pub fn custom_query_handler(query: &EthereumCustomQuery) -> MockQuerierCustomHandlerResult {
    match query {
        EthereumCustomQuery::AggregateVerify {
            public_keys,
            message,
            signature,
        } => {
            let public_keys = public_keys
                .iter()
                .map(|pk| pk.as_ref().try_into().unwrap())
                .collect::<Vec<BlsPublicKey>>();
            let message = B256::try_from(message.as_slice()).unwrap();
            let signature = BlsSignature::try_from(signature.as_slice()).unwrap();

            fast_aggregate_verify(&public_keys, message, signature).unwrap();

            SystemResult::Ok(cosmwasm_std::ContractResult::Ok::<Binary>(
                serde_json::to_vec(&true).unwrap().into(),
            ))
        }
        EthereumCustomQuery::Aggregate { public_keys } => {
            let public_keys = public_keys
                .iter()
                .map(|pk| pk.as_ref().try_into().unwrap())
                .collect::<Vec<&BlsPublicKey>>();

            let aggregate_pubkey = aggreagate(&public_keys).unwrap();

            SystemResult::Ok(cosmwasm_std::ContractResult::Ok::<Binary>(
                serde_json::to_vec(&Binary::from(aggregate_pubkey.as_slice()))
                    .unwrap()
                    .into(),
            ))
        }
    }
}

pub fn mk_deps(
) -> OwnedDeps<MockStorage, MockApi, MockQuerier<EthereumCustomQuery>, EthereumCustomQuery> {
    let deps = mock_dependencies();

    OwnedDeps {
        storage: deps.storage,
        api: deps.api,
        querier: MockQuerier::<EthereumCustomQuery>::new(&[])
            .with_custom_handler(custom_query_handler),
        custom_query_type: PhantomData,
    }
}
