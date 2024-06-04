// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;

import "./IERC165.sol";

// The ICS20Data from the FungiblePacketData in the ICS20 standard
struct ICS20Data {
    string denom;
    uint256 amount;
    address sender;
    string receiver;
    string memo;
}

// The custom ICS20Packet struct that includes the ICS20Data
struct ICS20Packet {
    // Identifies the port on the sending chain
    string sourcePort;
    // Identifies the channel end on the sending chain
    string sourceChannel;
    // Identifies the port of the receiving chain
    string destinationPort;
    // Identifies the channel of the receiving chain
    string destinationChannel;
    // Actual opaque bytes transferred directly to the application module
    ICS20Data data;
    // Block height after which the packet times out
    Height timeoutHeight;
    // Block timestamp (in nanoseconds) after which the packet times out
    uint64 timeoutTimestamp;
}

// Define the Height struct
struct Height {
    uint64 revisionNumber;
    uint64 revisionHeight;
}

/// @author Evmos Team
/// @title Packet Actor Interface
/// @dev The interface through which solidity contracts define their own IBC
/// packet callbacks handling logic
interface IPacketActor is IERC165 {

    /// @dev onSendPacket will be called on the IBCActor after the IBC application
    /// handles the RecvPacket callback if the packet has an IBC Actor as a receiver.
    /// @param relayer The relayer address that sent the packet.
    function onSendPacket(
        ICS20Packet calldata packet,
        address relayer
    ) external returns (bool success);

    /// @dev onRecvPacket will be called on the IBCActor after the IBC Application
    /// handles the RecvPacket callback if the packet has an IBC Actor as a receiver.
    /// @param packet The IBC packet received.
    function onRecvPacket(
        ICS20Packet calldata packet
    ) external returns (bool success);

    /// @dev onAcknowledgementPacket will be called on the IBC Actor
    /// after the IBC Application handles its own OnAcknowledgementPacket callback
    /// @param packet The IBC packet acknowledged.
    /// @param acknowledgement The IBC transaction acknowledgement (success or error) bytes.
    /// @param relayer The relayer that handled the acknowledgment.
    /// @return success The success or failure boolean.
    function onAcknowledgementPacket(
        ICS20Packet calldata packet,
        bytes calldata acknowledgement,
        address relayer
    ) external returns (bool success);

    /// @dev onTimeoutPacket will be called on the IBC Actor
    /// after the IBC Application handles its own OnTimeoutPacket callback.
    /// @param packet The IBC packet that timeouted.
    /// @param relayer The relayer that handled the timeout.
    /// @return success The success or failure boolean.
    function onTimeoutPacket(
        ICS20Packet calldata packet,
        address relayer
    ) external returns (bool success);
}

/// @dev The abstract contract that implements the IPacketActor interface and provides a default implementation for supportsInterface
abstract contract PacketActorBase is IPacketActor {

    function supportsInterface(bytes4 interfaceId) public view virtual override returns (bool) {
        return (
            interfaceId == this.supportsInterface.selector ||
            interfaceId == this.onSendPacket.selector ||
            interfaceId == this.onRecvPacket.selector ||
            interfaceId == this.onAcknowledgementPacket.selector ||
            interfaceId == this.onTimeoutPacket.selector
        );
    }
}