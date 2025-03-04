// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { IICS07TendermintMsgs } from "./msgs/IICS07TendermintMsgs.sol";
import { IUpdateClientMsgs } from "./msgs/IUpdateClientMsgs.sol";
import { IMembershipMsgs } from "./msgs/IMembershipMsgs.sol";
import { IUpdateClientAndMembershipMsgs } from "./msgs/IUcAndMembershipMsgs.sol";
import { IMisbehaviourMsgs } from "./msgs/IMisbehaviourMsgs.sol";
import { ISP1Msgs } from "./msgs/ISP1Msgs.sol";
import { ILightClientMsgs } from "../msgs/ILightClientMsgs.sol";
import { IICS02ClientMsgs } from "../msgs/IICS02ClientMsgs.sol";

import { ISP1ICS07TendermintErrors } from "./errors/ISP1ICS07TendermintErrors.sol";
import { ISP1ICS07Tendermint } from "./ISP1ICS07Tendermint.sol";
import { ILightClient } from "../interfaces/ILightClient.sol";
import { ISP1Verifier } from "@sp1-contracts/ISP1Verifier.sol";

import { Paths } from "./utils/Paths.sol";
import { Multicall } from "@openzeppelin-contracts/utils/Multicall.sol";
import { TransientSlot } from "@openzeppelin-contracts/utils/TransientSlot.sol";

