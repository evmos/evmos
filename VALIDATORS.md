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


Generate a new key/mnemonic for validator: ```evmosd keys add validatorkey```
You may want to save output somewhere because it contains your Evmos address and other usefull information.

Run the init script

Init you validator where [myvalidator] is your validator custom name which will be publicly visible
  
```evmosd init [myvalidator] --chain-id point_10721-1```

Copy `genesis.json` and `config.toml` files from this repository https://github.com/pointnetwork/point-chain-config/tree/main/testnet-xNet-Triton-1 into `~/.evmosd/config`

Validate it:
  
```evmosd validate-genesis```
  
## Run the Node

Then run the node and wait for fully sync:
  
```evmosd start --json-rpc.enable=true --json-rpc.api "eth,txpool,personal,net,debug,web3"```

You can run this command to see status of your node:
  
```evmosd status```

You will get the "latest_block_height" of your node.
  
To see current block height of blockchain run:

```curl  http://xnet-neptune-1.point.space:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'```

The result is in hexadecimal, just convert to decimal and see how far are you from full sync.

## Sending your first transaction

Now you need to send funds to your validator address. As mentioned, you should have received an airdrop of 1024 XPOINT if you filled in the form. To see them, you can import the private key into a wallet like Metamask (not a good idea for mainnet security, but ok for testnet tokens).

Then you need to add XNet-Triton into Metamask:

```
Network Title: Point XNet Triton
RPC URL: http://xnet-neptune-1.point.space:8545/
Chain ID: 10721
SYMBOL: XPOINT
```

In order to import the wallet in your metamask you will need the private key. You can get it with this command:

```evmosd keys unsafe-export-eth-key validatorkey --keyring-backend file```

Now let’s import the wallet in metamask. Go to the import account section, select type “Private key” and insert the private key you got from the command above.

Now that you have an account you need to get some point to stake and run your validator. (contact point team)

## Stake XPOINT and Join as a Validator

Once the node is fully synced, and you got some point to stake check your balance in the node you 
will see your balance in Metamask or you can check your balance with this command:

```evmosd query bank balances  <evmosaddress>```

If you have enough balance stake your assets and check the transaction:

```
evmosd tx staking create-validator  
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

You will have to provide your keystore password and approve the transaction.

If everything works ok you will get a txhash. You can check the status of the tx: ```evmosd query tx <txhash>```

Transaction receipt may contain errors, so please check if there are any or if it's live.

If the transaction was correct you should become part of the validators set. Check your pubkey first:

```evmosd tendermint show-validator```

You will see a key there, you can identify your node among other validators using that key:

```evmosd query tendermint-validator-set```

There you will find more info like your VotingPower that should be bigger than 0. Also you can check your VotingPower by running:

```evmosd status```

## What's Next?

Check out extra documentation for validators:

- https://hub.cosmos.network/main/validators/validator-faq.html#
- https://docs.evmos.org/validators/overview.html

If you experience any issues, join our Discord: https://pointnetwork.io/discord

And ask in #validators channel. This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

Share any feedback, questions, and ideas there!
