<!--
order: 8
-->

# Geth JavaScript Console Guide

Learn how to set up and use the Geth JS console with an Evmos node. {synopsis}

## ✅ Requirements

- Geth
- Evmos (local) node

### Installing Geth

Install geth following the procedure corresponding to your OS in the [geth docs](https://geth.ethereum.org/docs/install-and-build/installing-geth).

Check that the installation was successful by running the following command:

```
geth version
```

If everything went as expected, you should have an output similar to this:

```
Geth
Version: 1.10.26-stable
Git Commit: e5eb32acee19cc9fca6a03b10283b7484246b15a
Architecture: amd64
Go Version: go1.18.5
Operating System: linux
GOPATH=/home/tom/go
GOROOT=/usr/local/go
```

## 1️⃣ Install dependencies

Make sure you have installed all the dependencies mentioned in the **[Requirements](#-requirements)** section.

## 2️⃣ Run Evmos local node

- Clone the [evmos repository](https://github.com/evmos/evmos) (if you haven’t already)
- Run the `local_node.sh` script to start a local node

```
git clone https://github.com/evmos/evmos.git
cd evmos
./local_node.sh
```

## 3️⃣ Attach geth JS console

Wait a few seconds for the node to start up the JSON-RPC server.

The local node has the HTTP-RPC server enabled and listening at port 8545 by default. This is what we will connect to.

Attach geth console to your node with the following command:

```
geth attach http://127.0.0.1:8545
```

![https://i.imgur.com/rfN0T2i.png](https://i.imgur.com/rfN0T2i.png)

## 4️⃣ Use JSON-RPC methods

Now we can use all the implemented JSON-RPC methods. Find an exhaustive list of the supported JSON-RPC methods on [Evmos docs](https://docs.evmos.org/developers/json-rpc/endpoints.html).

Below are some examples of how to use the console.

### Check current block height

We can check the curreng block height of the chain.

```
eth.blockNumber
```

![https://i.imgur.com/Elaqhdl.png](https://i.imgur.com/Elaqhdl.png)

### Get accounts

Get an array of the existing accounts in the keyring. To do so, use the following method:

```
eth.accounts
```

![https://i.imgur.com/ONJczRV.png](https://i.imgur.com/ONJczRV.png)

### Get chain id

We can get the chain id using:

```
net.version
```

![https://i.imgur.com/LmqJW8T.png](https://i.imgur.com/LmqJW8T.png)

### Check balances

Check any account balance using the `eth.getBalance` method:

```
eth.getBalance(eth.accounts[0])
```

![https://i.imgur.com/4k2sNAe.png](https://i.imgur.com/4k2sNAe.png)

We get a big number because the result is denominated in `aevmos`. We can convert to Evmos (10 ^18 `aevmos`) using the `web3.fromWei` method:

```
web3.fromWei(eth.getBalance(eth.accounts[0]),"ether")
```

![https://i.imgur.com/YuRR19k.png](https://i.imgur.com/YuRR19k.png)

### Send transactions

We can perform token transfers using the corresponding method. For example, let's transfer 1 Evmos from our account to another account:

```
eth.sendTransaction({from:eth.accounts[0], to:"0xf6e443fd1c869c6a25d18a9866f3a6c7f8dfb703", value: web3.toWei(1, "ether")})
```

![https://i.imgur.com/9zd1pU7.png](https://i.imgur.com/9zd1pU7.png)

As a response, we get back the transaction hash

Now we can check the balance of the sender and receiver accounts.

Sender account balance:

```
web3.fromWei(eth.getBalance(eth.accounts[0]),"ether")
```

![https://i.imgur.com/E3mWzNI.png](https://i.imgur.com/E3mWzNI.png)

Receiver account balance:

```
web3.fromWei(eth.getBalance("0xf6e443fd1c869c6a25d18a9866f3a6c7f8dfb703"),"ether")
```

![https://i.imgur.com/SPm4gFR.png](https://i.imgur.com/SPm4gFR.png)

## 🚪 Exit geth console

To exit the geth console use:

```
exit
```

Or typing `Ctrl + D`

## 🪄 Tips & tricks

### List commands

A small trick to see the list of initial commands. Type 2 spaces then hit TAB twice. You will get:

![https://i.imgur.com/TZlM8M1.png](https://i.imgur.com/TZlM8M1.png)

The same applies to the different namespaces. For example, you can type `eth.` and hit TAB twice. You will get:

![https://i.imgur.com/97Xk3lo.png](https://i.imgur.com/97Xk3lo.png)
