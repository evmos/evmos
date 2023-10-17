// SPDX-License-Identifier: MIT
// NOTE: don't change the compiler version because
// test may fail due to changes in deploy contract transaction data (gas, input & output fields)
pragma solidity 0.8.18;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract TestERC20A is ERC20 {
    constructor() ERC20("TestERC20", "Test") {
        _mint(msg.sender, 100000000000000000000000000);
    }
}
