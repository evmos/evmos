pragma solidity ^0.8.9;

import "./IICS20Transfer.sol";
import "./ICS20Bank.sol";
import "../lib/strings.sol";

abstract contract ICS20 is IICS20Transfer, ICS20Bank
{
    /*
    * Channel Escrow Address is a mapping that maps the channel address to the escrow address.
    */
    mapping(string => address) channelEscrowAddresses; 

    constructor(string memory denom_) {
        _denom = denom_; // Setting a denomination for the ICS20
    }

    /*
    * Send transaction 
    */
    function sendTx (
        string calldata denom,
        uint64 amount,
        address receiver,
        string calldata sourcePort,
        string calldata sourceChannel,
        uint64 timeoutHeight
    ) public 
    {
         // Check if sender is on source chain
        if (!denom.toSlice().startsWith(_makeDenomPrefix(sourcePort, sourceChannel))) 
        { 
            require(_transferFrom(_msgSender(), _getEscrowAddress(sourceChannel), denom, amount));
        } 
        else 
        {
            // Burn the tokens from sender if not
            require(_burn(_msgSender(), denom, amount)); 
        }
        
        // Emit the transfer event with correct parameters
        emit Transfer(
            denom, amount, _msgSender(), receiver, sourcePort, sourceChannel, timeoutHeight
            ); 
    }

    /*
    * Recieve transaction from another ICS20
    */
    function onRecvTx(Packet.Data calldata packet) external
     virtual override 
     returns (bytes memory acknowledgement) 
     {
        // Fetches data structure from library, FungibleTokenPacketData
        FungibleTokenPacketData.Data memory data = FungibleTokenPacketData.decode(packet.data); 

        // Get denom of reciever chain
        strings.slice memory denom = data.denom.toSlice(); //
        strings.slice memory trimedDenom = data.denom.toSlice().beyond(
            _makeDenomPrefix(packet.source_port, packet.source_channel)
        );
        if (!denom.equals(trimedDenom)) { 
            // Check if receiver is source chain
            return _newAcknowledgement(
                _transferFrom(_getEscrowAddress(packet.destination_channel), data.receiver.toAddress(), trimedDenom.toString(), data.amount)
            );

        } else {
            string memory prefixedDenom = _makeDenomPrefix(packet.destination_port, packet.destination_channel).concat(denom);
            return _newAcknowledgement(
                _mint(data.receiver.toAddress(), prefixedDenom, data.amount)
            );
        }
     }

    /*
    Internal function interfaces
    */

    // Transfer Function to transfer amount in certain denom from sender to reciever
    function _transferFrom(address sender, address receiver, string memory denom, uint256 amount) virtual internal returns (bool);
        
    // Mint function to mint amount in certain denom to receiver
    function _mint(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    // Burn function to burn amount in certain denom from sender
    function _burn(address account, string memory denom, uint256 amount) virtual internal returns (bool);

    // Cleaning denom string
    function _makeDenomPrefix(string memory port, string memory channel) virtual internal pure returns (strings.slice memory) {
        return port.toSlice()
            .concat("/".toSlice()).toSlice()
            .concat(channel.toSlice()).toSlice()
            .concat("/".toSlice()).toSlice();
    }

    // Function to create a new acknowledgement for a certain transfer function
    function _newAcknowledgement(bool success) virtual internal pure returns (bytes memory) {
        bytes memory acknowledgement = new bytes(1);
        if (success) {
            acknowledgement[0] = 0x01; // Successful acknowledgement
        } else {
            acknowledgement[0] = 0x00; // Failed acknowledgement
        }
        return acknowledgement;
    }
    
    // Check for successful acknowledgement
    function _isSuccessAcknowledgement(bytes memory acknowledgement) virtual internal pure returns (bool) {
        require(acknowledgement.length == 1);
        return acknowledgement[0] == 0x01;
    }

    // Refund function to refund tokens to sender
    function _refundTokens(
        FungibleTokenPacketData memory data, 
        string memory sourcePort, 
        string memory sourceChannel) 
        virtual internal {


        if (!data.denom.toSlice().startsWith(_makeDenomPrefix(sourcePort, sourceChannel))) { // sender was source chain
            require(_transferFrom(_getEscrowAddress(sourceChannel), data.sender.toAddress(), data.denom, data.amount));
        } else {
            require(_mint(data.sender.toAddress(), data.denom, data.amount));
        }
    }


    // Get escrow address for a certain channel
    function _getEscrowAddress(string memory channel) virtual internal pure returns (address) {
        if (channelEscrowAddresses[channel] == address(0)) {
            channelEscrowAddresses[channel] = _getEscrowAddressImpl(channel);
        }   
        
        return channelEscrowAddresses[channel];
    }

}

/* 
* The sourcePort identifies the port on the sending chain.
* The sourceChannel identifies the channel end on the sending chain.
* The timeoutHeight indicates a consensus height on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
* The timeoutTimestamp indicates a timestamp on the destination chain after which the packet will no longer be processed, and will instead count as having timed-out.
*/