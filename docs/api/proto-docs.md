<!-- This file is auto-generated. Please do not modify it yourself. -->
# Protobuf Documentation
<a name="top"></a>

## Table of Contents

- [evmos/claims/v1/claims.proto](#evmos/claims/v1/claims.proto)
    - [Claim](#evmos.claims.v1.Claim)
    - [ClaimsRecord](#evmos.claims.v1.ClaimsRecord)
    - [ClaimsRecordAddress](#evmos.claims.v1.ClaimsRecordAddress)
  
    - [Action](#evmos.claims.v1.Action)
  
- [evmos/claims/v1/genesis.proto](#evmos/claims/v1/genesis.proto)
    - [GenesisState](#evmos.claims.v1.GenesisState)
    - [Params](#evmos.claims.v1.Params)
  
- [evmos/claims/v1/query.proto](#evmos/claims/v1/query.proto)
    - [QueryClaimsRecordRequest](#evmos.claims.v1.QueryClaimsRecordRequest)
    - [QueryClaimsRecordResponse](#evmos.claims.v1.QueryClaimsRecordResponse)
    - [QueryClaimsRecordsRequest](#evmos.claims.v1.QueryClaimsRecordsRequest)
    - [QueryClaimsRecordsResponse](#evmos.claims.v1.QueryClaimsRecordsResponse)
    - [QueryParamsRequest](#evmos.claims.v1.QueryParamsRequest)
    - [QueryParamsResponse](#evmos.claims.v1.QueryParamsResponse)
    - [QueryTotalUnclaimedRequest](#evmos.claims.v1.QueryTotalUnclaimedRequest)
    - [QueryTotalUnclaimedResponse](#evmos.claims.v1.QueryTotalUnclaimedResponse)
  
    - [Query](#evmos.claims.v1.Query)
  
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
  
- [evmos/incentives/v1/incentives.proto](#evmos/incentives/v1/incentives.proto)
    - [CancelIncentiveProposal](#evmos.incentives.v1.CancelIncentiveProposal)
    - [GasMeter](#evmos.incentives.v1.GasMeter)
    - [Incentive](#evmos.incentives.v1.Incentive)
    - [RegisterIncentiveProposal](#evmos.incentives.v1.RegisterIncentiveProposal)
  
- [evmos/incentives/v1/genesis.proto](#evmos/incentives/v1/genesis.proto)
    - [GenesisState](#evmos.incentives.v1.GenesisState)
    - [Params](#evmos.incentives.v1.Params)
  
- [evmos/incentives/v1/query.proto](#evmos/incentives/v1/query.proto)
    - [QueryAllocationMeterRequest](#evmos.incentives.v1.QueryAllocationMeterRequest)
    - [QueryAllocationMeterResponse](#evmos.incentives.v1.QueryAllocationMeterResponse)
    - [QueryAllocationMetersRequest](#evmos.incentives.v1.QueryAllocationMetersRequest)
    - [QueryAllocationMetersResponse](#evmos.incentives.v1.QueryAllocationMetersResponse)
    - [QueryGasMeterRequest](#evmos.incentives.v1.QueryGasMeterRequest)
    - [QueryGasMeterResponse](#evmos.incentives.v1.QueryGasMeterResponse)
    - [QueryGasMetersRequest](#evmos.incentives.v1.QueryGasMetersRequest)
    - [QueryGasMetersResponse](#evmos.incentives.v1.QueryGasMetersResponse)
    - [QueryIncentiveRequest](#evmos.incentives.v1.QueryIncentiveRequest)
    - [QueryIncentiveResponse](#evmos.incentives.v1.QueryIncentiveResponse)
    - [QueryIncentivesRequest](#evmos.incentives.v1.QueryIncentivesRequest)
    - [QueryIncentivesResponse](#evmos.incentives.v1.QueryIncentivesResponse)
    - [QueryParamsRequest](#evmos.incentives.v1.QueryParamsRequest)
    - [QueryParamsResponse](#evmos.incentives.v1.QueryParamsResponse)
  
    - [Query](#evmos.incentives.v1.Query)
  
- [evmos/inflation/v1/inflation.proto](#evmos/inflation/v1/inflation.proto)
    - [ExponentialCalculation](#evmos.inflation.v1.ExponentialCalculation)
    - [InflationDistribution](#evmos.inflation.v1.InflationDistribution)
  
- [evmos/inflation/v1/genesis.proto](#evmos/inflation/v1/genesis.proto)
    - [GenesisState](#evmos.inflation.v1.GenesisState)
    - [Params](#evmos.inflation.v1.Params)
  
- [evmos/inflation/v1/query.proto](#evmos/inflation/v1/query.proto)
    - [QueryEpochMintProvisionRequest](#evmos.inflation.v1.QueryEpochMintProvisionRequest)
    - [QueryEpochMintProvisionResponse](#evmos.inflation.v1.QueryEpochMintProvisionResponse)
    - [QueryParamsRequest](#evmos.inflation.v1.QueryParamsRequest)
    - [QueryParamsResponse](#evmos.inflation.v1.QueryParamsResponse)
    - [QueryPeriodRequest](#evmos.inflation.v1.QueryPeriodRequest)
    - [QueryPeriodResponse](#evmos.inflation.v1.QueryPeriodResponse)
  
    - [Query](#evmos.inflation.v1.Query)
  
- [Scalar Value Types](#scalar-value-types)



<a name="evmos/claims/v1/claims.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/claims/v1/claims.proto



<a name="evmos.claims.v1.Claim"></a>

### Claim
Claim marks defines the action, completed flag and the remaining claimable
amount for a given user. This is only used during client queries.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `action` | [Action](#evmos.claims.v1.Action) |  | action enum |
| `completed` | [bool](#bool) |  | true if the action has been completed |
| `claimable_amount` | [string](#string) |  | claimable token amount for the action. Zero if completed |






<a name="evmos.claims.v1.ClaimsRecord"></a>

### ClaimsRecord
ClaimsRecord defines the initial claimable airdrop amount and the list of
completed actions to claim the tokens.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `initial_claimable_amount` | [string](#string) |  | total initial claimable amount for the user |
| `actions_completed` | [bool](#bool) | repeated | slice of the available actions completed |






<a name="evmos.claims.v1.ClaimsRecordAddress"></a>

### ClaimsRecordAddress
ClaimsRecordAddress is the metadata of claims data per address


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  | bech32 or hex address of claim user |
| `initial_claimable_amount` | [string](#string) |  | total initial claimable amount for the user |
| `actions_completed` | [bool](#bool) | repeated | slice of the available actions completed |





 <!-- end messages -->


<a name="evmos.claims.v1.Action"></a>

### Action
Action defines the list of available actions to claim the airdrop tokens.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ACTION_UNSPECIFIED | 0 | UNSPECIFIED defines an invalid action. |
| ACTION_VOTE | 1 | VOTE defines a proposal vote. |
| ACTION_DELEGATE | 2 | DELEGATE defines an staking delegation. |
| ACTION_EVM | 3 | EVM defines an EVM transaction. |
| ACTION_IBC_TRANSFER | 4 | IBC Transfer defines a fungible token transfer transaction via IBC. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/claims/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/claims/v1/genesis.proto



<a name="evmos.claims.v1.GenesisState"></a>

### GenesisState
GenesisState defines the claims module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.claims.v1.Params) |  | params defines all the parameters of the module. |
| `claims_records` | [ClaimsRecordAddress](#evmos.claims.v1.ClaimsRecordAddress) | repeated | list of claim records with the corresponding airdrop recipient |






<a name="evmos.claims.v1.Params"></a>

### Params
Params defines the claims module's parameters.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_claims` | [bool](#bool) |  | enable claiming process |
| `airdrop_start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | timestamp of the airdrop start |
| `duration_until_decay` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration until decay of claimable tokens begin |
| `duration_of_decay` | [google.protobuf.Duration](#google.protobuf.Duration) |  | duration of the token claim decay period |
| `claims_denom` | [string](#string) |  | denom of claimable coin |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/claims/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/claims/v1/query.proto



<a name="evmos.claims.v1.QueryClaimsRecordRequest"></a>

### QueryClaimsRecordRequest
QueryClaimsRecordRequest is the request type for the Query/ClaimsRecord RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `address` | [string](#string) |  |  |






<a name="evmos.claims.v1.QueryClaimsRecordResponse"></a>

### QueryClaimsRecordResponse
QueryClaimsRecordResponse is the response type for the Query/ClaimsRecord RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `initial_claimable_amount` | [string](#string) |  | total initial claimable amount for the user |
| `claims` | [Claim](#evmos.claims.v1.Claim) | repeated |  |






<a name="evmos.claims.v1.QueryClaimsRecordsRequest"></a>

### QueryClaimsRecordsRequest
QueryClaimsRecordsRequest is the request type for the Query/ClaimsRecords RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="evmos.claims.v1.QueryClaimsRecordsResponse"></a>

### QueryClaimsRecordsResponse
QueryClaimsRecordsResponse is the response type for the Query/ClaimsRecords
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `claims` | [ClaimsRecordAddress](#evmos.claims.v1.ClaimsRecordAddress) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="evmos.claims.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="evmos.claims.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.claims.v1.Params) |  | params defines the parameters of the module. |






<a name="evmos.claims.v1.QueryTotalUnclaimedRequest"></a>

### QueryTotalUnclaimedRequest
QueryTotalUnclaimedRequest is the request type for the Query/TotalUnclaimed
RPC method.






<a name="evmos.claims.v1.QueryTotalUnclaimedResponse"></a>

### QueryTotalUnclaimedResponse
QueryTotalUnclaimedResponse is the response type for the Query/TotalUnclaimed
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `coins` | [cosmos.base.v1beta1.Coin](#cosmos.base.v1beta1.Coin) | repeated | coins define the unclaimed coins |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.claims.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `TotalUnclaimed` | [QueryTotalUnclaimedRequest](#evmos.claims.v1.QueryTotalUnclaimedRequest) | [QueryTotalUnclaimedResponse](#evmos.claims.v1.QueryTotalUnclaimedResponse) | TotalUnclaimed queries the total unclaimed tokens from the airdrop | GET|/evmos/claims/v1/total_unclaimed|
| `Params` | [QueryParamsRequest](#evmos.claims.v1.QueryParamsRequest) | [QueryParamsResponse](#evmos.claims.v1.QueryParamsResponse) | Params returns the claims module parameters | GET|/evmos/claims/v1/params|
| `ClaimsRecords` | [QueryClaimsRecordsRequest](#evmos.claims.v1.QueryClaimsRecordsRequest) | [QueryClaimsRecordsResponse](#evmos.claims.v1.QueryClaimsRecordsResponse) | ClaimsRecords returns all the claims record | GET|/evmos/claims/v1/claims_records|
| `ClaimsRecord` | [QueryClaimsRecordRequest](#evmos.claims.v1.QueryClaimsRecordRequest) | [QueryClaimsRecordResponse](#evmos.claims.v1.QueryClaimsRecordResponse) | ClaimsRecord returns the claims record for a given address | GET|/evmos/claims/v1/claims_record/{address}|

 <!-- end services -->



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
QueryTokenPairsRequest is the request type for the Query/TokenPairs RPC
method.


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



<a name="evmos/incentives/v1/incentives.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/incentives/v1/incentives.proto



<a name="evmos.incentives.v1.CancelIncentiveProposal"></a>

### CancelIncentiveProposal
CancelIncentiveProposal is a gov Content type to cancel an incentive


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `contract` | [string](#string) |  | contract address |






<a name="evmos.incentives.v1.GasMeter"></a>

### GasMeter
GasMeter tracks the cumulative gas spent per participant in one epoch


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | hex address of the incentivized contract |
| `participant` | [string](#string) |  | participant address that interacts with the incentive |
| `cumulative_gas` | [uint64](#uint64) |  | cumulative gas spent during the epoch |






<a name="evmos.incentives.v1.Incentive"></a>

### Incentive
Incentive defines an instance that organizes distribution conditions for a
given smart contract


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | contract address |
| `allocations` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | denoms and percentage of rewards to be allocated |
| `epochs` | [uint32](#uint32) |  | number of remaining epochs |
| `start_time` | [google.protobuf.Timestamp](#google.protobuf.Timestamp) |  | distribution start time |
| `total_gas` | [uint64](#uint64) |  | cumulative gas spent by all gasmeters of the incentive during the epoch |






<a name="evmos.incentives.v1.RegisterIncentiveProposal"></a>

### RegisterIncentiveProposal
RegisterIncentiveProposal is a gov Content type to register an incentive


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `title` | [string](#string) |  | title of the proposal |
| `description` | [string](#string) |  | proposal description |
| `contract` | [string](#string) |  | contract address |
| `allocations` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated | denoms and percentage of rewards to be allocated |
| `epochs` | [uint32](#uint32) |  | number of remaining epochs |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/incentives/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/incentives/v1/genesis.proto



<a name="evmos.incentives.v1.GenesisState"></a>

### GenesisState
GenesisState defines the module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.incentives.v1.Params) |  | module parameters |
| `incentives` | [Incentive](#evmos.incentives.v1.Incentive) | repeated | active incentives |
| `gas_meters` | [GasMeter](#evmos.incentives.v1.GasMeter) | repeated | active Gasmeters |






<a name="evmos.incentives.v1.Params"></a>

### Params
Params defines the incentives module params


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `enable_incentives` | [bool](#bool) |  | parameter to enable incentives |
| `allocation_limit` | [string](#string) |  | maximum percentage an incentive can allocate per denomination |
| `incentives_epoch_identifier` | [string](#string) |  | identifier for the epochs module hooks |
| `reward_scaler` | [string](#string) |  | scaling factor for capping rewards |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/incentives/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/incentives/v1/query.proto



<a name="evmos.incentives.v1.QueryAllocationMeterRequest"></a>

### QueryAllocationMeterRequest
QueryAllocationMeterRequest is the request type for the Query/AllocationMeter
RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `denom` | [string](#string) |  | denom is the coin denom to query an allocation meter for. |






<a name="evmos.incentives.v1.QueryAllocationMeterResponse"></a>

### QueryAllocationMeterResponse
QueryAllocationMeterResponse is the response type for the
Query/AllocationMeter RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allocation_meter` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) |  |  |






<a name="evmos.incentives.v1.QueryAllocationMetersRequest"></a>

### QueryAllocationMetersRequest
QueryAllocationMetersRequest is the request type for the
Query/AllocationMeters RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="evmos.incentives.v1.QueryAllocationMetersResponse"></a>

### QueryAllocationMetersResponse
QueryAllocationMetersResponse is the response type for the
Query/AllocationMeters RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `allocation_meters` | [cosmos.base.v1beta1.DecCoin](#cosmos.base.v1beta1.DecCoin) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="evmos.incentives.v1.QueryGasMeterRequest"></a>

### QueryGasMeterRequest
QueryGasMeterRequest is the request type for the Query/Incentive RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | contract identifier is the hex contract address of a contract |
| `participant` | [string](#string) |  | participant identifier is the hex address of a user |






<a name="evmos.incentives.v1.QueryGasMeterResponse"></a>

### QueryGasMeterResponse
QueryGasMeterResponse is the response type for the Query/Incentive RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas_meter` | [uint64](#uint64) |  |  |






<a name="evmos.incentives.v1.QueryGasMetersRequest"></a>

### QueryGasMetersRequest
QueryGasMetersRequest is the request type for the Query/Incentives RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | contract is the hex contract address of a incentivized smart contract |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="evmos.incentives.v1.QueryGasMetersResponse"></a>

### QueryGasMetersResponse
QueryGasMetersResponse is the response type for the Query/Incentives RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `gas_meters` | [GasMeter](#evmos.incentives.v1.GasMeter) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="evmos.incentives.v1.QueryIncentiveRequest"></a>

### QueryIncentiveRequest
QueryIncentiveRequest is the request type for the Query/Incentive RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `contract` | [string](#string) |  | contract identifier is the hex contract address of a contract |






<a name="evmos.incentives.v1.QueryIncentiveResponse"></a>

### QueryIncentiveResponse
QueryIncentiveResponse is the response type for the Query/Incentive RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incentive` | [Incentive](#evmos.incentives.v1.Incentive) |  |  |






<a name="evmos.incentives.v1.QueryIncentivesRequest"></a>

### QueryIncentivesRequest
QueryIncentivesRequest is the request type for the Query/Incentives RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `pagination` | [cosmos.base.query.v1beta1.PageRequest](#cosmos.base.query.v1beta1.PageRequest) |  | pagination defines an optional pagination for the request. |






<a name="evmos.incentives.v1.QueryIncentivesResponse"></a>

### QueryIncentivesResponse
QueryIncentivesResponse is the response type for the Query/Incentives RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `incentives` | [Incentive](#evmos.incentives.v1.Incentive) | repeated |  |
| `pagination` | [cosmos.base.query.v1beta1.PageResponse](#cosmos.base.query.v1beta1.PageResponse) |  | pagination defines the pagination in the response. |






<a name="evmos.incentives.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="evmos.incentives.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC
method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.incentives.v1.Params) |  |  |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.incentives.v1.Query"></a>

### Query
Query defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Incentives` | [QueryIncentivesRequest](#evmos.incentives.v1.QueryIncentivesRequest) | [QueryIncentivesResponse](#evmos.incentives.v1.QueryIncentivesResponse) | Incentives retrieves registered incentives | GET|/evmos/incentives/v1/incentives|
| `Incentive` | [QueryIncentiveRequest](#evmos.incentives.v1.QueryIncentiveRequest) | [QueryIncentiveResponse](#evmos.incentives.v1.QueryIncentiveResponse) | Incentive retrieves a registered incentive | GET|/evmos/incentives/v1/incentives/{contract}|
| `GasMeters` | [QueryGasMetersRequest](#evmos.incentives.v1.QueryGasMetersRequest) | [QueryGasMetersResponse](#evmos.incentives.v1.QueryGasMetersResponse) | GasMeters retrieves active gas meters for a given contract | GET|/evmos/incentives/v1/gas_meters/{contract}|
| `GasMeter` | [QueryGasMeterRequest](#evmos.incentives.v1.QueryGasMeterRequest) | [QueryGasMeterResponse](#evmos.incentives.v1.QueryGasMeterResponse) | GasMeter Retrieves a active gas meter | GET|/evmos/incentives/v1/gas_meters/{contract}/{participant}|
| `AllocationMeters` | [QueryAllocationMetersRequest](#evmos.incentives.v1.QueryAllocationMetersRequest) | [QueryAllocationMetersResponse](#evmos.incentives.v1.QueryAllocationMetersResponse) | AllocationMeters retrieves active allocation meters for a given denomination | GET|/evmos/incentives/v1/allocation_meters|
| `AllocationMeter` | [QueryAllocationMeterRequest](#evmos.incentives.v1.QueryAllocationMeterRequest) | [QueryAllocationMeterResponse](#evmos.incentives.v1.QueryAllocationMeterResponse) | AllocationMeter Retrieves a active gas meter | GET|/evmos/incentives/v1/allocation_meters/{denom}|
| `Params` | [QueryParamsRequest](#evmos.incentives.v1.QueryParamsRequest) | [QueryParamsResponse](#evmos.incentives.v1.QueryParamsResponse) | Params retrieves the incentives module params | GET|/evmos/incentives/v1/params|

 <!-- end services -->



<a name="evmos/inflation/v1/inflation.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/inflation/v1/inflation.proto



<a name="evmos.inflation.v1.ExponentialCalculation"></a>

### ExponentialCalculation
ExponentialCalculation holds factors to calculate exponential inflation on
each period. Calculation reference:
periodProvision = exponentialDecay       *  bondingRatio
f(x)            = (a * (1 - r) ^ x + c)  *  (2 - b) / 2


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `a` | [string](#string) |  | initial value |
| `r` | [string](#string) |  | reduction factor |
| `c` | [string](#string) |  | long term inflation |
| `b` | [string](#string) |  | bonding factor` |






<a name="evmos.inflation.v1.InflationDistribution"></a>

### InflationDistribution
InflationDistribution defines the distribution in which inflation is
allocated through minting on each epoch (staking, incentives, community). It
excludes the team vesting distribution, as this is minted once at genesis.
The initial InflationDistribution can be calculated from the Evmvos Token
Model like this:
mintDistribution1 = distribution1 / (1 - teamVestingDistribution)
0.5333333         = 40%           / (1 - 25%)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `staking_rewards` | [string](#string) |  | staking_rewards defines the proportion of the minted minted_denom that is to be allocated as staking rewards |
| `usage_incentives` | [string](#string) |  | usage_incentives defines the proportion of the minted minted_denom that is to be allocated to the incentives module address |
| `community_pool` | [string](#string) |  | community_pool defines the proportion of the minted minted_denom that is to be allocated to the community pool |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/inflation/v1/genesis.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/inflation/v1/genesis.proto



<a name="evmos.inflation.v1.GenesisState"></a>

### GenesisState
GenesisState defines the inflation module's genesis state.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.inflation.v1.Params) |  | params defines all the paramaters of the module. |
| `period` | [uint64](#uint64) |  | amount of past periods, based on the epochs per period param |
| `epoch_identifier` | [string](#string) |  | inflation epoch identifier |
| `epochs_per_period` | [int64](#int64) |  | number of epochs after which inflation is recalculated |






<a name="evmos.inflation.v1.Params"></a>

### Params
Params holds parameters for the inflation module.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `mint_denom` | [string](#string) |  | type of coin to mint |
| `exponential_calculation` | [ExponentialCalculation](#evmos.inflation.v1.ExponentialCalculation) |  | variables to calculate exponential inflation |
| `inflation_distribution` | [InflationDistribution](#evmos.inflation.v1.InflationDistribution) |  | inflation distribution of the minted denom |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->



<a name="evmos/inflation/v1/query.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## evmos/inflation/v1/query.proto



<a name="evmos.inflation.v1.QueryEpochMintProvisionRequest"></a>

### QueryEpochMintProvisionRequest
QueryEpochMintProvisionRequest is the request type for the
Query/EpochMintProvision RPC method.






<a name="evmos.inflation.v1.QueryEpochMintProvisionResponse"></a>

### QueryEpochMintProvisionResponse
QueryEpochMintProvisionResponse is the response type for the
Query/EpochMintProvision RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `epoch_mint_provision` | [bytes](#bytes) |  | epoch_mint_provision is the current minting per epoch provision value. |






<a name="evmos.inflation.v1.QueryParamsRequest"></a>

### QueryParamsRequest
QueryParamsRequest is the request type for the Query/Params RPC method.






<a name="evmos.inflation.v1.QueryParamsResponse"></a>

### QueryParamsResponse
QueryParamsResponse is the response type for the Query/Params RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `params` | [Params](#evmos.inflation.v1.Params) |  | params defines the parameters of the module. |






<a name="evmos.inflation.v1.QueryPeriodRequest"></a>

### QueryPeriodRequest
QueryPeriodRequest is the request type for the Query/Period RPC method.






<a name="evmos.inflation.v1.QueryPeriodResponse"></a>

### QueryPeriodResponse
QueryPeriodResponse is the response type for the Query/Period RPC method.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| `period` | [uint64](#uint64) |  | period is the current minting per epoch provision value. |





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->


<a name="evmos.inflation.v1.Query"></a>

### Query
Query provides defines the gRPC querier service.

| Method Name | Request Type | Response Type | Description | HTTP Verb | Endpoint |
| ----------- | ------------ | ------------- | ------------| ------- | -------- |
| `Period` | [QueryPeriodRequest](#evmos.inflation.v1.QueryPeriodRequest) | [QueryPeriodResponse](#evmos.inflation.v1.QueryPeriodResponse) | Period retrieves current period. | GET|/evmos/inflation/v1/period|
| `EpochMintProvision` | [QueryEpochMintProvisionRequest](#evmos.inflation.v1.QueryEpochMintProvisionRequest) | [QueryEpochMintProvisionResponse](#evmos.inflation.v1.QueryEpochMintProvisionResponse) | EpochMintProvision retrieves current minting epoch provision value. | GET|/evmos/inflation/v1/epoch_mint_provision|
| `Params` | [QueryParamsRequest](#evmos.inflation.v1.QueryParamsRequest) | [QueryParamsResponse](#evmos.inflation.v1.QueryParamsResponse) | Params retrieves the total set of minting parameters. | GET|/evmos/inflation/v1/params|

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

