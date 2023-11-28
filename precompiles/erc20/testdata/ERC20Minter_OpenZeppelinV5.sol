// SPDX-License-Identifier: MIT
//
// Based on OpenZeppelin Contracts v5.0.0 (token/ERC20/ERC20.sol)
//
// NOTE: This was compiled using REMIX IDE.

pragma solidity ^0.8.20;

import "https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.0.0/contracts/token/ERC20/ERC20.sol";

/**
 * @dev {ERC20} token, including:
 *
 *  - ability to mint tokens
 *
 * ATTENTION: This contract does not restrict minting tokens to any particular address
 * and should thus ONLY BE USED FOR TESTING.
 */
contract ERC20Minter_OpenZeppelinV5 is ERC20 {
    constructor(string memory name, string memory symbol)
    ERC20(name, symbol) {}

    function mint(address to, uint256 amount) public {
        _mint(to, amount);
    }
}
