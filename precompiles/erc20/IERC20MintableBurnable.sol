// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

/**
 * @title IERC20MintableBurnable Interface
 * @dev Interface of the ERC20 standard as defined in the EIP.
 */
interface IERC20MintableBurnable {
    /**
     * @notice Function to mint new tokens.
     * @dev Can only be called by the minter address.
     * @param to The address that will receive the minted tokens.
     * @param amount The amount of tokens to mint.
     */
    function mint(address to, uint256 amount) external;

    /**
     * @notice Function to burn tokens.
     * @dev Can only be called by the minter address.
     * @param from The address that will have its tokens burnt.
     * @param amount The amount of tokens to burn.
     */
    function burn(address from, uint256 amount) external;
}
