// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IIBCPausable {
    /// @notice The role identifier for the pauser role
    function PAUSER_ROLE() external view returns (bytes32);

    /// @notice Pauses the contract
    /// @dev The caller must have the pauser role
    function pause() external;

    /// @notice Unpauses the contract
    /// @dev The caller must have the pauser role
    function unpause() external;

    /// @notice Grants the pauser role to an account
    /// @dev The caller must be authorized by the derived contract
    /// @param account The account to grant the role to
    function grantPauserRole(address account) external;

    /// @notice Revokes the pauser role from an account
    /// @dev The caller must be authorized by the derived contract
    /// @param account The account to revoke the role from
    function revokePauserRole(address account) external;
}
