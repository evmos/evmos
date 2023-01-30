<!--
order: 8
-->

# Automated Coin Conversion

Learn how to the Automated Coin Conversion feature works. {synopsis}

In their [ERC-20](https://ethereum.org/en/developers/docs/standards/tokens/erc-20/) representations,
assets can be used to interact with dApps using the EVM.
This standard allows developers to build applications that are interoperable with other products and services.
In their Native Coin representation, they can be transferred
between accounts on Evmos and other Cosmos chains using IBC.
They cannot, however, be used to interact with dApps on Evmos,
as Native Coins are not supported by the EVM
since they don’t implement the ERC-20 standard.

In order to reduce end-user complexity,
Evmos should only allow single-token representation use between IBC Coin and ERC-20s.
Consequently, the Evmos team developed the Automated Coin Conversion feature to achieve this goal.
It converts incoming IBC vouchers to ERC-20s and modifies outgoing IBC transfers to convert ERC-20s to IBC Coins.
This automated conversion occurs if, and *only* if, the appropriate token mapping was registered through governance.
If the token pair is not registered, the IBC coin will be left as is.

## Outbound transactions

Your users may want to move their ERC-20 tokens from Evmos onto other Cosmos chains.
The automated coin conversion feature simplifies this operation, because it enables you
to send ERC-20 tokens via an IBC transfer in a single step.
To do so, there is no need to make any changes on your IBC transfer logic.
You only need to ensure that the corresponding denomination is passed as a parameter.
For example, if you want to transfer the ERC-20 representation of the `uosmo` token on Evmos,
specifying the corresponding denomination (`Token.Denom = "uosmo"`) on the `MsgTransfer` struct will suffice.
Another example, if you want to transfer Wrapped Bitcoin on Axelar (axlWBTC),
you could achieve this by using `Token.Denom = "ibc/C834CD421B4FD910BBC97E06E86B5E6F64EA2FE36D6AE0E4304C2E1FB1E7333C"`,
as that is the denomination in the registered token pair.
The same applies to any ERC-20 token that is not a representation of a Native Coin on other Cosmos chains.
For example, if we want to send an ERC-20 token called `TestCoin` via IBC,
use `Token.Denom = "erc20/<test-coin-contract-address>"`.
Before transferring ERC-20 tokens via IBC, make sure you
[register the ERC-20 token](https://docs.evmos.org/developers/guides/erc20_registration.html) for the conversion.
Under the hood, the protocol will automatically make the conversion from ERC-20 token to IBC coin
and perform the transfer to the desired Cosmos chain.

:::tip
**Note**: In case Evmos is not the source chain of the sent IBC coin,
you will have to specify the corresponding IBC denom
(e.g. `ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518`).
:::

```go
type MsgTransfer struct {
	// the port on which the packet will be sent
	SourcePort string `protobuf:"bytes,1,opt,name=source_port,json=sourcePort,proto3" json:"source_port,omitempty" yaml:"source_port"`
	// the channel by which the packet will be sent
	SourceChannel string `protobuf:"bytes,2,opt,name=source_channel,json=sourceChannel,proto3" json:"source_channel,omitempty" yaml:"source_channel"`
	// the tokens to be transferred
	Token types.Coin `protobuf:"bytes,3,opt,name=token,proto3" json:"token"`
	// the sender address
	Sender string `protobuf:"bytes,4,opt,name=sender,proto3" json:"sender,omitempty"`
	// the recipient address on the destination chain
	Receiver string `protobuf:"bytes,5,opt,name=receiver,proto3" json:"receiver,omitempty"`
	// Timeout height relative to the current block height.
	// The timeout is disabled when set to 0.
	TimeoutHeight types1.Height `protobuf:"bytes,6,opt,name=timeout_height,json=timeoutHeight,proto3" json:"timeout_height" yaml:"timeout_height"`
	// Timeout timestamp in absolute nanoseconds since unix epoch.
	// The timeout is disabled when set to 0.
	TimeoutTimestamp uint64 `protobuf:"varint,7,opt,name=timeout_timestamp,json=timeoutTimestamp,proto3" json:"timeout_timestamp,omitempty" yaml:"timeout_timestamp"`
	// optional memo
	Memo string `protobuf:"bytes,8,opt,name=memo,proto3" json:"memo,omitempty"`
}

type Coin struct {
	Denom  string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Amount Int    `protobuf:"bytes,2,opt,name=amount,proto3,customtype=Int" json:"amount"`
}
```

## Inbound transactions

Your users may want to move IBC Coins from other Cosmos chains onto Evmos.
To use these IBC coins on dApps deployed on Evmos, they need an ERC-20 representation of these.
The automated coin conversion feature automatically converts the incoming IBC coins into their ERC-20 representation.
In this way, you don't need to manually convert the incoming IBC coins into ERC-20 tokens.
As a result, your users can use the IBC coins as ERC-20 tokens as soon as they arrive to their wallets.

It should be considered that only registered token pairs are converted.
If the token pair is not registered,
users will receive the corresponding IBC coin on their wallet without any further changes.

:::tip
**Note**: If your users have IBC coins on Evmos already,
and they receive an IBC transfer in the denomination of an already registered token pair,
their **whole balance** will be converted to the ERC20 format
(i.e the current balance plus the transfer amount).
:::

## FAQ

### How do I send an ERC-20 via IBC?

With the new automated coin conversion feature, you can send ERC-20 via IBC right away.
The conversion step is done automatically under the hood.
To do this operation you only need to specify the corresponding denomination on the `MsgTransfer` struct.
For example, if we want to send an ERC-20 token called `TestCoin` via IBC,
use `Token.Denom = "erc20/<test-coin-contract-address>"`.
Keep in mind that to perform this operation, you need to
[register the token pair](https://docs.evmos.org/developers/guides/erc20_registration.html) previously.

### Can I send WEVMOS to other chains?

WEVMOS transfers are not supported at the moment.
However, you can unwrap manually the WEVMOS tokens
using the [Evmos dashboard](https://app.evmos.org/assets) or [Diffusion](https://app.diffusion.fi/).
Then you can perform a regular IBC transfer using the EVMOS tokens.

### Does automated coin conversion apply to all coins?

The automated coin conversion covers all IBC coins and ERC-20 tokens
as long as the appropriate token mapping was registered through governance
([guide to register an ERC-20 token](https://docs.evmos.org/developers/guides/erc20_registration.html)).
If the token pair is not registered, the IBC coin will be left as is.
Additionally, EVMOS token conversion is not automated.
Considering that the EVMOS token is used for staking and paying gas fees,
the team decided to exclude the native token automated conversion.
Thus, the user experience is not undermined by this feature.

### How do I convert the EVMOS token to ERC-20?

The conversion from EVMOS token to WEVMOS is not automated.
If you want to convert EVMOS tokens into its ERC-20 representation,
you will need to use [the Assets page](https://app.evmos.org/assets).

### Do I still need to use [the Assets page](https://app.evmos.org/assets)?

Yes! If you want to convert EVMOS tokens into their ERC-20 representation,
you will need to do it manually on [the Assets page](https://app.evmos.org/assets).
EVMOS token automated conversion was excluded in this feature
because it is used for staking, governance and paying for gas on the EVM.
Additionally, you can still manually convert IBC coins to ERC-20 tokens.
On top of that, the assets page allows you to see all your token balances.
