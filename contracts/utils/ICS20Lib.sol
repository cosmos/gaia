// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.28;

// solhint-disable no-inline-assembly

import { Strings } from "@openzeppelin-contracts/utils/Strings.sol";
import { Bytes } from "@openzeppelin-contracts/utils/Bytes.sol";
import { IICS20Errors } from "../errors/IICS20Errors.sol";

// This library was originally copied, with minor adjustments, from https://github.com/hyperledger-labs/yui-ibc-solidity
// It has since been modified heavily (e.g. replacing JSON with ABI encoding, adding new functions, etc.)
library ICS20Lib {
    /// @notice ICS20_VERSION is the version string for ICS20 packet data.
    string internal constant ICS20_VERSION = "ics20-1";

    /// @notice ICS20_ENCODING is the encoding string for ICS20 packet data.
    string internal constant ICS20_ENCODING = "application/x-solidity-abi";

    /// @notice IBC_DENOM_PREFIX is the prefix for IBC denoms.
    string internal constant IBC_DENOM_PREFIX = "ibc/";

    /// @notice DEFAULT_PORT_ID is the default port id for ICS20.
    string internal constant DEFAULT_PORT_ID = "transfer";

    /// @notice SUCCESSFUL_ACKNOWLEDGEMENT_JSON is the JSON bytes for a successful acknowledgement.
    bytes internal constant SUCCESSFUL_ACKNOWLEDGEMENT_JSON = bytes("{\"result\":\"AQ==\"}");

    /// @notice KECCAK256_ICS20_VERSION is the keccak256 hash of the ICS20_VERSION.
    bytes32 internal constant KECCAK256_ICS20_VERSION = keccak256(bytes(ICS20_VERSION));

    /// @notice KECCAK256_ICS20_ENCODING is the keccak256 hash of the ICS20_ENCODING.
    bytes32 internal constant KECCAK256_ICS20_ENCODING = keccak256(bytes(ICS20_ENCODING));

    /// @notice KECCAK256_DEFAULT_PORT_ID is the keccak256 hash of the DEFAULT_PORT_ID.
    bytes32 internal constant KECCAK256_DEFAULT_PORT_ID = keccak256(bytes(DEFAULT_PORT_ID));

    /// @notice mustHexStringToAddress converts a hex string to an address and reverts on failure.
    /// @param addrHexString hex address string
    /// @return address the converted address
    function mustHexStringToAddress(string memory addrHexString) internal pure returns (address) {
        (bool success, address addr) = Strings.tryParseAddress(addrHexString);
        require(success, IICS20Errors.ICS20InvalidAddress(addrHexString));
        return addr;
    }

    /// @notice hasPrefix checks a denom for a prefix
    /// @param denomBz the denom to check
    /// @param prefix the prefix to check with
    /// @return true if `denomBz` has the prefix `prefix`
    function hasPrefix(bytes memory denomBz, bytes memory prefix) internal pure returns (bool) {
        if (denomBz.length < prefix.length) {
            return false;
        }
        return keccak256(Bytes.slice(denomBz, 0, prefix.length)) == keccak256(prefix);
    }

    /// @notice getDenomPrefix returns an ibc path prefix
    /// @param portId Port
    /// @param clientId client
    /// @return Denom prefix
    function getDenomPrefix(string memory portId, string calldata clientId) internal pure returns (bytes memory) {
        return abi.encodePacked(portId, "/", clientId, "/");
    }

    /// @notice hasHops checks if a denom has any hops in it (i.e it has a "/" in it).
    /// @param denom Denom to check
    /// @return true if the denom has any hops in it
    function hasHops(bytes memory denom) internal pure returns (bool) {
        // check if the denom has any '/' in it
        for (uint256 i = 0; i < denom.length; i++) {
            if (denom[i] == "/") {
                return true;
            }
        }

        return false;
    }
}
