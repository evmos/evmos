# Run Triton Validator

Overview
Ensure Go is installed:
```go version```

If Golang is not installed, install it using official tutorial: https://go.dev/doc/install. 

Also to build node requires make to be installed.
```sudo apt-get install build-essential```


Pull the repository of the point chain: https://github.com/pointnetwork/point-chain
Stay on the main branch and run: ```make install```


Check if evmosd command is available for you. If you see ```evmosd: command not found``` 
message then export path for this command: ```export PATH=$PATH:$(go env GOPATH)/bin```


Configure your validator key: 

```evmosd config keyring-backend file```

```evmosd config chain-id point_10721-1```


Input you 24 words passphrase for you validator key using this command:
```evmosd keys add validatorkey --keyring-backend file --algo eth_secp256k1 --recover```


Run the init script
Init you validator where <myvalidator> is your validator custom name
```evmosd init <myvalidator> --chain-id point_10721-1```


Copy ```genesis.json``` and ```config.toml``` ```files from this repository https://github.com/pointnetwork/point-chain-config/tree/main/testnet-xNet-Triton-1```  into ```~/.evmosd/config```
Validate it: ```evmosd validate-genesis```

Then run the node and wait for fully sync:
```evmosd start --json-rpc.enable=true --json-rpc.api "eth,txpool,personal,net,debug,web3"```

You can run this command to see status of your node:
```evmosd status```


You will get the "latest_block_height" of your node.
To see current block height of blockchain run:


```curl  http://xnet-neptune-1.point.space:8545 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'``` 


The result is in hexadecimal, just convert to decimal and see how far are you from fully sync.

In order to import the wallet in your metamask you will need the private key. You can get it with this command:
```evmosd keys unsafe-export-eth-key validatorkey --keyring-backend file```


Now let’s import the wallet in metamask. Go to the import account section, select type “Private key” and insert the private key you got from the command above.

Let’s add the point blockchain to metamask:

Network title: xPOINT Triton testnet
Rpc Url: ```http://xnet-neptune-1.point.space:8545/```
Chain id: ```10721```
SYMBOL: ```xPOINT```

Now that you have an account you need to get some point to stake and run your validator. (contact point team)

Once the node is fully synced, and you got some point to stake check your balance in the node you 
will see your balance in Metamask or you chan check your balance with this command:
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
If everything works ok you will get a txhash
You can check the status of the tx. 
```evmosd query tx <txhash>```

Transaction receipt may contain errors, so please check if there are any.
If the transaction was correct you should be part of the validators. Check your pubkey first:
```evmosd tendermint show-validator```


You will see a key there, you can identify your node among other validators using that key:
```evmosd query tendermint-validator-set```


There you will find more info like your VotingPower that should be bigger than 0.
Also you can check your VotingPower by running:


```evmosd status```


