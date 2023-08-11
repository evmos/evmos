// SPDX-License-Identifier: LGPL-v3
pragma solidity >=0.8.17;

import "./../../ics20/ICS20.sol";
import "./../../common/Types.sol";

contract InterchainSender {
    /// @dev Approves the required spend limits for IBC transactions.
    /// @dev This creates a Cosmos Authorization Grants for the given methods.
    /// @dev This emits an Approval event.
    function testApprove(Allocation[] calldata allocs) public {
        bool success = ICS20_CONTRACT.approve(address(this), allocs);
        require(success, "Failed to perform approval");
    }

    function testRevoke() external {
        bool success = ICS20_CONTRACT.revoke(address(this));
        require(success, "Failed to revoke approval");
    }

    function testIncreaseAllowance(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount
    ) public {
        bool success = ICS20_CONTRACT.increaseAllowance(
            address(this),
            sourcePort,
            sourceChannel,
            denom,
            amount
        );
        require(success, "Failed to increase allowance");
    }

    function testDecreaseAllowance(
        string calldata sourcePort,
        string calldata sourceChannel,
        string calldata denom,
        uint256 amount
    ) public {
        bool success = ICS20_CONTRACT.decreaseAllowance(
            address(this),
            sourcePort,
            sourceChannel,
            denom,
            amount
        );
        require(success, "Failed to decrease allowance");
    }

    /// @dev transfer a given amount of tokens. Returns the IBC packet sequence of the IBC transaction.
    /// @dev This emits a IBCTransfer event.
    /// @param sourcePort The source port of the IBC transfer.
    /// @param sourceChannel The source channel of the IBC transfer.
    /// @param denom The denomination of the tokens to transfer.
    /// @param receiver The receiver address on the receiving chain.
    /// @param timeoutHeight The timeout height for the IBC packet.
    /// @param timeoutTimestamp The timeout timestamp of the IBC packet.
    /// @param memo The IBC transaction memo.
    /// @param amount The amount of tokens to transfer to another chain.
    /// @return nextSequence The IBC transaction sequence number.
    function testTransferUserFunds(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        Height memory timeoutHeight,
        uint64 timeoutTimestamp,
        string memory memo
    ) public returns (uint64 nextSequence) {
        return
            ICS20_CONTRACT.transfer(
                sourcePort,
                sourceChannel,
                denom,
                amount,
                msg.sender,
                receiver,
                timeoutHeight,
                timeoutTimestamp,
                memo
            );
    }

    function testTransferContractFunds(
        string memory sourcePort,
        string memory sourceChannel,
        string memory denom,
        uint256 amount,
        string memory receiver,
        Height memory timeoutHeight,
        uint64 timeoutTimestamp,
        string memory memo
    ) public returns (uint64 nextSequence) {
        return
            ICS20_CONTRACT.transfer(
                sourcePort,
                sourceChannel,
                denom,
                amount,
                address(this),
                receiver,
                timeoutHeight,
                timeoutTimestamp,
                memo
            );
    }

    // QUERIES
    function testDenomTraces(
        PageRequest calldata pageRequest
    )
        public
        returns (
            DenomTrace[] memory denomTraces,
            PageResponse memory pageResponse
        )
    {
        return ICS20_CONTRACT.denomTraces(pageRequest);
    }

    function testDenomTrace(
        string memory hash
    ) public returns (DenomTrace memory denomTrace) {
        return ICS20_CONTRACT.denomTrace(hash);
    }

    function testDenomHash(
        string memory trace
    ) public returns (string memory hash) {
        return ICS20_CONTRACT.denomHash(trace);
    }

    function testAllowance(
        address owner,
        address spender
    ) public view returns (Allocation[] memory allocations) {
        return ICS20_CONTRACT.allowance(owner, spender);
    }
}
