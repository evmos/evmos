// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.17;

import "../IWERC20.sol" as werc20Precompile;

contract WEVMOS9TestCaller {
    werc20Precompile.IWERC20 public wrappedToken;
    uint256 public counter;

    constructor(address payable _wrappedTokenAddress) {
        wrappedToken = werc20Precompile.IWERC20(_wrappedTokenAddress);
        counter = 0;
    }

    function depositWithRevert(bool before, bool aft) public payable {
        counter++;
        require(msg.value > 0, "No Ether sent");
        wrappedToken.deposit{value: msg.value}();

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
