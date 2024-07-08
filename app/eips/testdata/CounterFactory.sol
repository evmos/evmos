// SPDX-License-Identifier: LGPL-3.0-only

pragma solidity >=0.7.0 <0.9.0;

import "./Counter.sol";

contract Counterfactory {
    Counter public counterInstance;

    constructor() {
        counterInstance = new Counter();
    }

    function incrementCounter() public {
        counterInstance.increment();
    }

    function decrementCounter() public {
        counterInstance.decrement();
    }

    function getCounterValue() public view returns (uint256) {
        return counterInstance.counter();
    }
}
