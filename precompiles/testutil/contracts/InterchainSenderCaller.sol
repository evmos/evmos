// SPDX-License-Identifier: LGPL-v3
pragma solidity >=0.8.17;

import "./InterchainSender.sol";
import "./../../ics20/ICS20I.sol";
import "./../../common/Types.sol";

contract InterchainSenderCaller {
    int64 public counter;
    InterchainSender interchainSender;

    constructor(address _interchainSender) payable {
        interchainSender = InterchainSender(_interchainSender);
    }

    /// @dev This function will perform 2 calls to the testMultiTransferWithInternalTransfer
    /// @dev function of the InterchainSender contract
    /// @dev However, the second call will be reverted,
    /// @dev so in total should perform 2 IBC transfers
    function transfersWithRevert(
        address payable _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        counter++;
        interchainSender.testMultiTransferWithInternalTransfer(
            _source,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            receiver,
            true,
            true,
            true
        );
        try
            InterchainSenderCaller(address(this))
                .performMultiTransferWithRevert(
                    _source,
                    sourcePort,
                    sourceChannel,
                    denom,
                    amount,
                    receiver
                )
        {} catch {}
        counter++;
    }

    /// @dev This function will perform 2 calls to the testMultiTransferWithInternalTransfer
    /// @dev function of the InterchainSender contract
    /// @dev However, the second call will be reverted,
    /// @dev and this should revert all 4 IBC transfers
    function transfersWithNestedRevert(
        address payable _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        counter++;
        try
            InterchainSenderCaller(address(this)).performNestedTransfers(
                _source,
                sourcePort,
                sourceChannel,
                denom,
                amount,
                receiver
            )
        {} catch {}
        counter++;
    }

    function performNestedTransfers(
        address payable _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        interchainSender.testMultiTransferWithInternalTransfer(
            _source,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            receiver,
            true,
            true,
            true
        );
        InterchainSenderCaller(address(this)).performMultiTransferWithRevert(
            _source,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            receiver
        );
    }

    function performMultiTransferWithRevert(
        address payable _source,
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver
    ) external {
        interchainSender.testMultiTransferWithInternalTransfer(
            _source,
            sourcePort,
            sourceChannel,
            denom,
            amount,
            receiver,
            true,
            true,
            true
        );
        revert();
    }
}
