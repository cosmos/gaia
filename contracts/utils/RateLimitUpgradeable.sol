// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import { AccessControlUpgradeable } from "@openzeppelin-upgradeable/access/AccessControlUpgradeable.sol";
import { IRateLimitErrors } from "../errors/IRateLimitErrors.sol";
import { IRateLimit } from "../interfaces/IRateLimit.sol";

/// @title Rate Limit Upgradeable contract
/// @notice This contract is an abstract contract for adding rate limiting to escrow contracts.
/// @dev Rate limits are set per token address by the rate limiter role and are enforced per day.
/// @dev Rate limits are applied to tokens leaving the escrow contract.
abstract contract RateLimitUpgradeable is IRateLimitErrors, IRateLimit, AccessControlUpgradeable {
    /// @notice Storage of the RateLimit contract
    /// @dev It's implemented on a custom ERC-7201 namespace to reduce the risk of storage collisions when using with
    /// upgradeable contracts.
    /// @param _rateLimits Mapping of token addresses to their rate limits, 0 means no limit
    /// @param _dailyUsage Mapping of daily token keys to their usage
    struct RateLimitStorage {
        mapping(address token => uint256 limit) _rateLimits;
        mapping(bytes32 dailyTokenKey => uint256 usage) _dailyUsage;
    }

    /// @notice ERC-7201 slot for the RateLimit storage
    /// @dev keccak256(abi.encode(uint256(keccak256("ibc.storage.RateLimit")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant RATELIMIT_STORAGE_SLOT = 0xcb05b6cb8e6c87c443cb04d44193d7d46d51c1198725a0ee3478d5baa736c100;

    /// @inheritdoc IRateLimit
    bytes32 public constant RATE_LIMITER_ROLE = keccak256("RATE_LIMITER_ROLE");

    /// @notice The period for rate limiting
    uint256 private constant RATE_LIMIT_PERIOD = 1 days;

    /// @notice The initializer for the RateLimit contract
    function __RateLimit_init() internal onlyInitializing {
        __AccessControl_init();
    }

    /// @inheritdoc IRateLimit
    function setRateLimit(address token, uint256 rateLimit) external onlyRole(RATE_LIMITER_ROLE) {
        _getRateLimitStorage()._rateLimits[token] = rateLimit;
    }

    /// @inheritdoc IRateLimit
    function getRateLimit(address token) external view returns (uint256) {
        return _getRateLimitStorage()._rateLimits[token];
    }

    /// @inheritdoc IRateLimit
    function getDailyUsage(address token) external view returns (uint256) {
        return _getRateLimitStorage()._dailyUsage[_getDailyTokenKey(token)];
    }

    /// @notice Checks the rate limit for a token and updates the daily usage
    /// @param token The token address
    /// @param amount The amount to check against the rate limit
    function _assertAndUpdateRateLimit(address token, uint256 amount) internal {
        RateLimitStorage storage $ = _getRateLimitStorage();

        uint256 rateLimit = $._rateLimits[token];
        if (rateLimit == 0) {
            return;
        }

        bytes32 dailyTokenKey = _getDailyTokenKey(token);
        uint256 usage = $._dailyUsage[dailyTokenKey] + amount;
        require(usage <= rateLimit, RateLimitExceeded(rateLimit, usage));

        $._dailyUsage[dailyTokenKey] = usage;
    }

    /// @notice Reduces the daily usage for a token
    /// @dev This function is used in order to track the net usage a token
    /// @param token The token address
    /// @param amount The amount to reduce from the daily usage
    function _reduceDailyUsage(address token, uint256 amount) internal {
        RateLimitStorage storage $ = _getRateLimitStorage();

        uint256 rateLimit = $._rateLimits[token];
        if (rateLimit == 0) {
            return;
        }

        bytes32 dailyTokenKey = _getDailyTokenKey(token);
        uint256 usage = $._dailyUsage[dailyTokenKey];
        if (usage <= amount) {
            $._dailyUsage[dailyTokenKey] = 0;
        } else {
            $._dailyUsage[dailyTokenKey] = usage - amount;
        }
    }

    /// @inheritdoc IRateLimit
    function grantRateLimiterRole(address account) external {
        _authorizeSetRateLimiterRole(account);
        _grantRole(RATE_LIMITER_ROLE, account);
    }

    /// @inheritdoc IRateLimit
    function revokeRateLimiterRole(address account) external {
        _authorizeSetRateLimiterRole(account);
        _revokeRole(RATE_LIMITER_ROLE, account);
    }

    /// @notice Returns the daily token key for the current timestamp and token
    /// @param token The token address
    /// @return The daily token key
    function _getDailyTokenKey(address token) internal view returns (bytes32) {
        return keccak256(abi.encodePacked(block.timestamp / RATE_LIMIT_PERIOD, token));
    }

    /// @notice Authorizes the granting or revoking of the rate limiter role
    /// @param account The account to authorize
    function _authorizeSetRateLimiterRole(address account) internal virtual;

    /// @notice Returns the storage of the RateLimit contract
    function _getRateLimitStorage() internal pure returns (RateLimitStorage storage $) {
        // solhint-disable-next-line no-inline-assembly
        assembly {
            $.slot := RATELIMIT_STORAGE_SLOT
        }
    }
}
