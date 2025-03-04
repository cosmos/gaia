// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IIBCUUPSUpgradeable {
    /// @notice Returns the timelocked admin address
    /// @return The timelocked admin address
    function getTimelockedAdmin() external view returns (address);
    /// @notice Returns the governance admin address
    /// @return The governance admin address, 0 if not set
    function getGovAdmin() external view returns (address);
    /// @notice Sets the timelocked admin address
    /// @dev Either admin can set the timelocked admin address.
    /// @param newTimelockedAdmin The new timelocked admin address
    function setTimelockedAdmin(address newTimelockedAdmin) external;
    /// @notice Sets the governance admin address
    /// @dev Either admin can set the governance admin address.
    /// @dev Since timelocked admin is timelocked, this operation can be stopped by the govAdmin.
    /// @param newGovAdmin The new governance admin address
    function setGovAdmin(address newGovAdmin) external;
    /// @notice Returns true if the account is an admin
    /// @dev Used by other IBC contracts to check if upgrades are authorized
    /// @param account The account to check
    /// @return True if the account is an admin
    function isAdmin(address account) external view returns (bool);
}
