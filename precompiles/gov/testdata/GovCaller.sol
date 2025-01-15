/// SPDX-License-Identifier: LGPL-3.0-only

pragma solidity >=0.8.18;

import "../IGov.sol";

contract GovCaller {
    function getParams() external view returns (Params memory params) {
        return GOV_CONTRACT.getParams();
    }
}
