// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;

    struct Packet {
        // Sequence number corresponds to the order of sends and receives
        uint64 sequence;
        // Identifies the port on the sending chain
        string sourcePort;
        // Identifies the channel end on the sending chain
        string sourceChannel;
        // Identifies the port on the receiving chain
        string destinationPort;
        // Identifies the channel end on the receiving chain
        string destinationChannel;
        // Actual opaque bytes transferred directly to the application module
        bytes data;
        // Block height after which the packet times out
        Height timeoutHeight;
        // Block timestamp (in nanoseconds) after which the packet times out
        uint64 timeoutTimestamp;
    }

// Define the Height struct if it's not already defined elsewhere in your code.
    struct Height {
        uint64 revisionNumber;
        uint64 revisionHeight;
    }

/// @author Evmos Team
/// @title Packet Actor Interface
/// @dev The interface through which solidity contracts define their own IBC
/// packet callbacks handling logic
interface IPacketActor {

    /// @dev onSendPacket will be called on the IBCActor after the IBC application
    /// handles the RecvPacket callback if the packet has an IBC Actor as a receiver.
    /// @param relayer The relayer address that sent the packet.
    function onSendPacket(
        address relayer
    ) external returns (bool success);


    /// @dev onRecvPacket will be called on the IBCActor after the IBC Application
    /// handles the RecvPacket callback if the packet has an IBC Actor as a receiver.
    /// @param packet The IBC packet received.
    /// @param relayer The relayer address that sent the packet.
    /// @return acknowledgement The success or failure acknowledgement bytes.
    function onRecvPacket(
        Packet calldata packet,
        address relayer
    ) external returns (bytes calldata acknowledgement);

    /// @dev onAcknowledgementPacket will be called on the IBC Actor
    /// after the IBC Application handles its own OnAcknowledgementPacket callback
    /// @param packet The IBC packet acknowledged.
    /// @param acknowledgement The IBC transaction acknowledgement (success or error) bytes.
    /// @param relayer The relayer that handled the acknowledgment.
    /// @return success The success or failure boolean.
    function onAcknowledgementPacket(
        Packet calldata packet,
        bytes calldata acknowledgement,
        address relayer
    ) external returns (bool success);

    /// @dev onTimeoutPacket will be called on the IBC Actor
    /// after the IBC Application handles its own OnTimeoutPacket callback.
    /// @param packet The IBC packet that timeouted.
    /// @param relayer The relayer that handled the timeout.
    /// @return success The success or failure boolean.
    function onTimeoutPacket(
        Packet calldata packet,
        address relayer
    ) external returns (bool success);
}