// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../IERC20MetadataAllowance.sol" as erc20Allowance;

/// @title ERC20AllowanceCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the ERC20 precompile.
contract ERC20AllowanceCaller {
    erc20Allowance.IERC20MetadataAllowance public token;

    constructor(address tokenAddress) {
        token = erc20Allowance.IERC20MetadataAllowance(tokenAddress);
    }

    function transfer(address to, uint256 amount) external {
        bool success = token.transfer(to, amount);
        require(success, "ERC20Caller: transfer failed");
    }

    function transferFrom(address from, address to, uint256 amount) external {
        bool success = token.transferFrom(from, to, amount);
        require(success, "ERC20Caller: transferFrom failed");
    }

    function approve(address spender, uint256 amount) external {
        bool success = token.approve(spender, amount);
        require(success, "ERC20Caller: approve failed");
    }

    function allowance(address owner, address spender) external view returns (uint256) {
        return token.allowance(owner, spender);
    }

    function balanceOf(address owner) external view returns (uint256) {
        return token.balanceOf(owner);
    }

    function totalSupply() external view returns (uint256) {
        return token.totalSupply();
    }

    function name() external view returns (string memory) {
        return token.name();
    }

    function symbol() external view returns (string memory) {
        return token.symbol();
    }

    function decimals() external view returns (uint8) {
        return token.decimals();
    }

    function increaseAllowance(address spender, uint256 addedValue) external {
        bool success = token.increaseAllowance(spender, addedValue);
        require(success, "ERC20Caller: increaseAllowance failed");
    }

    function decreaseAllowance(address spender, uint256 subtractedValue) external {
        bool success = token.decreaseAllowance(spender, subtractedValue);
        require(success, "ERC20Caller: decreaseAllowance failed");
    }
}
