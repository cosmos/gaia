// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

interface IIBCERC20 {
    /// @notice Mint new tokens to the Escrow contract
    /// @dev This function can only be called by the ICS20 contract
    /// @dev This function can only mint tokens to the Escrow contract
    /// @param mintAddress Address to mint tokens to
    /// @param amount Amount of tokens to mint
    function mint(address mintAddress, uint256 amount) external;

    /// @notice Burn tokens from the Escrow contract
    /// @dev This function can only be called by the ICS20 contract
    /// @dev This function can only burn tokens from the Escrow contract
    /// @param mintAddress Address to burn tokens from
    /// @param amount Amount of tokens to burn
    function burn(address mintAddress, uint256 amount) external;

    /// @notice Get the full denom path of the token
    /// @return the full path of the token's denom
    function fullDenomPath() external view returns (string memory);

    /// @notice Get the escrow contract address
    /// @return the escrow contract address
    function escrow() external view returns (address);

    /// @notice Get the ICS20 contract address
    /// @return the ICS20 contract address
    function ics20() external view returns (address);

    /// @notice Initializes the IBCERC20 contract
    /// @dev This function is meant to be called by a proxy
    /// @param ics20_ The ICS20 contract address
    /// @param escrow_ The escrow contract address, can burn and mint tokens
    /// @param fullDenomPath_ The full IBC denom path for this token
    function initialize(address ics20_, address escrow_, string memory fullDenomPath_) external;
}
