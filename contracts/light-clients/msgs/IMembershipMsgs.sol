// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import { ISP1Msgs } from "./ISP1Msgs.sol";
import { IICS07TendermintMsgs } from "./IICS07TendermintMsgs.sol";

/// @title Membership Program Messages
/// @author srdtrk
/// @notice Defines shared types for the verify (non)membership program.
interface IMembershipMsgs {
    /// @notice The key-value pair used in the verify (non)membership program.
    /// @param path The path of the value in the key-value store.
    /// @param value The value of the key-value pair.
    struct KVPair {
        bytes[] path;
        bytes value;
    }

    /// @notice The public value output for the sp1 verify (non)membership program.
    /// @param commitmentRoot The app hash of the header.
    /// @param kvPairs The key-value pairs verified by the program.
    struct MembershipOutput {
        bytes32 commitmentRoot;
        KVPair[] kvPairs;
    }

    /// @notice The membership proof that can be submitted to the SP1Verifier contract.
    /// @param proofType The type of the membership proof.
    /// @param proof The membership proof.
    struct MembershipProof {
        MembershipProofType proofType;
        bytes proof;
    }

    /// @notice The membership proof for the sp1 verify (non)membership program.
    /// @param sp1Proof The sp1 proof for the membership program.
    /// @param trustedConsensusState The trusted consensus state that the proof is based on.
    struct SP1MembershipProof {
        ISP1Msgs.SP1Proof sp1Proof;
        IICS07TendermintMsgs.ConsensusState trustedConsensusState;
    }

    /// @notice The membership proof for the sp1 verify (non)membership and update client program.
    /// @param sp1Proof The sp1 proof for the membership and update client program.
    struct SP1MembershipAndUpdateClientProof {
        ISP1Msgs.SP1Proof sp1Proof;
    }

    /// @notice The type of the membership proof.
    enum MembershipProofType {
        /// The proof is for the verify membership program.
        SP1MembershipProof,
        /// The proof is for the verify membership and update client program.
        SP1MembershipAndUpdateClientProof
    }
}
