# Join Point-XNet-Triton as a Validator

DISCLAIMER: THE DOCUMENT IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF
OR IN CONNECTION WITH THE DOCUMENT.

Following this document requires highly experienced DevOps engineers that know how to run validators. This is not the tutorial for most Point Network users! Do not attempt to run the commands from your personal computer, just in case something goes wrong!

## Table of Contents

* [Overview](#overview)
* [Rewards](#rewards)
* [Prerequisites](#prerequisites)
* [Initialize the Node](#initialize-the-node)
* [Run the Node](#run-the-node)
* [Sending your first transaction](#sending-your-first-transaction)
* [Stake XPOINT and Join as a Validator](#stake-xpoint-and-join-as-a-validator)
* [What's Next?](#whats-next)

## Overview

This document describes step-by-step instructions on joining Point-XNet-Neptune testnet as a validator.

Validators have the responsibility to keep the network operational 24/7. Do not attempt to join the testnet (and especially mainnet) if you don’t have enough experience. For example, if you install it on your laptop, join as a validator, and then close the laptop, the network will penalize you for being offline by slashing your stake (+the network quality might degrade).

If you have any questions, join our Discord: https://pointnetwork.io/discord and ask in #validators channel. This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

## Rewards

If you submitted the form, you’ve received 1024 XPOINT testnet tokens on the address you’ve sent to us. The testnet is incentivized, meaning that although it runs on XPOINT, successful testnet validators will receive real POINT as the rewards.

But because everyone received the same amount (1024 XPOINT), your testnet rewards will be multiplied by the same factor shared by everyone, **and** also by the amount of real POINT you will have at the mainnet launch (basically, we want to reward you on the testnet ***as if you already do that with real POINT***, but before we can possibly know how much you would’ve had). Meaning that if you have 0 POINT at the mainnet starts, your stake on the testnet is basically 0, no matter your performance on the testnet. And for Validator A that will have 300 POINT and Validator B that will have 3000 POINT, their rewards will be adjusted by 300 times and 3000 times respectively (even though everyone receives 1024 XPOINT equally at the start).

## Prerequisites

Most of the commands here are provided for and tested on Ubuntu Server 22.04 LTS, so change them accordingly (pacman instead of apt-get if you’re on Arch for example, or brew on Mac OS)

Ensure Go is installed:

```go version```

If Golang is not installed, install it using official tutorial: https://go.dev/doc/install.

Also to build node you would have to have `make` installed:

```sudo apt-get install build-essential```

Pull the repository of the point chain: 

```git clone https://github.com/pointnetwork/point-chain```

Stay on the _main_ branch and run this to compile the node from the sources:

```make install```

_Note: Point Chain is a fork of evmos, and by default the working directory is ~/.evmos. Make sure you don’t already have data for another evmos node on the device you’re running the validator from._

## Initialize the Node

Check if evmosd command is available for you. If you see ```evmosd: command not found``` message then export path for this command:

```export PATH=$PATH:$(go env GOPATH)/bin```

Configure your validator key:

```evmosd config keyring-backend file```

```evmosd config chain-id point_10721-1```


Generate a new key/mnemonic for validator: ```evmosd keys add validatorkey --keyring-backend file```
You may want to save output somewhere because it contains your Evmos address and other usefull information.

Run the init script

Init you validator where [myvalidator] is your validator custom name which will be publicly visible
  
```evmosd init [myvalidator] --chain-id point_10721-1```

Copy `genesis.json` and `config.toml` files from this repository https://github.com/pointnetwork/point-chain-config/tree/main/testnet-xNet-Triton-1 into `~/.evmosd/config`

Validate it:
  
```evmosd validate-genesis```
  
## Run the Node

Then run the node and wait for fully sync:
  
```evmosd start```

If you want it to also respond to the RPC commands, you can instead run:

```evmosd start --json-rpc.enable=true --json-rpc.api "eth,txpool,personal,net,debug,web3"```

You can run this command to see status of your node:
  
```evmosd status```

You will get the "latest_block_height" of your node.
  
To see current block height of blockchain run:

```curl  http://xnet-neptune-1.point.space:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'```

The result is in hexadecimal, just convert to decimal and see how far are you from full sync.

## Sending your first transaction

### Add custom network

Now while you're waiting for the node to sync, you need to send funds to your validator address. As mentioned, you should have received an airdrop of 1024 XPOINT if you filled in the form. To see them, you can import the private key into a wallet like Metamask (not a good idea for mainnet security, but ok for testnet tokens).

Then you need to add XNet-Triton into Metamask:

```
Network Title: Point XNet Triton
RPC URL: https://xnet-triton-1.point.space/
Chain ID: 10721
SYMBOL: XPOINT
```

### Add the wallet with your 1024 XPOINT

Remember the wallet you sent to us to be funded? In the form? It now has 1024 XPOINT.

Import the wallet with the private key into your wallet (e.g. Metamask), and you should see 1024 XPOINT there. But this is your fund wallet, not validator wallet.

### Find out which address is your validator wallet

Evmos has two wallet formats: Cosmos format, and Ethereum format. Cosmos format starts with `evmos` prefix, and Ethereum format starts with `0x`. Most people don't need to know about Cosmos format, but validators should have a way to change from one to another.

Run ```evmosd keys list```, and you will see a list of keys attached to your node. Look at the one which has the name `validatorkey`, and note its address (it should be in Cosmos format and start with `evmos` prefix).

(In most cases it is not needed, but if something goes wrong and if you ever want to import your validator wallet in your Metamask you will need the private key. You can get it with this command: `evmosd keys unsafe-export-eth-key validatorkey --keyring-backend file`)

Use this tool to convert it to Ethereum format: https://evmos.me/utils/tools

This is your validator address in Ethereum format.

### Fund the validator

Finally, use the wallet to send however much you need from your fund address to the validator address (you can send all 1024 or choose a different strategy).

## Stake XPOINT and Join as a Validator

Now you have to wait for the node to fully sync, because otherwise it will not find your.

Once the node is fully synced, and you got some XPOINT to stake, check your balance in the node, you 
will see your balance in Metamask or you can check your balance with this command:

```evmosd query bank balances  <evmosaddress>```

If you have enough balance stake your assets and check the transaction:

```
evmosd tx staking create-validator \
--amount=100000000000000000000apoint \
--pubkey=$(evmosd tendermint show-validator) \
--moniker="<myvalidator>" \
--chain-id=point_10721-1 \
--commission-rate="0.10" \
--commission-max-rate="0.20" \
--commission-max-change-rate="0.01" \
--min-self-delegation="100000000000000000000" \
--gas="400000" \
--gas-prices="0.025apoint" \
--from=validatorkey \
--keyring-backend file
```

(Note the amount: it's in apoint (which is 1/1e18 XPOINT). 100000000000000000000apoint is 100 XPOINT (when you remove 18 zeroes at the end). If you decide to adjust the amount, don't forget to adjust `min-self-delegation` flag too.)

You will have to provide your keystore password and approve the transaction for this command.

If everything works ok you will get a txhash. You can check the status of the tx: ```evmosd query tx <txhash>```

Transaction receipt may contain errors, so please check if there are any or if it's live. You can use the explorer or ask the node to provide receipt.

If the transaction was correct you should instantly become part of the validators set. Check your pubkey first:

```evmosd tendermint show-validator```

You will see a key there, you can identify your node among other validators using that key:

```evmosd query tendermint-validator-set```

There you will find more info like your VotingPower that should be bigger than 0. Also you can check your VotingPower by running:

```evmosd status```

## What's Next?

Please post on Discord channel #validators when you succeed! https://pointnetwork.io/discord

And if you have any questions, ask in #validators channel. This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

Also, check out extra documentation for validators:

- https://hub.cosmos.network/main/validators/validator-faq.html#
- https://docs.evmos.org/validators/overview.html

Share any feedback, questions, and ideas there!
