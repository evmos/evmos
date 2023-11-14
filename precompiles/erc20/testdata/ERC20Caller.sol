// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.20;

import "../IERC20.sol" as erc20;

/// @title ERC20Caller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the ERC20 precompile.
contract ERC20Caller {
    erc20.IERC20 public token;

    constructor(address tokenAddress) {
        token = erc20.IERC20(tokenAddress);
    }

    function transfer(address to, uint256 amount) external {
        token.transfer(to, amount);
    }

    function transferFrom(address from, address to, uint256 amount) external {
        token.transferFrom(from, to, amount);
    }

    function approve(address spender, uint256 amount) external {
        token.approve(spender, amount);
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
}
