// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/*
    @title tokenTransfer
    @dev This contract is used to test that any addresses transferring ERC-20 tokens
    are tracked if they're ERC-20 representations of native Cosmos coins.
*/
contract tokenTransfer {
    ERC20 token;

    constructor(address tokenAddress){
        token = ERC20(tokenAddress);
    }

    /*
        @notice This function is used to transfer ERC-20 tokens to a given address.
        @param to The address to transfer the tokens to
        @param amount The amount of tokens to transfer
    */
    function transferToken(address to, uint256 amount) public {
        token.transfer(to, amount);
    }
}
