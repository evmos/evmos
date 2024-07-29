// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "../../distribution/DistributionI.sol";

contract Reverter {
    uint counter = 0;

    constructor() payable {}

    function run() external {
        counter++;

        // call Reverter::transferFunds() externally to create new context, sending this contract's full balance.
        try
            Reverter(payable(address(this))).transferFunds(
                bytes32(counter),
                address(this).balance
            )
        {} catch {
            // Transferer created by the call to Reverter::transferFunds(), as well as the funds sent
            // to it, should not exist. Changes should be reverted
            // and trying to call the Transferer(t).withdraw() should make the tx fail
            address t = predictAddress(bytes32(counter));
            Transferer(t).withdraw();
        }
        // increment the salt
        counter++;
    }

    function transferFunds(bytes32 salt, uint v) external {
        // create Transferer, and send it v native tokens
        new Transferer{value: v, salt: salt}();

        // call the distribution precompile
        DISTRIBUTION_CONTRACT.delegationTotalRewards(address(this));

        // newly-created Transferer is removed from the journal, and the native tokens are returned to this
        // contract.
        revert();
    }

    // calculates the CREATE2 address of deploying Transferer with some salt
    function predictAddress(bytes32 salt) internal view returns (address) {
        address predictedAddress = address(
            uint160(
                uint(
                    keccak256(
                        abi.encodePacked(
                            bytes1(0xff),
                            address(this),
                            salt,
                            keccak256(
                                abi.encodePacked(type(Transferer).creationCode)
                            )
                        )
                    )
                )
            )
        );

        return predictedAddress;
    }

    // receive native tokens
    receive() external payable {}
}

contract Transferer {
    constructor() payable {}

    function withdraw() external {
        payable(msg.sender).transfer(address(this).balance);
    }
}
