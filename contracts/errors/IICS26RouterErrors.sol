// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IICS26RouterErrors {
    /// @notice IBC port identifier already exists
    /// @param portId port identifier
    error IBCPortAlreadyExists(string portId);

    /// @notice IBC invalid port identifier
    /// @param portId port identifier
    error IBCInvalidPortIdentifier(string portId);

    /// @notice IBC invalid timeout timestamp
    /// @param timeoutTimestamp packet's timeout timestamp in seconds
    /// @param comparedTimestamp compared timestamp in seconds
    error IBCInvalidTimeoutTimestamp(uint256 timeoutTimestamp, uint256 comparedTimestamp);

    /// @notice IBC timeout period too long
    /// @param maxTimeoutDuration maximum timeout period in seconds
    /// @param actualTimeoutDuration actual timeout period in seconds
    error IBCInvalidTimeoutDuration(uint256 maxTimeoutDuration, uint256 actualTimeoutDuration);

    /// @notice IBC unexpected counterparty identifier
    /// @param expected expected counterparty identifier
    /// @param actual actual counterparty identifier
    error IBCInvalidCounterparty(string expected, string actual);

    /// @notice IBC async acknowledgement not supported
    error IBCAsyncAcknowledgementNotSupported();

    /// @notice IBC application cannot return the universal error acknowledgement
    error IBCErrorUniversalAcknowledgement();

    /// @notice IBC app for port not found
    /// @param portId port identifier
    error IBCAppNotFound(string portId);

    /// @notice IBC unauthorized packet sender
    /// @param caller unauthorized sender address
    error IBCUnauthorizedSender(address caller);

    /// @notice IBC callback failed due to unknown reason
    /// @dev Usually OOG
    error IBCFailedCallback();
}
