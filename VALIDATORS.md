# Join Point-XNet-Triton as a Validator

DISCLAIMER: THE DOCUMENT IS PROVIDED ON "AS IS" AND “AS DEVELOPED” BASIS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE DOCUMENT.

Following this document and/or performing validation activities requires highly experienced DevOps engineers that possess necessary knowledge how to run validators. You are fully responsible for your interaction with validator functionality and running any type of commands, we will not and shall not be liable to you for any omissions, defects, bugs, limitations and delays in the validator functionality or any other related software.

⚠️ This is not the tutorial for most Point Network users! Do not attempt to run the commands from your personal computer, just in case something goes wrong!

## Table of Contents

* [Overview](#overview)
* [Rewards](#rewards)
* [Prerequisites](#prerequisites)
* [Initialize the Node](#initialize-the-node)
* [Run the Node](#run-the-node)
* [Sending your first transaction](#sending-your-first-transaction)
* [Stake XPOINT and Join as a Validator](#stake-xpoint-and-join-as-a-validator)
* [What's Next?](#whats-next)
* [Useful Commands](#useful-commands)

## Overview

This document describes step-by-step instructions on joining Point-XNet-Neptune testnet as a validator.

Validators have the responsibility to keep the network operational 24/7. Do not attempt to join the testnet (and especially mainnet) if you don’t have enough experience. For example, if you install it on your laptop, join as a validator, and then close the laptop, the network will penalize you for being offline by slashing your stake (+the network quality might degrade).

If you have any questions, join our Discord: https://pointnetwork.io/discord and ask in #validators channel (in order to see #validators channel, you should add yourself a Validator role at #roles). This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

Evmos is based on Cosmos SDK (which in turn is based on Tendermint), so if you know Cosmos commands, most of them will work here too.

## Rewards

Validators receive rewards according to their stake, but because everyone received the same amount (1024 XPOINT), your testnet rewards will be multiplied by the same factor shared by everyone, **and** also by the amount of real POINT you will have at the mainnet launch. Here's what it means.

Imagine if we were already on the mainnet. Let's say there are two active validators: you and someone else. You stake with 1000 POINT, and they stake with 500 POINT (and everything else is equal - uptime, etc.) Obviously, you would receive 2x the rewards than the other validator on mainnet.

But right not we don't know how many real POINTs all of you would have at the mainnet start. So right now everyone gets equal amounts: 1024 XPOINT, and equal rewards (if all else considered equal - if your uptime and XPOINT stake is the same etc.)

So to get to the real number, on September 1 we will multiply the rewards by a constant which will be the same for everyone, and by how many real POINT validators will have, to simulate as if they put this amount at stake.

_Q: Why tie rewards to real POINTs and not give every validator equal amounts for participation?_

_A: If we did this, some could spawn hundreds of validator instances to claim their rewards, and not only this would impact our funds and make it unfair to others, but slow down the network. Multiplying it by a real stake solves this - if someone has 5000 POINTs, they can create 1 or 5000 validators, it doesn't matter - they would have to split the real stake too when they claim on Sep 1, which would also split the rewards instead of accumulating them as it would have been with a constant._

## Prerequisites

Minimum hardware requirements: https://docs.evmos.org/validators/overview.html#hardware

Most of the commands here are provided for and tested on Ubuntu Server 22.04 LTS, so change them accordingly (pacman instead of apt-get if you’re on Arch for example, or brew on Mac OS)

The commands have to be run in a virtual session (like `tmux` or `screen`), otherwise if you disconnect from the server, the node will be killed.

Run ```tmux``` to start the virtual session (if you disconnect and connect to your server again, use ```tmux attach``` to attach to that running session)

Ensure Go is installed (Note: Requires Go 1.18+):

```go version```

If Golang is not installed, install it using official tutorial: https://go.dev/doc/install.

Also to build node you would have to have `make` installed:

```sudo apt-get install build-essential```

Pull the repository of the point chain:

```git clone https://github.com/pointnetwork/point-chain```

Go inside the folder:

```cd point-chain```

Switch to the triton branch:

```git checkout xnet-triton```

Compile the node from the sources:

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

Init you validator where [myvalidator] is your validator custom name which will be publicly visible

```evmosd init myvalidator --chain-id point_10721-1```

Once you've initialized your validator is really important to back up the validator keys. They were generated inside ~/.evmosd/config/priv_validator_key.json
Save this file and don't share it. It's the id of your validator and you will need it for reinstallation or migration of the node

Copy `genesis.json` and `config.toml` files from this repository https://github.com/pointnetwork/point-chain-config/tree/main/testnet-xNet-Triton-1 into `~/.evmosd/config`:

`wget https://raw.githubusercontent.com/pointnetwork/point-chain-config/main/testnet-xNet-Triton-1/config.toml`

`wget https://raw.githubusercontent.com/pointnetwork/point-chain-config/main/testnet-xNet-Triton-1/genesis.json`

`mv config.toml genesis.json ~/.evmosd/config/`

Validate it:

```evmosd validate-genesis```

## Run the Node

Then run the node and wait for fully sync using bash script:

```./start.sh``` from repository root folder.

If you want it to also respond to the RPC commands, you can instead run:

```evmosd start --json-rpc.enable=true --json-rpc.api "eth,txpool,personal,net,debug,web3"```

Now that the node has started, you cannot type any commands in your terminal. But thankfully, your virtual session supports several windows. So if you're on tmux, you can press Ctrl+b and then letter "c" to create a new tab.

Then you can switch between the tabs like this: Ctrl+b and then the window ID (try window 0 where your node runs, and window 1 where you can type commands)

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

Run ```evmosd keys list --keyring-backend file```, and you will see a list of keys attached to your node. Look at the one which has the name `validatorkey`, and note its address (it should be in Cosmos format and start with `evmos` prefix).

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
--amount=1000000000000000000000apoint \
--pubkey=$(evmosd tendermint show-validator) \
--moniker="<myvalidator>" \
--chain-id=point_10721-1 \
--commission-rate="0.10" \
--commission-max-rate="0.20" \
--commission-max-change-rate="0.01" \
--min-self-delegation="1000000000000000000000" \
--gas="400000" \
--gas-prices="0.025apoint" \
--from=validatorkey \
--keyring-backend file
```

(Note the amount: it's in apoint (which is 1/1e18 XPOINT). 1000000000000000000000apoint is 1000 XPOINT (when you remove 18 zeroes at the end). If you decide to adjust the amount, don't forget to adjust `min-self-delegation` flag too.)

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

Please post on Discord channel #validators when you succeed! https://pointnetwork.io/discord (in order to see #validators channel, you should add yourself a Validator role at #roles)

And if you have any questions, ask in #validators channel. This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

Also, check out extra documentation for validators:

- https://hub.cosmos.network/main/validators/validator-faq.html#
- https://docs.evmos.org/validators/overview.html

Share any feedback, questions, and ideas there!

## Useful commands

* How to run the node as a service: https://medium.com/@anttiturunen/running-point-validator-as-a-service-d8e4b0391540

* Check the balance of an evmos-formatted address: `evmosd query bank balances <evmosaddress>`

* Check if your validator is active: `evmosd query tendermint-validator-set | grep "$(evmosd tendermint show-address)"` (if the output is non-empty, you are a validator)

* See the slashing status: `evmosd query slashing signing-info $(evmosd tendermint show-validator)` Jailed until year 1970 means you are not jailed!

* If the slashing status says you're jailed for downtime, you can unjail yourself once you're back online by first, starting the node, making sure it's synced to the last block, and then running: `evmosd tx slashing unjail --from=validatorkey --chain-id=point_10721-1`. Run `evmosd status` and `evmosd query tendermint-validator-set | grep "$(evmosd tendermint show-address)"` to confirm you're unjailed.

* Halting Your Validator:

  * When attempting to perform routine maintenance or planning for an upcoming coordinated upgrade, it can be useful to have your validator systematically and gracefully halt. You can achieve this by either setting the `halt-height` to the height at which you want your node to shutdown or by passing the `--halt-height` flag to `evmosd`. The node will shutdown with a zero exit code at that given height after committing the block.

