<!--
order: 1
-->

# Concepts

## Registration of a Contract

Any contract can be registered by a developer by submitting a signed transaction. The signer of this transaction must match the address of the deployer of the contract in order for the registration to succeed. After the transaction is executed successfully, the contract deployer will start receiving a portion of the transaction fees when a user interacts with the registered contract.

### Fee Distribution

As described above, developers will earn a portion of the transaction fee after they register their contracts. The transactions eligible for the transaction only correspond to EVM transactions (`MsgEthereumTx`).

### Address Derivation

When registering a smart contract, the deployer provides an array of nonces, used to derive the contract’s address.

The smart contract can be directly deployed by an Externally Owned Account ([EOA](https://ethereum.stackexchange.com/questions/5828/what-is-an-eoa-account)) or created through one or more [factory](https://en.wikipedia.org/wiki/Factory_method_pattern) pattern contracts. If it was deployed by an EOA account, then the array of nonces only contains the EOA nonce for the deployment transaction. If it was deployed by one or more factories, the nonces contain the EOA nonce for the origin factory contract, followed by the nonce of the factory at the time of the creation of the next factory/contract.
