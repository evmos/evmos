<!--
order: 9
-->

# Future Improvements

- The fee distribution registration could be extended to register the withdrawal address to the owner of the contract according to [EIP173](https://eips.ethereum.org/EIPS/eip-173).
- Extend the supported message types for the transaction fee distribution to Cosmos transactions that interact with the EVM (eg: ERC20 module, IBC transactions).
- Distribute fees for internal transaction calls to other registered contracts. At this time, we only send transaction fees to the deployer of the smart contract represented by the `to` field of the transaction request (`MyContract`). We do not distribute fees to smart contracts called internally by `MyContract`.
- `CREATE2` opcode support for address derivation. When registering a smart contract, we verify that its address is derived from the deployer’s address. At this time, we only support the derivation path using the `CREATE` opcode, which accounts for most cases.
