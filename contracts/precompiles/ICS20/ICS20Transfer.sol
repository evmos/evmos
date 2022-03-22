// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.9;

import "./IICS20Transfer.sol";
import "../core/types/Channel.sol";
import "../core/IBCModule.sol";
import "../core/IBCHandler.sol";
import "../core/IBCHost.sol";
import "../core/types/App.sol";
import "../lib/strings.sol";
import "../lib/Bytes.sol";
import "openzeppelin-solidity/contracts/utils/Context.sol";

abstract contract ICS20Transfer is Context, IICS20Transfer {
    using strings for *;
    using Bytes for *;

    IBCHandler ibcHandler;
    IBCHost ibcHost;

    mapping(string => address) channelEscrowAddresses;

    constructor(IBCHost host_, IBCHandler ibcHandler_) {
        ibcHost = host_;
        ibcHandler = ibcHandler_;
    }

    function sendTransfer(
        string calldata denom,
        uint64 amount,
        address receiver,
        string calldata sourcePort,
        string calldata sourceChannel,
        uint64 timeoutHeight
    ) override virtual external {
        if (!denom.toSlice().startsWith(_makeDenomPrefix(sourcePort, sourceChannel))) { // sender is source chain
            require(_transferFrom(_msgSender(), _getEscrowAddress(sourceChannel), denom, amount));
        } else {
            require(_burn(_msgSender(), denom, amount));
        }

        _sendPacket(
            FungibleTokenPacketData.Data({
                denom: denom,
                amount: amount,
                sender: abi.encodePacked(_msgSender()),
                receiver: abi.encodePacked(receiver)
            }),
            sourcePort,
            sourceChannel,
            timeoutHeight
        );

        emit Transfer(denom, amount, _msgSender(), receiver);
    }

    /// Module callbacks ///

    function onRecvPacket(Packet.Data calldata packet) external virtual override returns (bytes memory acknowledgement) {
        FungibleTokenPacketData.Data memory data = FungibleTokenPacketData.decode(packet.data);
        strings.slice memory denom = data.denom.toSlice();
        strings.slice memory trimedDenom = data.denom.toSlice().beyond(
            _makeDenomPrefix(packet.source_port, packet.source_channel)
        );
        if (!denom.equals(trimedDenom)) { // receiver is source chain
            return _newAcknowledgement(
                _transferFrom(_getEscrowAddress(packet.destination_channel), data.receiver.toAddress(), trimedDenom.toString(), data.amount)
            );
        } else {
            string memory prefixedDenom = _makeDenomPrefix(packet.destination_port, packet.destination_channel).concat(denom);
            return _newAcknowledgement(
                _mint(data.receiver.toAddress(), prefixedDenom, data.amount)
            );
        }

        emit recieveTransfer(data.denom, data.amount, data.sender.toAddress(), data.receiver.toAddress());
    }

    function onAcknowledgementPacket(Packet.Data calldata packet, bytes calldata acknowledgement) external virtual override {
        if (!_isSuccessAcknowledgement(acknowledgement)) {
            _refundTokens(FungibleTokenPacketData.decode(packet.data), packet.source_port, packet.source_channel);
            emit ackFailed(
                FungibleTokenPacketData.decode(packet.data).denom,
                FungibleTokenPacketData.decode(packet.data).amount,
                FungibleTokenPacketData.decode(packet.data).sender.toAddress(),
                FungibleTokenPacketData.decode(packet.data).receiver.toAddress(),
                packet.timeout_height
            );
        }

        emit ackSuccessful
        (
            FungibleTokenPacketData.decode(packet.data).denom,
            FungibleTokenPacketData.decode(packet.data).amount,
            FungibleTokenPacketData.decode(packet.data).sender.toAddress(),
            FungibleTokenPacketData.decode(packet.data).receiver.toAddress(),
            packet.timeout_height
        );
    }

    function onChanOpenInit(Channel.Order, string[] calldata, string calldata, string calldata channelId, ChannelCounterparty.Data calldata, string calldata) external virtual override {
        // TODO authenticate a capability
        channelEscrowAddresses[channelId] = address(this);
    }

    function onChanOpenTry(Channel.Order, string[] calldata, string calldata, string calldata channelId, ChannelCounterparty.Data calldata, string calldata, string calldata) external virtual override {
        // TODO authenticate a capability
        channelEscrowAddresses[channelId] = address(this);
    }

    function onChanOpenAck(string calldata portId, string calldata channelId, string calldata counterpartyVersion) external virtual override {}

    function onChanOpenConfirm(string calldata portId, string calldata channelId) external virtual override {}

    function onChanCloseInit(string calldata portId, string calldata channelId) external virtual override {}

    function onChanCloseConfirm(string calldata portId, string calldata channelId) external virtual override {}

    /// Internal functions ///

    function _transferFrom(address sender, address receiver, string memory denom, uint256 amount) virtual internal returns (bool);

    function _mint(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    function _burn(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    function _sendPacket(FungibleTokenPacketData.Data memory data, string memory sourcePort, string memory sourceChannel, uint64 timeoutHeight) virtual internal {
        (Channel.Data memory channel, bool found) = ibcHost.getChannel(sourcePort, sourceChannel);
        require(found, "channel not found");
        ibcHandler.sendPacket(Packet.Data({
            sequence: ibcHost.getNextSequenceSend(sourcePort, sourceChannel),
            source_port: sourcePort,
            source_channel: sourceChannel,
            destination_port: channel.counterparty.port_id,
            destination_channel: channel.counterparty.channel_id,
            data: FungibleTokenPacketData.encode(data),
            timeout_height: Height.Data({revision_number: 0, revision_height: timeoutHeight}),
            timeout_timestamp: 0
        }));
    }

    function _getEscrowAddress(string memory sourceChannel) virtual internal view returns (address) {
        address escrow = channelEscrowAddresses[sourceChannel];
        require(escrow != address(0));
        return escrow;
    }

    function _newAcknowledgement(bool success) virtual internal pure returns (bytes memory) {
        bytes memory acknowledgement = new bytes(1);
        if (success) {
            acknowledgement[0] = 0x01;
        } else {
            acknowledgement[0] = 0x00;
        }
        return acknowledgement;
    }
    
    function _isSuccessAcknowledgement(bytes memory acknowledgement) virtual internal pure returns (bool) {
        require(acknowledgement.length == 1);
        return acknowledgement[0] == 0x01;
    }

    function _refundTokens(FungibleTokenPacketData.Data memory data, string memory sourcePort, string memory sourceChannel) virtual internal {
        if (!data.denom.toSlice().startsWith(_makeDenomPrefix(sourcePort, sourceChannel))) { // sender was source chain
            require(_transferFrom(_getEscrowAddress(sourceChannel), data.sender.toAddress(), data.denom, data.amount));
        } else {
            require(_mint(data.sender.toAddress(), data.denom, data.amount));
        }
    }

    /// Helper functions ///

    function _makeDenomPrefix(string memory port, string memory channel) virtual internal pure returns (strings.slice memory) {
        return port.toSlice()
            .concat("/".toSlice()).toSlice()
            .concat(channel.toSlice()).toSlice()
            .concat("/".toSlice()).toSlice();
    }
}
