// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import { IICS02ClientMsgs } from "../../msgs/IICS02ClientMsgs.sol";

/// @title ICS07 Tendermint Messages
/// @author srdtrk
/// @notice Defines shared types for ICS07Tendermint implementations.
interface IICS07TendermintMsgs {
    /// @notice Fraction of validator overlap needed to update header
    /// @param numerator Numerator of the fraction
    /// @param denominator Denominator of the fraction
    struct TrustThreshold {
        uint8 numerator;
        uint8 denominator;
    }

    /// @notice Defines the ICS07Tendermint ClientState for ibc-lite
    /// @param chainId Chain ID
    /// @param trustLevel Fraction of validator overlap needed to update header
    /// @param latestHeight Latest height the client was updated to
    /// @param trustingPeriod duration of the period since the LatestTimestamp during which the
    /// submitted headers are valid for upgrade in seconds.
    /// @param unbondingPeriod duration of the staking unbonding period in seconds
    /// @param isFrozen whether or not client is frozen (due to misbehavior)
    /// @param zkAlgorithm The zk algorithm supported by this contract (for the relayers).
    struct ClientState {
        string chainId;
        TrustThreshold trustLevel;
        IICS02ClientMsgs.Height latestHeight;
        uint32 trustingPeriod;
        uint32 unbondingPeriod;
        bool isFrozen;
        SupportedZkAlgorithm zkAlgorithm;
    }

    /// @notice Defines the Tendermint light client's consensus state at some height.
    /// @param timestamp timestamp that corresponds to the counterparty block height
    /// in which the ConsensusState was generated.
    /// @param root commitment root (i.e app hash)
    /// @param nextValidatorsHash next validators hash
    struct ConsensusState {
        uint64 timestamp;
        bytes32 root;
        bytes32 nextValidatorsHash;
    }

    /// @notice Defines the supported zk algorithms
    /// @param Groth16 Groth16 zk algorithm
    /// @param Plonk Plonk zk algorithm
    enum SupportedZkAlgorithm {
        Groth16,
        Plonk
    }
}
