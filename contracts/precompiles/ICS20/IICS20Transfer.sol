// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

interface IICS20Transfer {

    // Transfer event: To be emitted when user calls the transfer function
    event Transfer(
        string indexed denom,
        uint64 amount,
        address indexed sender,
        address indexed receiver
    );

    // Transfer event: To be emitted when user recieves a transfer from another ICS20 successfully
    event recieveTransfer(
        string indexed denom,
        uint64 amount,
        address indexed sender,
        address indexed receiver
    );

    // Failed acknowledgement
    event ackFailed(
        string indexed denom,
        uint64 amount,
        address indexed sender,
        address indexed receiver,
        uint64 timeoutHeight
    );

    // Successful acknowledgement
    event ackSuccessful
    (
        string indexed denom,
        uint64 amount,
        address indexed sender,
        address indexed receiver,
        uint64 timeoutHeight
    );
}

