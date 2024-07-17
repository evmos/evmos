// SPDX-License-Identifier: LGPL-3.0-only

pragma solidity >=0.7.0 <0.9.0;

contract Counter {
    uint256 public counter = 1;

    function increment() external {
        counter++;
    }

    function decrement() external {
        counter--;
    }
}
