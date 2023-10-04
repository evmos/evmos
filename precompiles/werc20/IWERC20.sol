// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;

import "./../erc20/IERC20MetadataAllowance.sol";

/**
 * @author Evmos Team
 * @title Wrapped ERC20 Interface
 * @dev Interface for the wrapped native token as ERC20 standard.
 * This defines the interface for WEVMOS the wrapped ERC20 token of EVMOS.
 */
interface IWERC20 is IERC20MetadataAllowance {
		/// @dev Emitted when ERC20 WEVMOS tokens are deposited in exchange for native EVMOS.
    /// @param dst The account for which the deposit is made.
    /// @param wad The amount of ERC20 WEVMOS tokens deposited.
    event Deposit(address indexed dst, uint wad);

    /// @dev Emitted when native EVMOS is deposited in exchange for ERC20 WEVMOS tokens.
    /// @param src The account for which the withdrawal is made.
    /// @param wad The amount of native EVMOS coins withdrawn.
    event Withdrawal(address indexed src, uint wad);

		/// @dev Default fallback payable function which will serve as deposit function.
    fallback() external payable;

    /// @dev Deposits native EVMOS coins in exchange for wrapped ERC20 token.
    /// @dev After execution of this function the SetBalance function
    /// @dev burns the EVMOS coins and increases the contract balance of the ERC20 tokens.
    /// @dev Emits a Deposit Event.
    function deposit() external payable;

    /// @dev Withdraws native EVMOS coins in exchange for wrapped ERC20 token.
    /// @dev After execution of this function the SetBalance function
    /// @dev decreases the contract balance of the ERC20 tokens and mints the EVMOS coins.
    /// @dev Emits a Withdrawal Event.
    /// @param wad The amount of EVMOS coins to be withdrawn.
    function withdraw(uint wad) external;
}