// SPDX-License-Identifier: MIT
pragma solidity >0.8.0;

contract TestChainID {
    function currentChainID() public view returns (uint) {
        uint id;
        assembly {
            id := chainid()
        }
        return id;
    }
}

