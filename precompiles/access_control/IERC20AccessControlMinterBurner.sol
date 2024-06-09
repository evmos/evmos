// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.20;

import "./IAccessControl.sol";
import {IERC20MetadataAllowance} from "../erc20/IERC20MetadataAllowance.sol";

/// @author Evmos Team
// @title IERC20AccessControlMinterBurner
// @dev The interface for the ERC20 token with access control, minting and burning capabilities.
interface IERC20AccessControlMinterBurner is IAccessControl, IERC20MetadataAllowance {

    // @dev Emitted when `value` tokens are burned from `from`.
    // @param from The address from which the tokens are burned.
    // @param value The amount of tokens burned.
    event Burn(address indexed from, uint256 value);

    // @dev Emitted when `value` tokens are minted to `to`.
    // @param to The address to which the tokens are minted.
    // @param value The amount of tokens minted.
    event Mint(address indexed to, uint256 value);

    // @dev Burns `value` tokens from the caller.
    // @param value The amount of token to be burned.
    function burn(uint256 value) external;

    // @dev Mints `amount` tokens to `to`.
    // @param to The address to mint tokens to.
    // @param amount The amount of tokens to mint.
    function mint(address to, uint256 amount) external;
}

