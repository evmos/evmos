// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;


/// @author Evmos Team
/// @title Packet Actor Interface
/// @dev The interface through which solidity contracts define their own IBC
/// packet callbacks handling logic
interface IPacketActor {
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