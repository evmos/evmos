---
order: 4
---

# Submitting a Governance Proposal

If you have a final draft of your proposal ready to submit, you may want to push your proposal live on the testnet first. These are the three primary steps to getting your proposal live on-chain.

1. (**Optional**) [Hosting supplementary materials](#hosting-supplementary-materials) for your proposal with IPFS (InterPlanetary File System)
2. [Formatting the JSON file](#formatting-the-json-file-for-the-governance-proposal) for the governance proposal transaction that will be on-chain
3. [Sending the transaction](#sending-the-transaction-that-submits-your-governance-proposal) that submits your governance proposal on-chain

## Hosting supplementary materials

In general we try to minimize the amount of data pushed to the blockchain.
Hence, detailed documentation about a proposal is usually hosted on a separate
censorship resistant data-hosting platform, like IPFS.

Once you have drafted your proposal, ideally as a Markdown file, you
can upload it to the IPFS network:

1. either by [running an IPFS node and the IPFS software](https://ipfs.io), or
2. using a service such as [https://pinata.cloud](https://pinata.cloud)

Ensure that you "pin" the file so that it continues to be available on the network. You should get a URL like this: https://ipfs.io/ipfs/QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD
The value QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD is called the `CID` of
your file - it is effectively the file's hash.

If you uploaded a markdown file, you can use the IPFS markdown viewer to render
the document for better viewing. Links for the markdown viewer look like
`https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/<CID>`, where `<CID>` is your CID. For instance the link above would be:
[https://ipfs.io/ipfs/QmTk...HSWD](https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/QmbkQNtCAdR1CNbFE8ujub2jcpwUcmSRpSCg8gVWrTHSWD)

Share the URL with others and verify that your file is publicly accessible.

The reason we use IPFS is that it is a decentralized means of storage, making it resistant to censorship or single points of failure. This increases the likelihood that the file will remain available in the future.

## Formatting the JSON file for the governance proposal

Prior to sending the transaction that submits your proposal on-chain, you must create a JSON file. This file will contain the information that will be stored on-chain as the governance proposal. Begin by creating a new text (.txt) file to enter this information. Use [these best practices](./best-practices.md) as a guide for the contents of your proposal. When you're done, save the file as a .json file. See the examples that follow to help format your proposal.

Each proposal type is unique in how the JSON should be formatted.
See the relevant section for the type of proposal you are drafting:

- [Text Proposals](./formatting.md)
- [Community Pool Spend Proposals](./formatting.md)
- [Parameter Change Proposals](./formatting.md)

Once on-chain, most people will rely upon network explorers to interpret this information with a graphical user interface (GUI).

**Note**: In future, this formatting [may be changed to be more standardized](https://github.com/cosmos/cosmos-sdk/issues/5783) with other the types of governance proposals.

## Sending the transaction that submits your governance proposal

For information on how to use evmosd (the command line interface) to submit an on-chain proposal through the governance module, please refer to the [evmosd resource](../resources/evmosd.md) for Evmos documentation.

### Walkthrough example

This is the command format for using evmosd (the command-line interface) to submit your proposal on-chain:

```
evmosd tx gov submit-proposal \
  --title=<title> \
  --description=<description> \
  --type="Text" \
  --deposit="1000000aevmos" \
  --from=<name> \
  --chain-id=<chain_id>
```

If `<proposal type>` is left blank, the type will be a Text proposal. Otherwise, it can be set to `param-change` or `community-pool-spend`. Use `--help` to get more info from the tool.

For instance, this is the complete command that I could use to submit a **testnet** parameter-change proposal right now:
`evmosd tx gov submit-proposal param-change param.json --from gavin --chain-id gaia-13007 --node 45.77.218.219:26657`

This is the complete command that I could use to submit a **mainnet** parameter-change proposal right now:
`evmosd tx gov submit-proposal param-change param.json --from gavin --gas 500000 --fees 7500aevmos --chain-id cosmoshub-3 --node cosmos-node-1.figment.network:26657`

1. `evmosd` is the command-line interface client that is used to send transactions and query Evmos
2. `tx gov submit-proposal param-change` indicates that the transaction is submitting a parameter-change proposal
3. `--from gavin` is the account key that pays the transaction fee and deposit amount
4. `--gas 500000` is the maximum amount of gas permitted to be used to process the transaction
   - the more content there is in the description of your proposal, the more gas your transaction will consume
   - if this number isn't high enough and there isn't enough gas to process your transaction, the transaction will fail
   - the transaction will only use the amount of gas needed to process the transaction
5. `--fees` is a flat-rate incentive for a validator to process your transaction
   - the network still accepts zero fees, but many nodes will not transmit your transaction to the network without a minimum fee
   - many nodes (including the Figment node) use a minimum fee to disincentivize transaction spamming
   - 7500aevmos is equal to 0.0075 EVMOS
6. `--chain-id cosmoshub-3` is Evmos 3. For current and past chain-id's, please look at the [cosmos/mainnet resource](https://github.com/cosmos/mainnet)
   - the testnet chain ID is [gaia-13007](https://hubble.figment.network/cosmos/chains/gaia-13007). For current and past testnet information, please look at the [testnet repository](https://github.com/cosmos/testnets)
7. `--node cosmos-node-1.figment.network:26657` is using Figment Networks' node to send the transaction to Evmos 3 network

**Note**: be careful what you use for `--fees`. A mistake here could result in spending hundreds or thousands of ATOMs accidentally, which cannot be recovered.

### Verifying your transaction

After posting your transaction, your command line interface (evmosd) will provide you with the transaction's hash, which you can either query using evmosd or by searching the transaction hash using [Mintscan](https://www.mintscan.io/cosmos/txs/0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57). The hash should look something like this: `0506447AE8C7495DE970736474451CF23536DF8EA837FAF1CF6286565589AB57`

### Troubleshooting a failed transaction

There are a number of reasons why a transaction may fail. Here are two examples:

1. **Running out of gas** - The more data there is in a transaction, the more gas it will need to be processed. If you don't specify enough gas, the transaction will fail.
2. **Incorrect denomination** - You may have specified an amount in 'evmos' instead of 'aevmos', causing the transaction to fail.

If you encounter a problem, try to troubleshoot it first, and then ask for help on the All in Bits Cosmos forum: [https://forum.cosmos.network/c/governance](https://forum.cosmos.network/c/governance). We can learn from failed attempts and use them to improve upon this guide.

### Depositing funds after a proposal has been submitted

Sometimes a proposal is submitted without having the minimum token amount deposited yet. In these cases you would want to be able to deposit more tokens to get the proposal into the voting stage. In order to deposit tokens, you'll need to know what your proposal ID is after you've submitted your proposal. You can query all proposals by the following command:

```
evmosd q gov proposals
```

If there are a lot of proposals on the chain already, you can also filter by your own address. For the proposal above, that would be:

```
evmosd q gov proposals --depositor evmos1hxv7mpztvln45eghez6evw2ypcw4vjmsmr8cdx
```

Once you have the proposal ID, this is the command to deposit extra tokens:

```
evmosd tx gov deposit <proposal-id> <deposit> --from <name>
```

In our case above, the `<proposal-id>` would be 59 as queried earlier.
The `<deposit>` is written as `500000aevmos`, just like the example above.

### Submitting your proposal to the testnet

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons:

1. To see what the proposal description will look like
2. To signal that your proposal is about to go live on the mainnet
3. To share what the proposal will look like in advance with stakeholders
4. To test the functionality of the governance features

Submitting your proposal to the testnet increases the likelihood that you will discover a flaw before deploying your proposal on mainnet. A few things to keep in mind:

- you'll need testnet tokens for your proposal (ask around for a faucet)
- the parameters for testnet proposals are different (eg. voting period timing, deposit amount, deposit denomination)
- the deposit denomination is in 'atevmos' instead of 'aevmos'
