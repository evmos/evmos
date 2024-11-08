// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../IWERC20.sol";

contract WEVMOS9TestCaller {
    address payable public immutable WEVMOS;
    uint256 public counter;

    constructor(address payable _wrappedTokenAddress) {
        WEVMOS = _wrappedTokenAddress;
        counter = 0;
    }

    event Log(string message);

    function depositWithRevert(bool before, bool aft) public payable {
        counter++;

        uint amountIn = msg.value;
        IWERC20(WEVMOS).deposit{value: amountIn}();

        if (before) {
            require(false, "revert here");
        }

        counter--;

        if (aft) {
            require(false, "revert here");
        }
        return;
    }
}
