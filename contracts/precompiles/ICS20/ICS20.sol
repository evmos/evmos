pragma solidity ^0.8.9;
import "./IICS20Transfer.sol";

abstract contract ICS20 is IICS20Transfer
{
    // mapping(string => address) channelEscrowAddresses;

    // constructor(IBCHost host_, IBCHandler ibcHandler_) 
    // {
    //     ibcHost = host_;
    //     ibcHandler = ibcHandler_;
    // }

    function sendTransfer(
        string calldata denom,
        uint64 amount,
        address receiver,
        string calldata sourcePort,
        string calldata sourceChannel,
        uint64 timeoutHeight
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

}