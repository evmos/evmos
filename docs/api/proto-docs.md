<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [evmos/epochs/v1/genesis.proto](#evmos/epochs/v1/genesis.proto)
    - [EpochInfo](#evmos.epochs.v1.EpochInfo)
    - [GenesisState](#evmos.epochs.v1.GenesisState)
  
- [evmos/epochs/v1/query.proto](#evmos/epochs/v1/query.proto)
    - [QueryCurrentEpochRequest](#evmos.epochs.v1.QueryCurrentEpochRequest)
    - [QueryCurrentEpochResponse](#evmos.epochs.v1.QueryCurrentEpochResponse)
    - [QueryEpochsInfoRequest](#evmos.epochs.v1.QueryEpochsInfoRequest)
    - [QueryEpochsInfoResponse](#evmos.epochs.v1.QueryEpochsInfoResponse)
  
    - [Query](#evmos.epochs.v1.Query)
  
- [evmos/erc20/v1/erc20.proto](#evmos/erc20/v1/erc20.proto)
    - [RegisterCoinProposal](#evmos.erc20.v1.RegisterCoinProposal)
    - [RegisterERC20Proposal](#evmos.erc20.v1.RegisterERC20Proposal)
    - [ToggleTokenRelayProposal](#evmos.erc20.v1.ToggleTokenRelayProposal)
    - [TokenPair](#evmos.erc20.v1.TokenPair)
    - [UpdateTokenPairERC20Proposal](#evmos.erc20.v1.UpdateTokenPairERC20Proposal)
  
    - [Owner](#evmos.erc20.v1.Owner)
  
- [evmos/erc20/v1/genesis.proto](#evmos/erc20/v1/genesis.proto)
    - [GenesisState](#evmos.erc20.v1.GenesisState)
    - [Params](#evmos.erc20.v1.Params)
  
- [evmos/erc20/v1/query.proto](#evmos/erc20/v1/query.proto)
    - [QueryParamsRequest](#evmos.erc20.v1.QueryParamsRequest)
    - [QueryParamsResponse](#evmos.erc20.v1.QueryParamsResponse)
    - [QueryTokenPairRequest](#evmos.erc20.v1.QueryTokenPairRequest)
    - [QueryTokenPairResponse](#evmos.erc20.v1.QueryTokenPairResponse)
    - [QueryTokenPairsRequest](#evmos.erc20.v1.QueryTokenPairsRequest)
    - [QueryTokenPairsResponse](#evmos.erc20.v1.QueryTokenPairsResponse)
  
    - [Query](#evmos.erc20.v1.Query)
  
- [evmos/erc20/v1/tx.proto](#evmos/erc20/v1/tx.proto)
    - [MsgConvertCoin](#evmos.erc20.v1.MsgConvertCoin)
    - [MsgConvertCoinResponse](#evmos.erc20.v1.MsgConvertCoinResponse)
    - [MsgConvertERC20](#evmos.erc20.v1.MsgConvertERC20)
    - [MsgConvertERC20Response](#evmos.erc20.v1.MsgConvertERC20Response)
  
    - [Msg](#evmos.erc20.v1.Msg)
  
- [Scalar Value Types](#scalar-value-types)



<a name="evmos/epochs/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/epochs/v1/genesis.proto



<a name="evmos.epochs.v1.EpochInfo"></a>

### EpochInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  |  |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `duration` | [google.protobuf.Duration](#google.protobuf.Duration) |  |  |
| `current_epoch` | [int64](#int64) |  |  |
| `current_epoch_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  |  |
| `epoch_counting_started` | [bool](#bool) |  |  |
| `current_epoch_start_height` | [int64](#int64) |  |  |






<a name="evmos.epochs.v1.GenesisState"></a>

### GenesisState
GenesisState defines the epochs module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#evmos.epochs.v1.EpochInfo) | repeated |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/epochs/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/epochs/v1/query.proto



<a name="evmos.epochs.v1.QueryCurrentEpochRequest"></a>

### QueryCurrentEpochRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `identifier` | [string](#string) |  |  |






<a name="evmos.epochs.v1.QueryCurrentEpochResponse"></a>

### QueryCurrentEpochResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `current_epoch` | [int64](#int64) |  |  |






<a name="evmos.epochs.v1.QueryEpochsInfoRequest"></a>

### QueryEpochsInfoRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  |  |






<a name="evmos.epochs.v1.QueryEpochsInfoResponse"></a>

### QueryEpochsInfoResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epochs` | [EpochInfo](#evmos.epochs.v1.EpochInfo) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.epochs.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `EpochInfos` | [QueryEpochsInfoRequest](#evmos.epochs.v1.QueryEpochsInfoRequest) | [QueryEpochsInfoResponse](#evmos.epochs.v1.QueryEpochsInfoResponse) | EpochInfos provide running epochInfos | GET|/evmos/epochs/v1/epochs|
| `CurrentEpoch` | [QueryCurrentEpochRequest](#evmos.epochs.v1.QueryCurrentEpochRequest) | [QueryCurrentEpochResponse](#evmos.epochs.v1.QueryCurrentEpochResponse) | CurrentEpoch provide current epoch of specified identifier | GET|/evmos/epochs/v1/current_epoch|

 <!-- end services -->



<a name="evmos/erc20/v1/erc20.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/erc20/v1/erc20.proto



<a name="evmos.erc20.v1.RegisterCoinProposal"></a>

### RegisterCoinProposal
RegisterCoinProposal is a gov Content type to register a token pair


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `metadata` | [cosmos.bank.v1beta1.Metadata](#cosmos.bank.v1beta1.Metadata) |  | token pair of Cosmos native denom and ERC20 token address |






<a name="evmos.erc20.v1.RegisterERC20Proposal"></a>

### RegisterERC20Proposal
RegisterCoinProposal is a gov Content type to register a token pair


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `erc20address` | [string](#string) |  | contract address of ERC20 token |






<a name="evmos.erc20.v1.ToggleTokenRelayProposal"></a>

### ToggleTokenRelayProposal
ToggleTokenRelayProposal is a gov Content type to toggle
the internal relaying of a token pair.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `token` | [string](#string) |  | token identifier can be either the hex contract address of the ERC20 or the Cosmos base denomination |






<a name="evmos.erc20.v1.TokenPair"></a>

### TokenPair
TokenPair defines an instance that records pairing consisting of a Cosmos
native Coin and an ERC20 token address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `erc20_address` | [string](#string) |  | address of ERC20 contract token |
| `denom` | [string](#string) |  | cosmos base denomination to be mapped to |
| `enabled` | [bool](#bool) |  | shows token mapping enable status |
| `contract_owner` | [Owner](#evmos.erc20.v1.Owner) |  | ERC20 owner address ENUM (0 invalid, 1 ModuleAccount, 2 external address) |






<a name="evmos.erc20.v1.UpdateTokenPairERC20Proposal"></a>

### UpdateTokenPairERC20Proposal
UpdateTokenPairERC20Proposal is a gov Content type to update a token pair's
ERC20 contract address.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `erc20_address` | [string](#string) |  | contract address of ERC20 token |
| `new_erc20_address` | [string](#string) |  | new address of ERC20 token contract |





 <!-- end messages -->


<a name="evmos.erc20.v1.Owner"></a>

### Owner
Owner enumerates the ownership of a ERC20 contract.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OWNER_UNSPECIFIED | 0 | OWNER_UNSPECIFIED defines an invalid/undefined owner. |
| OWNER_MODULE | 1 | OWNER_MODULE erc20 is owned by the erc20 module account. |
| OWNER_EXTERNAL | 2 | EXTERNAL erc20 is owned by an external account. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/erc20/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/erc20/v1/genesis.proto



<a name="evmos.erc20.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.erc20.v1.Params) |  | module parameters |
| `token_pairs` | [TokenPair](#evmos.erc20.v1.TokenPair) | repeated | registered token pairs |






<a name="evmos.erc20.v1.Params"></a>

### Params
Params defines the erc20 module params


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_erc20` | [bool](#bool) |  | parameter to enable the intrarelaying of Cosmos coins <--> ERC20 tokens. |
| `enable_evm_hook` | [bool](#bool) |  | parameter to enable the EVM hook to convert an ERC20 token to a Cosmos Coin by transferring the Tokens through a MsgEthereumTx to the ModuleAddress Ethereum address. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/erc20/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/erc20/v1/query.proto



<a name="evmos.erc20.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="evmos.erc20.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.erc20.v1.Params) |  |  |






<a name="evmos.erc20.v1.QueryTokenPairRequest"></a>

### QueryTokenPairRequest
QueryTokenPairRequest is the request type for the Query/TokenPair RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token` | [string](#string) |  | token identifier can be either the hex contract address of the ERC20 or the Cosmos base denomination |






<a name="evmos.erc20.v1.QueryTokenPairResponse"></a>

### QueryTokenPairResponse
QueryTokenPairResponse is the response type for the Query/TokenPair RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_pair` | [TokenPair](#evmos.erc20.v1.TokenPair) |  |  |






<a name="evmos.erc20.v1.QueryTokenPairsRequest"></a>

### QueryTokenPairsRequest
QueryTokenPairsRequest is the request type for the Query/TokenPairs RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="evmos.erc20.v1.QueryTokenPairsResponse"></a>

### QueryTokenPairsResponse
QueryTokenPairsResponse is the response type for the Query/TokenPairs RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `token_pairs` | [TokenPair](#evmos.erc20.v1.TokenPair) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.erc20.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `TokenPairs` | [QueryTokenPairsRequest](#evmos.erc20.v1.QueryTokenPairsRequest) | [QueryTokenPairsResponse](#evmos.erc20.v1.QueryTokenPairsResponse) | Retrieves registered token pairs | GET|/evmos/erc20/v1/token_pairs|
| `TokenPair` | [QueryTokenPairRequest](#evmos.erc20.v1.QueryTokenPairRequest) | [QueryTokenPairResponse](#evmos.erc20.v1.QueryTokenPairResponse) | Retrieves a registered token pair | GET|/evmos/erc20/v1/token_pairs/{token}|
| `Params` | [QueryParamsRequest](#evmos.erc20.v1.QueryParamsRequest) | [QueryParamsResponse](#evmos.erc20.v1.QueryParamsResponse) | Params retrieves the erc20 module params | GET|/evmos/erc20/v1/params|

 <!-- end services -->



<a name="evmos/erc20/v1/tx.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/erc20/v1/tx.proto



<a name="evmos.erc20.v1.MsgConvertCoin"></a>

### MsgConvertCoin
MsgConvertCoin defines a Msg to convert a Cosmos Coin to a ERC20 token


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coin` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) |  | Cosmos coin which denomination is registered on erc20 bridge. The coin amount defines the total ERC20 tokens to convert. |
| `receiver` | [string](#string) |  | recipient hex address to receive ERC20 token |
| `sender` | [string](#string) |  | cosmos bech32 address from the owner of the given ERC20 tokens |






<a name="evmos.erc20.v1.MsgConvertCoinResponse"></a>

### MsgConvertCoinResponse
MsgConvertCoinResponse returns no fields






<a name="evmos.erc20.v1.MsgConvertERC20"></a>

### MsgConvertERC20
MsgConvertERC20 defines a Msg to convert an ERC20 token to a Cosmos SDK coin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract_address` | [string](#string) |  | ERC20 token contract address registered on erc20 bridge |
| `amount` | [string](#string) |  | amount of ERC20 tokens to mint |
| `receiver` | [string](#string) |  | bech32 address to receive SDK coins. |
| `sender` | [string](#string) |  | sender hex address from the owner of the given ERC20 tokens |






<a name="evmos.erc20.v1.MsgConvertERC20Response"></a>

### MsgConvertERC20Response
MsgConvertERC20Response returns no fields





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.erc20.v1.Msg"></a>

### Msg
Msg defines the erc20 Msg service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `ConvertCoin` | [MsgConvertCoin](#evmos.erc20.v1.MsgConvertCoin) | [MsgConvertCoinResponse](#evmos.erc20.v1.MsgConvertCoinResponse) | ConvertCoin mints a ERC20 representation of the SDK Coin denom that is registered on the token mapping. | GET|/evmos/erc20/v1/tx/convert_coin|
| `ConvertERC20` | [MsgConvertERC20](#evmos.erc20.v1.MsgConvertERC20) | [MsgConvertERC20Response](#evmos.erc20.v1.MsgConvertERC20Response) | ConvertERC20 mints a Cosmos coin representation of the ERC20 token contract that is registered on the token mapping. | GET|/evmos/erc20/v1/tx/convert_erc20|

 <!-- end services -->



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

