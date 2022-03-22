pragma solidity ^0.8.9;

import "./IICS20Transfer.sol";
import "./ICS20Bank.sol";
import "../lib/strings.sol";

abstract contract ICS20 is IICS20Transfer, ICS20Bank
{
    mapping(string => address) channelEscrowAddresses;

    constructor(string memory denom_) {
        _denom = denom_;
    }

    FungibleTokenPacketData _packetData = FungibleTokenPacketData(_denom);

    function sendTx
    (
        string denom,
        uint256 amount,
        address receiver,
        uint256 timeoutHeight
    ) override external 
    {
        if (!denom.toSlice().startsWith(_makeDenomPrefix(sourcePort, sourceChannel))) 
        { // sender is source chain
            require(_transferFrom(_msgSender(), _getEscrowAddress(sourceChannel), denom, amount));
        } 
        else 
        {
            require(_burn(_msgSender(), denom, amount));
        }

        emit Transfer(denom, amount, _msgSender(), receiver);
    }

    function recvTx(string memory _denom) override external returns (bytes memory acknowledgement) {
        // TODO: Validate receieve txns
        // TODO: Emit events for receive txns
    }

    function _transferFrom(address sender, address receiver, string memory denom, uint256 amount) virtual internal returns (bool);

    function _mint(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    function _burn(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    function _makeDenomPrefix(string memory port, string memory channel) virtual internal pure returns (strings.slice memory) {
        return port.toSlice()
            .concat("/".toSlice()).toSlice()
            .concat(channel.toSlice()).toSlice()
            .concat("/".toSlice()).toSlice();
    }

    function _getEscrowAddress(string memory channel) virtual internal pure returns (address) {
        if (channelEscrowAddresses[channel] == address(0)) {
            channelEscrowAddresses[channel] = _getEscrowAddressImpl(channel);
        }
        return channelEscrowAddresses[channel];
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
}

/* GLOSSARY
* The sourcePort identifies the port on the sending chain.
* The sourceChannel identifies the channel end on the sending chain.
* The timeoutHeight indicates a consensus height on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
* The timeoutTimestamp indicates a timestamp on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
*/