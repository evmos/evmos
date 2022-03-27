// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

interface IICS20Transfer is IModuleCallbacks {

    function sendTransfer(
        string calldata denom,
        uint64 amount,
        address receiver,
        string calldata sourcePort,
        string calldata sourceChannel,
        uint64 timeoutHeight
    ) external;

    event Transfer(
        string indexed denom,
        uint64 indexed amount,
        address indexed sender,
        address indexed receiver
    );

    event recieveTransfer(
        string indexed denom,
        uint64 indexed amount,
        address indexed sender,
        address indexed receiver
    );

    event ackFailed(
        string indexed denom,
        uint64 indexed amount,
        address indexed sender,
        address indexed receiver,
        uint64 indexed timeoutHeight
    );

    event ackSuccessful
    (
        string indexed denom,
        uint64 indexed amount,
        address indexed sender,
        address indexed receiver,
        uint64 indexed timeoutHeight
    );
}

