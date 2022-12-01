<!--
order: 8
-->

# Automated Coin Conversion

Learn how to the Automated Coin Conversion feature works. {synopsis}

In their ERC-20 representation, assets can be used to interact with dApps using the EVM. This standard allows developers to build applications that are interoperable with other products and services. In their Native Coin representation, they can be transferred between accounts on Evmos and other Cosmos chains using IBC. They cannot, however, be used to interact with dApps on Evmos, as Native Coins are not supported by the EVM since they don’t implement the ERC-20 standard.

In order to reduce end-user complexity, Evmos should only allow single-token representation use between IBC Coin and ERC-20s. Consequently, the Evmos team developed the Automated Coin Conversion feature to achieve this goal. It consists on converting incoming IBC vouchers to ERC-20s and modifying outgoing IBC transfers to convert ERC-20s to IBC Coins. This automated conversion occurs if, and only if, the appropriate token mapping was registered through governance. If the token pair is not registered, the IBC coin will be left as is.

Please read on for further understanding of this feature scope and functionality.

## Outbound transactions

As an Evmos user, you may want to move your ERC-20 tokens onto another Cosmos chains. You may want to do this to use your tokens on dApps on other Cosmos chains. The automated coin conversion feature makes this operation smooth. You can send ERC-20 tokens via an IBC transfer with a single step. You can perform this operation using the Evmos [IBC transfer page](https://app.evmos.org/transfer). Under the hood, the protocol will atomatically make the conversion from ERC-20 token to IBC coin and perform the transfer to the desired Cosmos chain.

## Inbound transactions

As an Evmos user, you may want to move IBC Coins from other Cosmos chains onto Evmos. To use these IBC coins on dApps deployed on Evmos, you need an ERC-20 representation of these. The automated coin conversion feature automatically converts the incoming IBC coins into their ERC-20 representation. In this way, you don't need to manually convert your IBC coins into ERC-20 tokens. As a result, you can use the IBC coins as ERC-20 tokens as soon as they arrive to your Evmos wallet.

The user should note that only the registered token pairs are converted. If the token pair is not registered, you will receive the corresponding IBC coin on your wallet without any further change.

:::tip
**Note**: If you have some IBC coins on Evmos already, and the token pair is registered, when you receive an IBC transfer of this denomination, the **whole balance** will be converted (the current balance plus the transfer amount).
:::

## FAQ

### How do I send an ERC-20 via IBC?

The [IBC transfer page](https://app.evmos.org/transfer) allows you to perform IBC transfers of either ERC-20, IBC coins or Evmos tokens. With the new automated coin conversion feature, you can send ERC-20 via IBC right away. The conversion step is done automatically under the hood. Users don't need to manually convert the ERC-20 tokens into IBC coins anymore to perform this operation.

### Can I send WEVMOS to other chains?

Yes! The automated coin conversion feature allows you to send ERC-20 tokens via IBC to other chains. This includes WEVMOS tokens. You can perform this operation using the [IBC transfer page](https://app.evmos.org/transfer).

### Does automated coin conversion apply to all coins?

The automated coin conversion covers all IBC coins and ERC-20 tokens as long as the appropriate token mapping was registered through governance. If the token pair is not registered, the IBC coin will be left as is. Additionally, Evmos token conversion is not automated. Considering that the Evmos token is used for staking and paying gas fees, the team decided to exclude the native token automated conversion. Thus, the user experience is not undermined by this feature.

### How do I convert the Evmos token to ERC-20?

The conversion from EVMOS token to WEVMOS is not automated. If you want to convert EVMOS tokens into its ERC-20 representation, you will need to use [the Assets page](https://app.evmos.org/assets).

### Do I still need to use [the Assets page](https://app.evmos.org/assets)?

Yes! If you want to convert Evmos tokens into their ERC-20 representation, you will need to do it manually on [the Assets page](https://app.evmos.org/assets). Evmos token automated conversion was excluded in this feature to avoid damaging user experience. Additionally, you can still convert manually IBC coins to ERC-20 tokens. On top of that, the assets page allows you to see all your token balances.
