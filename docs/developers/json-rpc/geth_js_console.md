<!--
order: 6
-->

# Geth JavaScript Console Guide

Use this guide to learn how to set up and use the Geth JS console with an Evmos node. {synopsis}

Go-ethereum responds to instructions encoded as JSON objects as defined in the [JSON-RPC-API](https://geth.ethereum.org/docs/rpc/server). To perform and test these instructions, developers can use tools like curl. However, this is a low level and rather error-prone way to interact with the node. Most developers prefer to use convenient libraries that abstract away some of the more tedious and awkward tasks such as converting values from hexadecimal strings into numbers, or converting between denominations of ether (Wei, Gwei, etc). One such library is Web3.js. The purpose of Geth’s Javascript console is to provide a built-in environment to use a subset of the Web3.js libraries to interact with a Geth node. You can use this powerful tool to interact with an Evmos node too!

## Pre-requisite Readings

- [Go-ethereum CLI](https://geth.ethereum.org/docs/interface/javascript-console) {prereq}
- [Evmos (local) node](https://docs.evmos.org/developers/localnet/single_node.html) {prereq}

### Installing Go-Ethereum

Install the Go-ethereum CLI (`geth`) following the procedure corresponding to your OS in the [geth docs](https://geth.ethereum.org/docs/install-and-build/installing-geth). This will include the Javascript console.

Check that the installation was successful by running the following command:

```bash
geth version
```

If everything went as expected, you should have an output similar to this:

```bash
Geth
Version: 1.10.26-stable
Git Commit: e5eb32acee19cc9fca6a03b10283b7484246b15a
Architecture: amd64
Go Version: go1.18.5
Operating System: linux
GOPATH=/home/tom/go
GOROOT=/usr/local/go
```

## Install dependencies

<!-- markdown-link-check-disable-next-line -->

Make sure you have installed all the dependencies mentioned in the **[Pre-requisite Readings](#pre-requisite-readings)** section.

## Run Evmos local node

- Clone the [evmos repository](https://github.com/evmos/evmos) (if you haven’t already)
- Run the `local_node.sh` script to start a local node

```bash
git clone https://github.com/evmos/evmos.git
cd evmos
./local_node.sh
```

## Attach geth JS console

Wait a few seconds for the node to start up the JSON-RPC server. The local node has the HTTP-RPC server enabled and is listening at port 8545 by default. Attach a `geth` console to your node with the following command:

```bash
 $ geth attach http://127.0.0.1:8545
Welcome to the Geth JavaScript console!

instance: Version dev ()
Compiled at  using Go go1.19.2 (amd64)
coinbase: 0x7771fD5e52cf6A81B49d7EF40Bfc2bd0eA8A92E6
at block: 975 (Fri Nov 25 2022 01:21:52 GMT-0300 (-03))
 modules: debug:1.0 eth:1.0 net:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0

To exit, press ctrl-d or type exit
> 
```

## Use JSON-RPC methods

Now we can use all implemented JSON-RPC methods. Find an exhaustive list of the supported JSON-RPC methods on [Evmos docs](https://docs.evmos.org/developers/json-rpc/endpoints.html).

Below are some examples of how to use the console.

### Check current block height

We can check the current block height of the chain:

```javascript
> eth.blockNumber
1003
> eth.blockNumber
1004
```

### Get accounts

Get an array of the existing accounts in the keyring. To do so, use the following method:

```javascript
> eth.accounts
["0xf0c3878dd8de6edc0702c06c2bb9a8e380397173", "0xfdd268dfeca95cff23ba385dec161defea031682", "0x35ab07f08f9af9166e9225b3407ad5e63756a084", "0x6a36c1efef7dd58981b3999217cdb3ae720cf330"]
```

### Get chain id

We can get the chain id using:

```javascript
> net.version
"9000"
```

### Check balances

Check any account balance using the `eth.getBalance` method:

```javascript
> eth.getBalance(eth.accounts[0])
9.9999e+25
```

We get a big number because the result is denominated in `aevmos`. We can convert to Evmos (10 ^18 `aevmos`) using the `web3.fromWei` method:

```javascript
> web3.fromWei(eth.getBalance(eth.accounts[0]),"ether")
99999000
```

### Send transactions

We can perform token transfers using the corresponding method. For example, let's transfer 1 Evmos token from our account to another account:

```javascript
> eth.sendTransaction({from:eth.accounts[0], to: eth.accounts[1], value: web3.toWei(1, "ether")})
"0x902dfba22a8b7aaa599aa3ea35c8d60991f497ba2fe6c519ad7a7e1e4a2f3e8f"
```

As a response, we get back the transaction hash.

Now we can check the balance of the sender and receiver accounts.

The sender balance is reduced by 1 Evmos token and the fees paid for the transaction:

```javascript
> web3.fromWei(eth.getBalance(eth.accounts[0]),"ether")
99998998.999990548370552
```

The receiver account balance initially was 100000000 Evmos tokens. After the transaction, the account balance has increased by 1 Evmos token.

```javascript
> web3.fromWei(eth.getBalance(eth.accounts[1]),"ether")
100000001
```

## 🚪 Exit geth console

To exit the geth console use:

```bash
 > exit
```

or type `Ctrl + D`.

## 🪄 Tips & tricks

### List commands

A small trick to see the list of initial commands. Type 2 spaces then hit TAB twice. You will get:

```bash
>
AggregateError        Function              Object                TypeError             _consoleWeb3Transport encodeURIComponent    loadScript            toLocaleString
Array                 GoError               Promise               URIError              _setInterval          escape                message               toString
ArrayBuffer           Infinity              Proxy                 Uint16Array           _setTimeout           eth                   net                   txpool
BigNumber             Int16Array            RangeError            Uint32Array           clearInterval         eval                  parseFloat            undefined
Boolean               Int32Array            ReferenceError        Uint8Array            clearTimeout          globalThis            parseInt              unescape
DataView              Int8Array             Reflect               Uint8ClampedArray     console               hasOwnProperty        personal              valueOf
Date                  JSON                  RegExp                WeakMap               constructor           inspect               propertyIsEnumerable  web3
Error                 Map                   Set                   WeakSet               debug                 isFinite              require
EvalError             Math                  String                Web3                  decodeURI             isNaN                 rpc
Float32Array          NaN                   Symbol                XMLHttpRequest        decodeURIComponent    isPrototypeOf         setInterval
Float64Array          Number                SyntaxError           __proto__             encodeURI             jeth                  setTimeout
```

The same applies to the different namespaces. For example, you can type `eth.` and hit TAB twice. You will get:

```bash
> eth.
eth._requestManager            eth.fillTransaction            eth.getGasPrice                eth.getTransaction             eth.protocolVersion
eth.accounts                   eth.filter                     eth.getHashrate                eth.getTransactionCount        eth.resend
eth.blockNumber                eth.gasPrice                   eth.getHeaderByHash            eth.getTransactionFromBlock    eth.sendIBANTransaction
eth.call                       eth.getAccounts                eth.getHeaderByNumber          eth.getTransactionReceipt      eth.sendRawTransaction
eth.chainId                    eth.getBalance                 eth.getLogs                    eth.getUncle                   eth.sendTransaction
eth.coinbase                   eth.getBlock                   eth.getMaxPriorityFeePerGas    eth.getWork                    eth.sign
eth.compile                    eth.getBlockByHash             eth.getMining                  eth.hashrate                   eth.signTransaction
eth.constructor                eth.getBlockByNumber           eth.getPendingTransactions     eth.iban                       eth.submitTransaction
eth.contract                   eth.getBlockNumber             eth.getProof                   eth.icapNamereg                eth.submitWork
eth.createAccessList           eth.getBlockTransactionCount   eth.getProtocolVersion         eth.isSyncing                  eth.syncing
eth.defaultAccount             eth.getBlockUncleCount         eth.getRawTransaction          eth.maxPriorityFeePerGas
eth.defaultBlock               eth.getCode                    eth.getRawTransactionFromBlock eth.mining
eth.estimateGas                eth.getCoinbase                eth.getStorageAt               eth.namereg
eth.feeHistory                 eth.getCompilers               eth.getSyncing                 eth.pendingTransactions
```
