// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity ^0.8.0;

import "./IPacketActor.sol";

contract PacketActorCounter is IPacketActor {

    int public counter = 0;

    function onSendPacket(
        address relayer
    ) external returns (bool success) {
        counter += 1;
        return true;
    }

    function onRecvPacket(
        Packet calldata packet,
        address relayer
    ) external returns (bytes calldata acknowledgement) {

        return packet.data;
    }

    function onAcknowledgementPacket(
        Packet calldata packet,
        bytes calldata acknowledgement,
        address relayer
    ) external returns (bool success) {

        return true;
    }

    function onTimeoutPacket(
        Packet calldata packet,
        address relayer
    ) external returns (bool success) {
        return true;
    }
}