// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.8.18;

import "./evmos/ics20/ICS20I.sol";
import "./evmos/common/Types.sol";


contract ICS20FromContract {
    int64 public counter;

    function balanceOfContract() public view returns (uint256) {
        return address(this).balance;
    }

    function deposit() public payable {}

    function transfer(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        counter += 1;
        Height memory timeoutHeight =  Height(100, 100);
        ICS20_CONTRACT.transfer(
            sourcePort,
            sourceChannel,
            denom,
            amount,
            address(this),
            receiver,
            timeoutHeight,
            0,
            ""
        );
        counter -= 1;
    }

    function transferFromEOA(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        counter += 1;
        Height memory timeoutHeight =  Height(100, 100);
        ICS20_CONTRACT.transfer(
            sourcePort,
            sourceChannel,
            denom,
            amount,
            msg.sender,
            receiver,
            timeoutHeight,
            0,
            ""
        );
        counter -= 1;
    }
}
