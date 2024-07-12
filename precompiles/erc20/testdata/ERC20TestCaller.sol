// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../IERC20.sol" as erc20Precompile;

/// @title ERC20TestCaller
/// @author Evmos Core Team
/// @dev This contract is used to test external contract calls to the ERC20 precompile.
contract ERC20TestCaller {
    erc20Precompile.IERC20 public token;
    uint256 public counter;

    constructor(address tokenAddress) {
        token = erc20Precompile.IERC20(tokenAddress);
        counter = 0;
    }

    function transferWithRevert(
        address to,
        uint256 amount,
        bool before,
        bool aft
    ) public returns (bool) {
        counter++;
        
        bool res = token.transfer(to, amount);
        if (before) {
            require(false, "revert here");
        }
        counter--;
        
        if (aft) {
            require(false, "revert here");
        }
        return res;
    }
}
