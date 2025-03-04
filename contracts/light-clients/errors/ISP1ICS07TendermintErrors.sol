// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface ISP1ICS07TendermintErrors {
    /// @notice The error that is returned when the verification key does not match the expected value.
    /// @param expected The expected verification key.
    /// @param actual The actual verification key.
    error VerificationKeyMismatch(bytes32 expected, bytes32 actual);

    /// @notice The error that is returned when the client state is frozen.
    error FrozenClientState();

    /// @notice The error that is returned when a proof is in the future.
    /// @param now The current timestamp in seconds.
    /// @param proofTimestamp The timestamp in the proof in seconds.
    error ProofIsInTheFuture(uint256 now, uint256 proofTimestamp);

    /// @notice The error that is returned when a proof is too old.
    /// @param now The current timestamp in seconds.
    /// @param proofTimestamp The timestamp in the proof in seconds.
    error ProofIsTooOld(uint256 now, uint256 proofTimestamp);

    /// @notice The error that is returned when the chain ID does not match the expected value.
    /// @param expected The expected chain ID.
    /// @param actual The actual chain ID.
    error ChainIdMismatch(string expected, string actual);

    /// @notice The error that is returned when the trust threshold does not match the expected value.
    /// @param expectedNumerator The expected numerator of the trust threshold.
    /// @param expectedDenominator The expected denominator of the trust threshold.
    /// @param actualNumerator The actual numerator of the trust threshold.
    /// @param actualDenominator The actual denominator of the trust threshold.
    error TrustThresholdMismatch(
        uint256 expectedNumerator, uint256 expectedDenominator, uint256 actualNumerator, uint256 actualDenominator
    );

    /// @notice The error that is returned when the trusting period does not match the expected value.
    /// @param expected The expected trusting period in seconds.
    /// @param actual The actual trusting period in seconds.
    error TrustingPeriodMismatch(uint256 expected, uint256 actual);

    /// @notice The error that is returned when the unbonding period does not match the expected value.
    /// @param expected The expected unbonding period in seconds.
    /// @param actual The actual unbonding period in seconds.
    error UnbondingPeriodMismatch(uint256 expected, uint256 actual);

    /// @notice The error that is returned when the trusting period is longer than the unbonding period.
    /// @param trustingPeriod The trusting period in seconds.
    /// @param unbondingPeriod The unbonding period in seconds.
    error TrustingPeriodTooLong(uint256 trustingPeriod, uint256 unbondingPeriod);

    /// @notice The error that is returned when the consensus state hash does not match the expected value.
    /// @param expected The expected consensus state hash.
    /// @param actual The actual consensus state hash.
    error ConsensusStateHashMismatch(bytes32 expected, bytes32 actual);

    /// @notice The error that is returned when the consensus state is not found.
    error ConsensusStateNotFound();

    /// @notice The error that is returned when the length of a value is out of range.
    /// @param length The length of the value.
    /// @param min The minimum length of the value.
    /// @param max The maximum length of the value.
    error LengthIsOutOfRange(uint256 length, uint256 min, uint256 max);

    /// @notice The error that is returned when the key-value pair's value does not match the expected value.
    /// @param expected The expected value.
    /// @param actual The actual value.
    error MembershipProofValueMismatch(bytes expected, bytes actual);

    /// @notice The error that is returned when the key-value pair's path is not contained in the proof.
    /// @param path The path of the key-value pair.
    error MembershipProofKeyNotFound(bytes[] path);

    /// @notice The error that is returned when the consensus state root does not match the expected value.
    /// @param expected The expected consensus state root.
    /// @param actual The actual consensus state root.
    error ConsensusStateRootMismatch(bytes32 expected, bytes32 actual);

    /// @notice The error that is returned when the client state does not match the expected value.
    /// @param expected The expected client state.
    /// @param actual The actual client state.
    error ClientStateMismatch(bytes expected, bytes actual);

    /// @notice The error that is returned when the update client and membership program contains misbehavior.
    /// @dev Misbehavior cannot be handled in membership handler, so it is returned as an error.
    error CannotHandleMisbehavior();

    /// @notice The error that is returned when the proof height does not match the expected value.
    /// @param expectedRevisionNumber The expected revision number.
    /// @param expectedRevisionHeight The expected revision height.
    /// @param actualRevisionNumber The actual revision number.
    /// @param actualRevisionHeight The actual revision height.
    error ProofHeightMismatch(
        uint64 expectedRevisionNumber,
        uint64 expectedRevisionHeight,
        uint64 actualRevisionNumber,
        uint64 actualRevisionHeight
    );

    /// @notice The error that is returned when the membership proof type is unknown.
    /// @param proofType The unknown membership proof type.
    error UnknownMembershipProofType(uint8 proofType);

    /// @notice The error that is returned when the zk algorithm is unknown.
    /// @param algorithm The unknown zk algorithm.
    error UnknownZkAlgorithm(uint8 algorithm);

    /// @notice Returned when the feature is not supported.
    error FeatureNotSupported();

    /// @notice Returned when the membership proof is invalid.
    error InvalidMembershipProof();

    /// @notice Returned when a key-value pair is not in the cache.
    /// @param path The path of the key-value pair.
    /// @param value The value of the key-value pair.
    error KeyValuePairNotInCache(bytes[] path, bytes value);

    /// @notice Returned when the membership value is empty.
    error EmptyValue();
}
