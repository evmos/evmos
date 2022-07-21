<!--
order: 1
-->

# Wallet Integration

Learn how to properly integrate [Metamask](https://metamask.io/) or [Keplr](https://www.keplr.app/) with a dApp on Evmos. {synopsis}

:::tip
**Note**: want to learn more about wallet integration beyond what's covered here? Check out both the [MetaMask Wallet documentation](https://docs.metamask.io/guide/) and [Keplr Wallet documentation](https://docs.keplr.app/).
:::

## Pre-requisite Readings

- [MetaMask documentation](https://docs.metamask.io/guide/) {prereq}
- [Keplr documentation](https://docs.keplr.app/) {prereq}

## Implementation Checklist

The integration implementation checklist for dApp developers consists of three categories:

1. [Frontend features](#frontend)
2. [Transactions and wallet interactions](#transactions)
3. [Client-side provider](#connections)

### Frontend

Make sure to create a wallet-connection button for Metamask and/or Keplr on the frontend of the application. For instance, consider the "Connect to a wallet" button on the interface of [Diffusion Finance](https://app.diffusion.fi/) or the analagous button on the interface of [EvmoSwap](https://app.evmoswap.org/).

### Transactions

Developers enabling transactions on their dApp have to [determine wallet type](#determining-wallet-type) of the user, [create the transaction](#create-the-transaction), [request signatures](#sign-and-broadcast-the-transaction) from the corresponding wallet, and finally [broadcast the transaction](#sign-and-broadcast-the-transaction) to the network.

#### Determining Wallet Type

Developers should determine whether users are using Keplr or MetaMask. Whether MetaMask or Keplr is installed on the user device can be determined by checking the corresponding `window.ethereum` or `window.keplr` value.

- **For MetaMask**: `await window.ethereum.enable(chainId);`
- **For Keplr**: `await window.keplr.enable(chainId);`

If either `window.ethereum` or `window.keplr` returns `undefined` after `document.load`, then MetaMask (or, correspondingly, Keplr) is not installed. There are several ways to wait for the load event to check the status: for instance, developers can register functions to `window.onload`, or they can track the document's ready state through the document event listener.

After the user's wallet type has been determined, developers can proceed with creating, signing, and sending transactions.

#### Create the Transaction

:::tip
**Note**: The example below uses the Evmos Testnet `chainID`. For more info, check the Evmos Chain IDs reference document [here](../../users/technical_concepts/chain_id.md).
:::

Developers can create `MsgSend` transactions using the [evmosjs](../libraries/evmosjs.md) library.

```js
import { createMessageSend } from @tharsis/transactions

const chain = {
    chainId: 9000,
    cosmosChainId: 'evmos_9000-4',
}

const sender = {
    accountAddress: 'evmos1mx9nqk5agvlsvt2yc8259nwztmxq7zjq50mxkp',
    sequence: 1,
    accountNumber: 9,
    pubkey: 'AgTw+4v0daIrxsNSW4FcQ+IoingPseFwHO1DnssyoOqZ',
}

const fee = {
    amount: '20',
    denom: 'aevmos',
    gas: '200000',
}

const memo = ''

const params = {
    destinationAddress: 'evmos1pmk2r32ssqwps42y3c9d4clqlca403yd9wymgr',
    amount: '1',
    denom: 'aevmos',
}

const msg = createMessageSend(chain, sender, fee, memo, params)

// msg.signDirect is the transaction in Keplr format
// msg.legacyAmino is the transaction with legacy amino
// msg.eipToSign is the EIP712 data to sign with metamask
```

#### Sign and Broadcast the Transaction

:::tip
**Note**: The example below uses an Evmos Testnet [RPC node](../connect.md#public-available-endpoints).
:::

<!-- textlint-disable -->
After creating the transaction, developers need to send the payload to the appropriate wallet to be signed ([`msg.signDirect`](https://docs.keplr.app/api/#sign-direct-protobuf) is the transaction in Keplr format, and `msg.eipToSign` is the [`EIP712`](https://eips.ethereum.org/EIPS/eip-712) data to sign with MetaMask).

With the signature, we add a Web3Extension to the transaction and broadcast it to the Evmos node.

<!-- textlint-enable -->
```js
// Note that this example is for MetaMask, using evmosjs

// Follow the previous code block to generate the msg object
import { evmosToEth } from '@tharsis/address-converter'
import { generateEndpointBroadcast, generatePostBodyBroadcast } from '@tharsis/provider'
import { createTxRawEIP712, signatureToWeb3Extension } from '@tharsis/transactions'

// Init Metamask
await window.ethereum.enable();

// Request the signature
let signature = await window.ethereum.request({
    method: 'eth_signTypedData_v4',
    params: [evmosToEth(sender.accountAddress), JSON.stringify(msg.eipToSign)],
});

// The chain and sender objects are the same as the previous example
let extension = signatureToWeb3Extension(chain, sender, signature)

// Create the txRaw
let rawTx = createTxRawEIP712(msg.legacyAmino.body, msg.legacyAmino.authInfo, extension)

// Broadcast it
const postOptions = {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: generatePostBodyBroadcast(rawTx),
};

let broadcastPost = await fetch(
    `https://eth.bd.evmos.dev:8545${generateEndpointBroadcast()}`,
    postOptions
);
let response = await broadcastPost.json();
```

### Connections

For Ethereum RPC, Evmos gRPC, and/or REST queries, dApp developers should implement providers client-side, and store RPC details in the environment variable as secrets.
