<!--
order: 4
-->

# Query Balances

Learn how to query balances of IBC Cosmos Coins and ERC-20s on Evmos. {synopsis}

This guide will cover the following query methods:

- [`evmosd` & Tendermint RPC](#evmosd--tendermint-rpc)
- [JSON-RPC](#json-rpc)
- [gRPC](#grpc)

:::warning
**Note**: In this document, the command line is used to interact with endpoints. For dApp developers, using libraries such as [cosmjs](https://github.com/cosmos/cosmjs) and [evmosjs](../libraries/evmosjs.md) is recommended instead.
:::

## `evmosd` & Tendermint RPC

Upon [installation](../../validators/quickstart/installation.md) and [configuration](../../validators/quickstart/binary.md) of the Evmos Daemon, developers can query account balances using `evmosd` with the following CLI command:

```bash
$ evmosd query bank balances $EVMOSADDRESS --count-total=$COUNTTOTAL --height=$HEIGHT --output=$OUTPUT --node=$NODE
balances:
- amount: "1000000000000000000"
  denom: aevmos
- amount: "100000"
  denom: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
pagination:
  next_key: null
  total: "0"
```

where:

- `$EVMOSADDRESS` is the Evmos address with balances of interest (eg. `evmos1...`).
- (optional) `$COUNTTOTAL` counts the total number of records in all balances to query for.
- (optional) `$HEIGHT` is the specific height to query state at (can error if node is pruning state).
- (optional) `$OUTPUT` is the output format (eg. `text`).
- (optional if running local node) `$NODE` is the Tendermint RPC node information is requested from (eg. `https://tendermint.bd.evmos.org:26657`).

Details of non-native currencies (ie. not `aevmos`) can be queried with the following CLI command:

```bash
$ evmosd query erc20 token-pair $DENOM --node=$NODE --height=$HEIGHT --output=$OUTPUT
token_pair:
  contract_owner: OWNER_MODULE
  denom: ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518
  enabled: true
  erc20_address: 0xFA3C22C069B9556A4B2f7EcE1Ee3B467909f4864
```

where `$DENOM` is the denomination of the coin (eg. `ibc/ED07A3391A1...`).

## JSON-RPC

Developers can query account balances of `aevmos` using the [`eth_getBalance`](../json-rpc/endpoints.md#ethgetbalance) JSON-RPC method in conjunction with [`curl`](https://curl.se/):

```bash
# Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":[`$ETHADDRESS`, `$BLOCK`],"id":1}' -H "Content-Type: application/json" $NODE

# Result
{"jsonrpc":"2.0","id":1,"result":"0x36354d5575577c8000"}
```

where:

- `$ETHADDRESS` is the Etherum hex-address the balance is to be queried from.
    Note that Evmos addresses (those beginning with `evmos1...`) can be converte.d to Ethereum addresses using libraries such as [evmosjs](../libraries/evmosjs.md).
- `$BLOCK` is the block number or block hash (eg. `"0x0"`).
    The reasoning for this parameter is due to [EIP-1898](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md).
- (optional if running local node) `$NODE` is the JSON-RPC node information is requested from (eg. `https://eth.bd.evmos.org:8545`).

Developers can also query account balances of `x/erc20`-module registered coins using the [`eth_call`](../json-rpc/endpoints.md#ethcall) JSON-RPC method in conjunction with [`curl`](https://curl.se/):

```bash
# Request
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_call","params":[{"from":`SENDERCONTRACTADDRESS`, "to":`ERCCONTRACTADDRESS`, "data":`$DATA`}, `$BLOCK`],"id":1}'  -H "Content-Type: application/json" $NODE

# Result
{"jsonrpc":"2.0","id":1,"result":"0x"}
```

where:

- `$SENDERCONTRACTADDRESS` is the Ethereum hex-address this smart contract call is sent from.
- `$ERCCONTRACTADDRESS` is the Ethereum hex-address of the ERC-20 contract corresponding to the coin denomination being queried.
- `$DATA` is the hash of the [`balanceof`](https://docs.openzeppelin.com/contracts/2.x/api/token/erc20#ERC20) method signature and encoded parameters.
    `balanceOf` is a required method in every ERC-20 contract, and the encoded parameter is the address which is having its balance queried. For additional information, see the [Ethereum Contract ABI](https://docs.soliditylang.org/en/v0.8.13/abi-spec.html).
- `$BLOCK` is the block number or block hash (eg. `"0x0"`).
    The reasoning for this parameter is due to [EIP-1898](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1898.md).
- (optional if running local node) `$NODE` is the JSON-RPC node information is requested from (eg. `https://eth.bd.evmos.org:8545`).

## gRPC

Developers can use [`grpcurl`](https://github.com/fullstorydev/grpcurl) with the `AllBalances` endpoint to query account balance by address for all denominations:

```bash
# Request
grpcurl $OUTPUT -d '{"address":`$EVMOSADDRESS`}' $NODE cosmos.bank.v1beta1.Query/AllBalances

# Result
{
  "balances": [
    {
      "denom": "stake",
      "amount": "1000000000"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

where:

- `$EVMOSADDRESS` is the Evmos address with balances of interest (eg. `"evmos1..."`).
- `$NODE` is the Cosmos gRPC node information is requested from (eg. `https://grpc.bd.evmos.org:9090`).
- (optional) `$OUTPUT` is the output format (eg. `plaintext`).

State can also be queried using gRPC within a Go program. The idea is to create a gRPC connection, then use the [Protobuf](https://developers.google.com/protocol-buffers)-generated client code to query the gRPC server.

```go
import (
    "context"
    "fmt"

  "google.golang.org/grpc"

    sdk "github.com/cosmos/cosmos-sdk/types"
  "github.com/cosmos/cosmos-sdk/types/tx"
)

func queryState() error {
    myAddress, err := GetEvmosAddressFromBech32("evmos1...") // evmos address with balances of interest.
    if err != nil {
        return err
    }

    // Create a connection to the gRPC server.
    grpcConn := grpc.Dial(
        "https://grpc.bd.evmos.org:9090", // your gRPC server address.
        grpc.WithInsecure(), // the SDK doesn't support any transport security mechanism.
    )
    defer grpcConn.Close()

    // This creates a gRPC client to query the x/bank service.
    bankClient := banktypes.NewQueryClient(grpcConn)
    bankRes, err := bankClient.AllBalances(
        context.Background(),
        &banktypes.QueryAllBalancesRequest{Address: myAddress},
    )
    if err != nil {
        return err
    }

    fmt.Println(bankRes.GetBalances()) // prints the account balances.

    return nil
}

// evmosjs address converter.
func GetEvmosAddressFromBech32(address string) (string, error) {...}
```

:::tip
**Note**: The following tools will be useful when using gRPC:

- [Evmos Swagger API](https://api.evmos.dev/): a comprehensive description of all gRPC endpoints
- [Cosmos SDK Go API](https://pkg.go.dev/github.com/cosmos/cosmos-sdk) & [Evmos Go API](https://pkg.go.dev/github.com/tharsis/evmos): packages to implement queries in Go scripts

:::