/// @title SP1 ICS07 Tendermint Light Client
/// @author srdtrk
/// @notice This contract implements an ICS07 IBC tendermint light client using SP1.
/// @custom:poc This is a proof of concept implementation.
contract SP1ICS07Tendermint is ISP1ICS07TendermintErrors, ISP1ICS07Tendermint, ILightClient, Multicall {
    using TransientSlot for *;

    /// @inheritdoc ISP1ICS07Tendermint
    bytes32 public immutable UPDATE_CLIENT_PROGRAM_VKEY;
    /// @inheritdoc ISP1ICS07Tendermint
    bytes32 public immutable MEMBERSHIP_PROGRAM_VKEY;
    /// @inheritdoc ISP1ICS07Tendermint
    bytes32 public immutable UPDATE_CLIENT_AND_MEMBERSHIP_PROGRAM_VKEY;
    /// @inheritdoc ISP1ICS07Tendermint
    bytes32 public immutable MISBEHAVIOUR_PROGRAM_VKEY;
    /// @inheritdoc ISP1ICS07Tendermint
    ISP1Verifier public immutable VERIFIER;

    /// @notice The ICS07Tendermint client state
    IICS07TendermintMsgs.ClientState public clientState;
    /// @notice The mapping from height to consensus state keccak256 hashes.
    mapping(uint32 height => bytes32 hash) private consensusStateHashes;

    /// @notice Allowed clock drift.
    /// @inheritdoc ISP1ICS07Tendermint
    uint16 public constant ALLOWED_SP1_CLOCK_DRIFT = 30 minutes;

    /// @notice The constructor sets the program verification key and the initial client and consensus states.
    /// @param updateClientProgramVkey The verification key for the update client program.
    /// @param membershipProgramVkey The verification key for the verify (non)membership program.
    /// @param updateClientAndMembershipProgramVkey The verification key for the update client and membership program.
    /// @param misbehaviourProgramVkey The verification key for the misbehaviour program.
    /// @param sp1Verifier The address of the SP1 verifier contract.
    /// @param _clientState The encoded initial client state.
    /// @param _consensusState The encoded initial consensus state.
    constructor(
        bytes32 updateClientProgramVkey,
        bytes32 membershipProgramVkey,
        bytes32 updateClientAndMembershipProgramVkey,
        bytes32 misbehaviourProgramVkey,
        address sp1Verifier,
        bytes memory _clientState,
        bytes32 _consensusState
    ) {
        UPDATE_CLIENT_PROGRAM_VKEY = updateClientProgramVkey;
        MEMBERSHIP_PROGRAM_VKEY = membershipProgramVkey;
        UPDATE_CLIENT_AND_MEMBERSHIP_PROGRAM_VKEY = updateClientAndMembershipProgramVkey;
        MISBEHAVIOUR_PROGRAM_VKEY = misbehaviourProgramVkey;

        clientState = abi.decode(_clientState, (IICS07TendermintMsgs.ClientState));
        consensusStateHashes[clientState.latestHeight.revisionHeight] = _consensusState;

        VERIFIER = ISP1Verifier(sp1Verifier);

        require(
            clientState.trustingPeriod + ALLOWED_SP1_CLOCK_DRIFT <= clientState.unbondingPeriod,
            TrustingPeriodTooLong(clientState.trustingPeriod, clientState.unbondingPeriod)
        );
    }

    /// @inheritdoc ILightClient
    function getClientState() external view returns (bytes memory) {
        return abi.encode(clientState);
    }

    /// @inheritdoc ISP1ICS07Tendermint
    function getConsensusStateHash(uint32 revisionHeight) public view returns (bytes32) {
        bytes32 hash = consensusStateHashes[revisionHeight];
        require(hash != 0, ConsensusStateNotFound());
        return hash;
    }

    /// @notice The entrypoint for updating the client.
    /// @dev This function verifies the public values and forwards the proof to the SP1 verifier.
    /// @param updateMsg The encoded update message.
    /// @return The result of the update.
    /// @inheritdoc ILightClient
    function updateClient(bytes calldata updateMsg) public notFrozen returns (ILightClientMsgs.UpdateResult) {
        IUpdateClientMsgs.MsgUpdateClient memory msgUpdateClient =
            abi.decode(updateMsg, (IUpdateClientMsgs.MsgUpdateClient));
        require(
            msgUpdateClient.sp1Proof.vKey == UPDATE_CLIENT_PROGRAM_VKEY,
            VerificationKeyMismatch(UPDATE_CLIENT_PROGRAM_VKEY, msgUpdateClient.sp1Proof.vKey)
        );

        IUpdateClientMsgs.UpdateClientOutput memory output =
            abi.decode(msgUpdateClient.sp1Proof.publicValues, (IUpdateClientMsgs.UpdateClientOutput));

        validateUpdateClientPublicValues(output);

        ILightClientMsgs.UpdateResult updateResult = checkUpdateResult(output);
        if (updateResult == ILightClientMsgs.UpdateResult.Update) {
            // adding the new consensus state to the mapping
            if (output.newHeight.revisionHeight > clientState.latestHeight.revisionHeight) {
                clientState.latestHeight = output.newHeight;
            }
            consensusStateHashes[output.newHeight.revisionHeight] = keccak256(abi.encode(output.newConsensusState));
        } else if (updateResult == ILightClientMsgs.UpdateResult.Misbehaviour) {
            clientState.isFrozen = true;
        } else if (updateResult == ILightClientMsgs.UpdateResult.NoOp) {
            return ILightClientMsgs.UpdateResult.NoOp;
        }

        verifySP1Proof(msgUpdateClient.sp1Proof);

        return updateResult;
    }

    /// @inheritdoc ILightClient
    function verifyMembership(ILightClientMsgs.MsgVerifyMembership calldata msg_)
        external
        notFrozen
        returns (uint256)
    {
        require(msg_.value.length > 0, EmptyValue());
        return membership(msg_.proof, msg_.proofHeight, msg_.path, msg_.value);
    }

    /// @inheritdoc ILightClient
    function verifyNonMembership(ILightClientMsgs.MsgVerifyNonMembership calldata msg_)
        external
        notFrozen
        returns (uint256)
    {
        return membership(msg_.proof, msg_.proofHeight, msg_.path, bytes(""));
    }

    /// @notice The entrypoint for verifying (non)membership proof.
    /// @dev This is a non-membership proof if the value is empty.
    /// @param proof The encoded proof.
    /// @param proofHeight The height of the proof.
    /// @param path The path of the key-value pair.
    /// @param value The value of the key-value pair.
    /// @return timestamp The timestamp of the trusted consensus state.
    function membership(
        bytes calldata proof,
        IICS02ClientMsgs.Height calldata proofHeight,
        bytes[] calldata path,
        bytes memory value
    )
        private
        returns (uint256 timestamp)
    {
        if (proof.length == 0) {
            // cached proof
            return getCachedKvPair(proofHeight.revisionHeight, IMembershipMsgs.KVPair(path, value));
        }

        IMembershipMsgs.MembershipProof memory membershipProof = abi.decode(proof, (IMembershipMsgs.MembershipProof));
        if (membershipProof.proofType == IMembershipMsgs.MembershipProofType.SP1MembershipProof) {
            return handleSP1MembershipProof(proofHeight, membershipProof.proof, path, value);
        } else if (membershipProof.proofType == IMembershipMsgs.MembershipProofType.SP1MembershipAndUpdateClientProof) {
            return handleSP1UpdateClientAndMembership(proofHeight, membershipProof.proof, path, value);
        }

        // unreachable
        revert UnknownMembershipProofType(uint8(membershipProof.proofType));
    }

    /// @notice The entrypoint for misbehaviour.
    /// @inheritdoc ILightClient
    function misbehaviour(bytes calldata misbehaviourMsg) public notFrozen {
        IMisbehaviourMsgs.MsgSubmitMisbehaviour memory msgSubmitMisbehaviour =
            abi.decode(misbehaviourMsg, (IMisbehaviourMsgs.MsgSubmitMisbehaviour));
        require(
            msgSubmitMisbehaviour.sp1Proof.vKey == MISBEHAVIOUR_PROGRAM_VKEY,
            VerificationKeyMismatch(MISBEHAVIOUR_PROGRAM_VKEY, msgSubmitMisbehaviour.sp1Proof.vKey)
        );

        IMisbehaviourMsgs.MisbehaviourOutput memory output =
            abi.decode(msgSubmitMisbehaviour.sp1Proof.publicValues, (IMisbehaviourMsgs.MisbehaviourOutput));

        validateMisbehaviourOutput(output);

        verifySP1Proof(msgSubmitMisbehaviour.sp1Proof);

        // If the misbehaviour and proof is valid, the client needs to be frozen
        clientState.isFrozen = true;
    }

    /// @notice The entrypoint for upgrading the client.
    /// @inheritdoc ILightClient
    function upgradeClient(bytes calldata) public view notFrozen {
        // TODO: Not yet implemented. (#78)
        revert FeatureNotSupported();
    }

    /// @notice Handles the `SP1MembershipProof` proof type.
    /// @param proofHeight The height of the proof.
    /// @param proofBytes The encoded proof.
    /// @param kvPath The path of the key-value pair.
    /// @param kvValue The value of the key-value pair.
    /// @return The timestamp of the trusted consensus state.
    function handleSP1MembershipProof(
        IICS02ClientMsgs.Height calldata proofHeight,
        bytes memory proofBytes,
        bytes[] calldata kvPath,
        bytes memory kvValue
    )
        private
        returns (uint256)
    {
        IMembershipMsgs.SP1MembershipProof memory proof = abi.decode(proofBytes, (IMembershipMsgs.SP1MembershipProof));
        require(
            proof.sp1Proof.vKey == MEMBERSHIP_PROGRAM_VKEY,
            VerificationKeyMismatch(MEMBERSHIP_PROGRAM_VKEY, proof.sp1Proof.vKey)
        );

        IMembershipMsgs.MembershipOutput memory output =
            abi.decode(proof.sp1Proof.publicValues, (IMembershipMsgs.MembershipOutput));
        require(
            output.kvPairs.length > 0 && output.kvPairs.length <= 256, LengthIsOutOfRange(output.kvPairs.length, 1, 256)
        );

        {
            // loop through the key-value pairs and validate them
            bool found = false;
            for (uint8 i = 0; i < output.kvPairs.length; i++) {
                if (!Paths.equal(output.kvPairs[i].path, kvPath)) {
                    continue;
                }

                bytes memory value = output.kvPairs[i].value;
                require(
                    value.length == kvValue.length && keccak256(value) == keccak256(kvValue),
                    MembershipProofValueMismatch(kvValue, value)
                );

                found = true;
                break;
            }
            require(found, MembershipProofKeyNotFound(kvPath));
        }

        validateMembershipOutput(output.commitmentRoot, proofHeight.revisionHeight, proof.trustedConsensusState);

        verifySP1Proof(proof.sp1Proof);

        // We avoid the cost of caching for single kv pairs, as reusing the proof is not necessary
        if (output.kvPairs.length > 1) {
            cacheKvPairs(proofHeight.revisionHeight, output.kvPairs, proof.trustedConsensusState.timestamp);
        }
        return proof.trustedConsensusState.timestamp;
    }

    /// @notice The entrypoint for handling the `SP1MembershipAndUpdateClientProof` proof type.
    /// @dev This function verifies the public values and forwards the proof to the SP1 verifier.
    /// @param proofHeight The height of the proof.
    /// @param proofBytes The encoded proof.
    /// @param kvPath The path of the key-value pair.
    /// @param kvValue The value of the key-value pair.
    /// @return The timestamp of the new consensus state.
    // solhint-disable-next-line code-complexity
    function handleSP1UpdateClientAndMembership(
        IICS02ClientMsgs.Height calldata proofHeight,
        bytes memory proofBytes,
        bytes[] calldata kvPath,
        bytes memory kvValue
    )
        private
        returns (uint256)
    {
        // validate proof and deserialize output
        IUpdateClientAndMembershipMsgs.UcAndMembershipOutput memory output;
        {
            IMembershipMsgs.SP1MembershipAndUpdateClientProof memory proof =
                abi.decode(proofBytes, (IMembershipMsgs.SP1MembershipAndUpdateClientProof));
            require(
                proof.sp1Proof.vKey == UPDATE_CLIENT_AND_MEMBERSHIP_PROGRAM_VKEY,
                VerificationKeyMismatch(UPDATE_CLIENT_AND_MEMBERSHIP_PROGRAM_VKEY, proof.sp1Proof.vKey)
            );

            output = abi.decode(proof.sp1Proof.publicValues, (IUpdateClientAndMembershipMsgs.UcAndMembershipOutput));
            require(
                output.kvPairs.length > 0 && output.kvPairs.length <= 256,
                LengthIsOutOfRange(output.kvPairs.length, 1, 256)
            );

            require(
                proofHeight.revisionHeight == output.updateClientOutput.newHeight.revisionHeight
                    && proofHeight.revisionNumber == output.updateClientOutput.newHeight.revisionNumber,
                ProofHeightMismatch(
                    proofHeight.revisionNumber,
                    proofHeight.revisionHeight,
                    output.updateClientOutput.newHeight.revisionNumber,
                    output.updateClientOutput.newHeight.revisionHeight
                )
            );

            validateUpdateClientPublicValues(output.updateClientOutput);

            verifySP1Proof(proof.sp1Proof);
        }

        // check update result
        {
            ILightClientMsgs.UpdateResult updateResult = checkUpdateResult(output.updateClientOutput);
            if (updateResult == ILightClientMsgs.UpdateResult.Update) {
                // adding the new consensus state to the mapping
                if (proofHeight.revisionHeight > clientState.latestHeight.revisionHeight) {
                    clientState.latestHeight = proofHeight;
                }
                consensusStateHashes[proofHeight.revisionHeight] =
                    keccak256(abi.encode(output.updateClientOutput.newConsensusState));
            } else if (updateResult == ILightClientMsgs.UpdateResult.Misbehaviour) {
                revert CannotHandleMisbehavior();
            } // else: NoOp
        }

        // loop through the key-value pairs and validate them
        {
            bool found = false;
            for (uint8 i = 0; i < output.kvPairs.length; i++) {
                if (!Paths.equal(output.kvPairs[i].path, kvPath)) {
                    continue;
                }

                bytes memory value = output.kvPairs[i].value;
                require(
                    value.length == kvValue.length && keccak256(value) == keccak256(kvValue),
                    MembershipProofValueMismatch(kvValue, value)
                );

                found = true;
                break;
            }
            require(found, MembershipProofKeyNotFound(kvPath));
        }

        validateMembershipOutput(
            output.updateClientOutput.newConsensusState.root,
            output.updateClientOutput.newHeight.revisionHeight,
            output.updateClientOutput.newConsensusState
        );

        // We avoid the cost of caching for single kv pairs, as reusing the proof is not necessary
        if (output.kvPairs.length > 1) {
            cacheKvPairs(
                proofHeight.revisionHeight, output.kvPairs, output.updateClientOutput.newConsensusState.timestamp
            );
        }
        return output.updateClientOutput.newConsensusState.timestamp;
    }

    /// @notice Validates the MembershipOutput public values.
    /// @param outputCommitmentRoot The commitment root of the output.
    /// @param proofHeight The height of the proof.
    /// @param trustedConsensusState The trusted consensus state
    function validateMembershipOutput(
        bytes32 outputCommitmentRoot,
        uint32 proofHeight,
        IICS07TendermintMsgs.ConsensusState memory trustedConsensusState
    )
        private
        view
    {
        bytes32 trustedConsensusStateHash = keccak256(abi.encode(trustedConsensusState));
        bytes32 storedConsensusStateHash = getConsensusStateHash(proofHeight);
        require(
            trustedConsensusStateHash == storedConsensusStateHash,
            ConsensusStateHashMismatch(storedConsensusStateHash, trustedConsensusStateHash)
        );

        require(
            outputCommitmentRoot == trustedConsensusState.root,
            ConsensusStateRootMismatch(trustedConsensusState.root, outputCommitmentRoot)
        );
    }

    /// @notice Validates the SP1ICS07UpdateClientOutput public values.
    /// @param output The public values.
    function validateUpdateClientPublicValues(IUpdateClientMsgs.UpdateClientOutput memory output) private view {
        validateClientStateAndTime(output.clientState, output.time);

        bytes32 outputConsensusStateHash = keccak256(abi.encode(output.trustedConsensusState));
        bytes32 storedConsensusStateHash = getConsensusStateHash(output.trustedHeight.revisionHeight);
        require(
            outputConsensusStateHash == storedConsensusStateHash,
            ConsensusStateHashMismatch(storedConsensusStateHash, outputConsensusStateHash)
        );
    }

    /// @notice Validates the SP1ICS07MisbehaviourOutput public values.
    /// @param output The public values.
    function validateMisbehaviourOutput(IMisbehaviourMsgs.MisbehaviourOutput memory output) private view {
        validateClientStateAndTime(output.clientState, output.time);

        // make sure the trusted consensus state from header 1 is known (trusted) by matching it with the one in the
        // mapping
        bytes32 outputConsensusStateHash1 = keccak256(abi.encode(output.trustedConsensusState1));
        bytes32 storedConsensusStateHash1 = getConsensusStateHash(output.trustedHeight1.revisionHeight);
        require(
            outputConsensusStateHash1 == storedConsensusStateHash1,
            ConsensusStateHashMismatch(storedConsensusStateHash1, outputConsensusStateHash1)
        );

        // make sure the trusted consensus state from header 2 is known (trusted) by matching it with the one in the
        // mapping
        bytes32 outputConsensusStateHash2 = keccak256(abi.encode(output.trustedConsensusState2));
        bytes32 storedConsensusStateHash2 = getConsensusStateHash(output.trustedHeight2.revisionHeight);
        require(
            outputConsensusStateHash2 == storedConsensusStateHash2,
            ConsensusStateHashMismatch(storedConsensusStateHash2, outputConsensusStateHash2)
        );
    }

    /// @notice Validates the client state and time.
    /// @dev This function does not check the equality of the latest height and isFrozen.
    /// @param publicClientState The public client state.
    /// @param time The time.
    function validateClientStateAndTime(
        IICS07TendermintMsgs.ClientState memory publicClientState,
        uint64 time
    )
        private
        view
    {
        require(time <= block.timestamp, ProofIsInTheFuture(block.timestamp, time));
        require(block.timestamp - time <= ALLOWED_SP1_CLOCK_DRIFT, ProofIsTooOld(block.timestamp, time));

        // Check client state equality
        // NOTE: We do not check the equality of latest height and isFrozen
        require(
            bytes(publicClientState.chainId).length == bytes(clientState.chainId).length
                && keccak256(bytes(publicClientState.chainId)) == keccak256(bytes(clientState.chainId)),
            ChainIdMismatch(clientState.chainId, publicClientState.chainId)
        );
        require(
            publicClientState.trustLevel.numerator == clientState.trustLevel.numerator
                && publicClientState.trustLevel.denominator == clientState.trustLevel.denominator,
            TrustThresholdMismatch(
                clientState.trustLevel.numerator,
                clientState.trustLevel.denominator,
                publicClientState.trustLevel.numerator,
                publicClientState.trustLevel.denominator
            )
        );
        require(
            publicClientState.trustingPeriod == clientState.trustingPeriod,
            TrustingPeriodMismatch(clientState.trustingPeriod, publicClientState.trustingPeriod)
        );
        require(
            publicClientState.unbondingPeriod == clientState.unbondingPeriod,
            UnbondingPeriodMismatch(clientState.unbondingPeriod, publicClientState.unbondingPeriod)
        );
    }

    /// @notice Checks for basic misbehaviour.
    /// @dev This function checks if the consensus state at the new height is different than the one in the mapping
    /// @dev or if the timestamp is not increasing.
    /// @dev If any of these conditions are met, it returns a Misbehaviour UpdateResult.
    /// @param output The public values of the update client program.
    /// @return The result of the update.
    function checkUpdateResult(IUpdateClientMsgs.UpdateClientOutput memory output)
        private
        view
        returns (ILightClientMsgs.UpdateResult)
    {
        bytes32 consensusStateHash = consensusStateHashes[output.newHeight.revisionHeight];
        if (consensusStateHash == bytes32(0)) {
            // No consensus state at the new height, so no misbehaviour
            return ILightClientMsgs.UpdateResult.Update;
        } else if (
            consensusStateHash != keccak256(abi.encode(output.newConsensusState))
                || output.trustedConsensusState.timestamp >= output.newConsensusState.timestamp
        ) {
            // The consensus state at the new height is different than the one in the mapping
            // or the timestamp is not increasing
            return ILightClientMsgs.UpdateResult.Misbehaviour;
        } else {
            // The consensus state at the new height is the same as the one in the mapping
            return ILightClientMsgs.UpdateResult.NoOp;
        }
    }

    /// @notice Verifies the SP1 proof
    /// @param proof The SP1 proof.
    function verifySP1Proof(ISP1Msgs.SP1Proof memory proof) private view {
        VERIFIER.verifyProof(proof.vKey, proof.publicValues, proof.proof);
    }

    /// @notice Caches the key-value pairs to the transient storage with the timestamp.
    /// @param proofHeight The height of the proof.
    /// @param kvPairs The key-value pairs.
    /// @param timestamp The timestamp of the trusted consensus state.
    /// @dev WARNING: Transient store is not reverted even if a message within a transaction reverts.
    /// @dev WARNING: This function must be called after all proof and validation checks.
    function cacheKvPairs(uint32 proofHeight, IMembershipMsgs.KVPair[] memory kvPairs, uint256 timestamp) private {
        for (uint8 i = 0; i < kvPairs.length; i++) {
            bytes32 kvPairHash = keccak256(abi.encode(proofHeight, kvPairs[i]));
            kvPairHash.asUint256().tstore(timestamp);
        }
    }

    /// @notice Gets the timestamp of the cached key-value pair from the transient storage.
    /// @param proofHeight The height of the proof.
    /// @param kvPair The key-value pair.
    /// @return The timestamp of the cached key-value pair.
    function getCachedKvPair(uint32 proofHeight, IMembershipMsgs.KVPair memory kvPair) private view returns (uint256) {
        bytes32 kvPairHash = keccak256(abi.encode(proofHeight, kvPair));
        uint256 timestamp = kvPairHash.asUint256().tload();
        require(timestamp != 0, KeyValuePairNotInCache(kvPair.path, kvPair.value));
        return timestamp;
    }

    /// @notice Modifier to check if the client is not frozen.
    modifier notFrozen() {
        require(!clientState.isFrozen, FrozenClientState());
        _;
    }
}
