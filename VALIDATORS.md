# Join Point-XNet-Uranus as a Validator

DISCLAIMER: THE DOCUMENT IS PROVIDED ON "AS IS" AND “AS DEVELOPED” BASIS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE DOCUMENT.

Following this document and/or performing validation activities requires highly experienced DevOps engineers that possess necessary knowledge how to run validators. You are fully responsible for your interaction with validator functionality and running any type of commands, we will not and shall not be liable to you for any omissions, defects, bugs, limitations and delays in the validator functionality or any other related software.

⚠️ This is not the tutorial for most Point Network users! Do not attempt to run the commands from your personal computer, just in case something goes wrong!

## Table of Contents

- [Join Point-XNet-Uranus as a Validator](#join-point-xnet-uranus-as-a-validator)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Prerequisites](#prerequisites)
  - [Initialize the Node](#initialize-the-node)
  - [Run the Node](#run-the-node)
  - [Sending your first transaction](#sending-your-first-transaction)
    - [Add custom network](#add-custom-network)
    - [Add the wallet with your 1024 XPOINT](#add-the-wallet-with-your-1024-xpoint)
    - [Find out which address is your validator wallet](#find-out-which-address-is-your-validator-wallet)
    - [Fund the validator](#fund-the-validator)
  - [Stake XPOINT and Join as a Validator](#stake-xpoint-and-join-as-a-validator)
  - [What's Next?](#whats-next)
  - [Useful commands](#useful-commands)

## Overview

This document describes step-by-step instructions on joining Point-XNet-uranus testnet as a validator.

Validators have the responsibility to keep the network operational 24/7. Do not attempt to join the testnet (and especially mainnet) if you don’t have enough experience. For example, if you install it on your laptop, join as a validator, and then close the laptop, the network will penalize you for being offline by slashing your stake (+the network quality might degrade).

If you have any questions, join our Discord: https://pointnetwork.io/discord and ask in #validators channel (in order to see #validators channel, you should add yourself a Validator role at #roles). This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

point is based on Cosmos SDK (which in turn is based on Tendermint), so if you know Cosmos commands, most of them will work here too.

## Prerequisites

Minimum hardware requirements: https://docs.point.org/validators/overview.html#hardware

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

Switch to the uranus branch:

```git checkout xnet-uranus```

Compile the node from the sources:

```make install```

_Note: Point Chain is a fork of point, and by default the working directory is ~/.point. Make sure you don’t already have data for another point node on the device you’re running the validator from._

## Initialize the Node

Check if pointd command is available for you. If you see ```pointd: command not found``` message then export path for this command:

```export PATH=$PATH:$(go env GOPATH)/bin```

Configure your validator key:

```pointd config keyring-backend file```

```pointd config chain-id point_10731-1```


Generate a new key/mnemonic for validator: ```pointd keys add validatorkey --keyring-backend file```
You may want to save output somewhere because it contains your point address and other usefull information.

Init you validator where [myvalidator] is your validator custom name which will be publicly visible

```pointd init myvalidator --chain-id point_10731-1```

Once you've initialized your validator is really important to back up the validator keys. They were generated inside ~/.pointd/config/priv_validator_key.json
Save this file and don't share it. It's the id of your validator and you will need it for reinstallation or migration of the node

Copy `genesis.json` and `config.toml` files from this repository https://github.com/pointnetwork/point-chain-config/tree/main/testnet-xNet-Uranus-1 into `~/.pointd/config`:

`wget https://raw.githubusercontent.com/pointnetwork/point-chain-config/main/testnet-xNet-Uranus-1/config.toml`

`wget https://raw.githubusercontent.com/pointnetwork/point-chain-config/main/testnet-xNet-Uranus-1/genesis.json`

`mv config.toml genesis.json ~/.pointd/config/`

Validate it:

```pointd validate-genesis```

## Run the Node

Then run the node and wait for fully sync using bash script:

```./start.sh``` from repository root folder.

If you want it to also respond to the RPC commands, you can instead run:

```pointd start --json-rpc.enable=true --json-rpc.api "eth,txpool,personal,net,debug,web3"```

Now that the node has started, you cannot type any commands in your terminal. But thankfully, your virtual session supports several windows. So if you're on tmux, you can press Ctrl+b and then letter "c" to create a new tab.

Then you can switch between the tabs like this: Ctrl+b and then the window ID (try window 0 where your node runs, and window 1 where you can type commands)

You can run this command to see status of your node:

```pointd status```

You will get the "latest_block_height" of your node.

To see current block height of blockchain run:

```curl  http://xnet-uranus-1.point.space:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'```

The result is in hexadecimal, just convert to decimal and see how far are you from full sync.

## Sending your first transaction

### Add custom network

Now while you're waiting for the node to sync, you need to send funds to your validator address. As mentioned, you should have received an airdrop of 1024 XPOINT if you filled in the form. To see them, you can import the private key into a wallet like Metamask (not a good idea for mainnet security, but ok for testnet tokens).

Then you need to add XNet-Uranus into Metamask:

```
Network Title: Point XNet Uranus
RPC URL: https://xnet-uranus-1.point.space/
Chain ID: 10731
SYMBOL: XPOINT
```

### Add the wallet with your 1024 XPOINT

Remember the wallet you sent to us to be funded? In the form? It now has 1024 XPOINT.

Import the wallet with the private key into your wallet (e.g. Metamask), and you should see 1024 XPOINT there. But this is your fund wallet, not validator wallet.

### Find out which address is your validator wallet

point has two wallet formats: Cosmos format, and Ethereum format. Cosmos format starts with `point` prefix, and Ethereum format starts with `0x`. Most people don't need to know about Cosmos format, but validators should have a way to change from one to another.

Run ```pointd keys list --keyring-backend file```, and you will see a list of keys attached to your node. Look at the one which has the name `validatorkey`, and note its address (it should be in Cosmos format and start with `point` prefix).

(In most cases it is not needed, but if something goes wrong and if you ever want to import your validator wallet in your Metamask you will need the private key. You can get it with this command: `pointd keys unsafe-export-eth-key validatorkey --keyring-backend file`)

Use this tool to convert it to Ethereum format: https://point.me/utils/tools

This is your validator address in Ethereum format.

### Fund the validator

Finally, use the wallet to send however much you need from your fund address to the validator address (you can send all 1024 or choose a different strategy).

## Stake XPOINT and Join as a Validator

Now you have to wait for the node to fully sync, because otherwise it will not find your.

Once the node is fully synced, and you got some XPOINT to stake, check your balance in the node, you
will see your balance in Metamask or you can check your balance with this command:

```pointd query bank balances  <pointaddress>```

If you have enough balance stake your assets and check the transaction:

```
pointd tx staking create-validator \
--amount=1000000000000000000000apoint \
--pubkey=$(pointd tendermint show-validator) \
--moniker="<myvalidator>" \
--chain-id=point_10731-1 \
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

If everything works ok you will get a txhash. You can check the status of the tx: ```pointd query tx <txhash>```

Transaction receipt may contain errors, so please check if there are any or if it's live. You can use the explorer or ask the node to provide receipt.

If the transaction was correct you should instantly become part of the validators set. Check your pubkey first:

```pointd tendermint show-validator```

You will see a key there, you can identify your node among other validators using that key:

```pointd query tendermint-validator-set```

There you will find more info like your VotingPower that should be bigger than 0. Also you can check your VotingPower by running:

```pointd status```

## What's Next?

Please post on Discord channel #validators when you succeed! https://pointnetwork.io/discord (in order to see #validators channel, you should add yourself a Validator role at #roles)

And if you have any questions, ask in #validators channel. This is the channel where we will sync our testnet efforts and communicate with each other about what's happening.

Also, check out extra documentation for validators:

- https://hub.cosmos.network/main/validators/validator-faq.html#
- https://docs.point.org/validators/overview.html

Share any feedback, questions, and ideas there!

## Useful commands

* How to run the node as a service: https://medium.com/@anttiturunen/running-point-validator-as-a-service-d8e4b0391540

* Check the balance of an point-formatted address: `pointd query bank balances <pointaddress>`

* Check if your validator is active: `pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"` (if the output is non-empty, you are a validator)

* See the slashing status: `pointd query slashing signing-info $(pointd tendermint show-validator)` Jailed until year 1970 means you are not jailed!

* If the slashing status says you're jailed for downtime, you can unjail yourself once you're back online by first, starting the node, making sure it's synced to the last block, and then running: `pointd tx slashing unjail --from=validatorkey --chain-id=point_10731-1`. Run `pointd status` and `pointd query tendermint-validator-set | grep "$(pointd tendermint show-address)"` to confirm you're unjailed.

* Halting Your Validator:

  * When attempting to perform routine maintenance or planning for an upcoming coordinated upgrade, it can be useful to have your validator systematically and gracefully halt. You can achieve this by either setting the `halt-height` to the height at which you want your node to shutdown or by passing the `--halt-height` flag to `pointd`. The node will shutdown with a zero exit code at that given height after committing the block.

