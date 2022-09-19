<!--
order: 3
-->

# State Transitions

The erc20 modules allows for two types of registration state transitions. Depending on how token pairs are registered, with `RegisterCoinProposal` or `RegisterERC20Proposal`, there are four possible conversion state transitions.

## Token Pair Registration

### 1. Register Coin

A user registers a native Cosmos Coin. Once the proposal passes (i.e is Approvald by governance), the ERC20 module uses a factory pattern to deploy an ERC20 token contract representation of the Cosmos Coin. Note that the native Evmos coin cannot be registered, as any coin including "evm" in its denomination cannot be registered. Instead the Evmos token can be converted by Nomand's wrapped Evmos (WEVMOS) contract.

1. User submits a `RegisterCoinProposal`
2. Validators of the Evmos Hub vote on the proposal using `MsgVote` and proposal passes
3. If Cosmos coin or IBC voucher exist on the bank module supply, create the [ERC20 token contract](https://github.com/evmos/evmos/blob/main/contracts/ERC20MinterBurnerDecimals.sol) on the EVM based on the ERC20Mintable ([ERC20Mintable by openzeppelin](https://github.com/OpenZeppelin/openzeppelin-contracts/tree/master/contracts/token/ERC20)) interface
    - Initial supply: 0
    - Token details (Name, Symbol, Decimals, etc) are derived from the bank module `Metadata` field on the proposal content.

### 2. Register ERC20

A user registers a ERC20 token contract that is already deployed on the EVM module. Once the proposal passes (i.e is approved by governance), the ERC20 module creates a Cosmos coin representation of the ERC20 token.

1. User submits a `RegisterERC20Proposal`
2. Validators of the EVMOS chain vote on the proposal using `MsgVote` and proposal passes
3. If ERC-20 contract is deployed on the EVM module, create a bank coin `Metadata` from the ERC20 details.

## Token Pair Conversion

Conversion of a registered `TokenPair` can be done via:

- Cosmos transaction (`ConvertCoin` and `ConvertERC20)`
- Ethereum transaction (i.e sending a `MsgEthereumTx` that leverages the EVM hook)

### 1. Registered Coin

::: tip
👉 **Context:** A `TokenPair` has been created through a `RegisterCoinProposal` governance proposal. The proposal created an `ERC20` contract ([ERC20Mintable by openzeppelin](https://github.com/OpenZeppelin/openzeppelin-contracts/tree/master/contracts/token/ERC20)) of the ERC20 token representation of the Coin from the `ModuleAccount`, assigning it as the `owner` of the contract and thus granting it the permission to call the `mint()` and `burnFrom()` methods of the ERC20.
:::

#### Invariants

- Only the `ModuleAccount` should have the Minter Role on the ERC20. Otherwise,
  the user could unilaterally mint an infinite supply of the ERC20 token and
  then convert them to the native Coin
- The user and the `ModuleAccount` (owner) should be the only ones that have the
  Burn Role for a Cosmos Coin
- There shouldn't exist any native Cosmos Coin ERC20 Contract (eg Evmos, Atom,
  Osmo ERC20 contracts) that is not owned by the governance
- Token/Coin supply is maintained at all times:
    - Total Coin supply = Coins + Escrowed Coins
    - Total Token supply = Escrowed Coins = Minted Tokens

#### 1.1 Coin to ERC20

1. User submits `ConvertCoin` Tx
2. Check if conversion is allowed for the pair, sender and recipient
    - global parameter is enabled
    - token pair is enabled
    - sender tokens are not vesting (checked in the bank module)
    - recipient address is not blocklisted
3. If Coin is a native Cosmos Coin and Token Owner is `ModuleAccount`
    1. Escrow Cosmos coin by sending them to the erc20 module account
    2. Call `mint()` ERC20 tokens from the `ModuleAccount` address and send minted tokens to recipient address
4. Check if token balance increased by amount

#### 1.2 ERC20 to Coin

1. User submits a `ConvertERC20` Tx
2. Check if conversion is allowed for the pair, sender and recipient (see [1.1 Coin to ERC20](#11-coin-to-erc20))
3. If token is a ERC20 and Token Owner is `ModuleAccount`
    1. Call `burnCoins()` on ERC20 to burn ERC20 tokens from the user balance
    2. Send Coins (previously escrowed, see [1.1 Coin to ERC20](#11-coin-to-erc20)) from module to the recipient address.
4. Check if
   - Coin balance increased by amount
   - Token balance decreased by amount

### 2. Registered ERC20

::: tip
👉 **Context:** A `TokenPair` has been created through a `RegisterERC20Proposal` governance proposal. The `ModuleAccount` is not the owner of the contract, so it can't mint new tokens or burn on behalf of the user. The mechanism described below follows the same model as the ICS20 standard, by using escrow & mint / burn & unescrow logic.
:::

#### Invariants

- ERC20 Token supply on the EVM runtime is maintained at all times:
    - Escrowed ERC20 + Minted Cosmos Coin representation of ERC20 =  Burned Cosmos Coin representation of ERC20 + Unescrowed ERC20
        - Convert 10 ERC20 → Coin, the total supply increases by 10. Mint on Cosmos side, no changes on EVM
        - Convert 10 Coin → ERC20, the total supply decreases by 10. Burn on Cosmos side , no changes of supply on EVM
    - Total ERC20 token supply = Non Escrowed Tokens + Escrowed Tokens (on Module account address)
    - Total Coin supply for the native ERC20 = Escrowed ERC20 Tokens on module account  (i.e balance) = Minted Coins

#### 2.1 ERC20 to Coin

1. User submits a `ConvertERC20` Tx
2. Check if conversion is allowed for the pair, sender and recipient (See [1.1 Coin to ERC20](#11-coin-to-erc20))
3. If token is a ERC20 and Token Owner is **not** `ModuleAccount`
    1. Escrow ERC20 token by sending them to the erc20 module account
    2. Mint Cosmos coins of the corresponding token pair denomination and send coins to the recipient address
4. Check if
   - Coin balance increased by amount
   - Token balance decreased by amount
5. Fail if unexpected `Approval` event found in logs to prevent malicious contract behaviour

#### 2.2 Coin to ERC20

1. User submits `ConvertCoin` Tx
2. Check if conversion is allowed for the pair, sender and recipient
3. If coin is a native Cosmos coin and Token Owner is **not** `ModuleAccount`
    1. Escrow Cosmos Coins by sending them to the erc20 module account
    2. Unlock escrowed ERC20 from the module address by sending it to the recipient
    3. Burn escrowed Cosmos coins
4. Check if token balance increased by amount
5. Fail if unexpected `Approval` event found in logs to prevent malicious contract behaviour
