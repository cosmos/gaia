// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IICS02ClientErrors {
    /// @notice Invalid client type
    /// @param clientType client type
    error IBCInvalidClientType(string clientType);

    /// @notice Client not found
    /// @param clientId client identifier
    error IBCClientNotFound(string clientId);

    /// @notice Counterparty client not found
    /// @param counterpartyClientId counterparty client identifier
    error IBCCounterpartyClientNotFound(string counterpartyClientId);

    /// @notice Unreachable code
    error Unreachable();
}
