
# Update the system
```
sudo apt-get update -y && sudo apt upgrade -y
```

# Install git, gcc and make
```
sudo apt-get install make build-essential gcc git jq chrony -y
```

# Install Go 
```
wget https://golang.org/dl/go1.18.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.18.5.linux-amd64.tar.gz
```

# Export environment variables
### Please don't forget to define your KEY_NAME and MONIKER_NAME for own at the rows of the end
```
cat <<EOF >> $HOME/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export CHAIN_ID=bamboo_9000-1
export SERVICE_NAME=acred 
export PROJECT_PATH=.acred 
export PROJECT_NAME=acred 
export TOKEN=uacre
export KEY_NAME=write_your_key_name
export MONIKER_NAME=write_your_moniker_name
EOF
```
```
source $HOME/.profile

go version
# Output should be: go version go1.18.5 linux/amd64
```

# Build
```
git clone https://github.com/ArableProtocol/acrechain && cd acrechain
git checkout testnet_bamboo
make install

```
```
acred version --long

# name: acre
# server_name: acred
# version: ""
# commit: 01482d6deddda2b0b4a399857857dc2a0dd38555
# build_tags: netgo,ledger
# go: go version go1.18.5 linux/amd64
```

# Copy binary - Setting up config
```
sudo cp $HOME/go/bin/$SERVICE_NAME /usr/local/bin/$SERVICE_NAME


$SERVICE_NAME config chain-id $CHAIN_ID
$SERVICE_NAME config keyring-backend test
$SERVICE_NAME init $MONIKER_NAME --chain-id $CHAIN_ID
```

# Create service
```
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null <<EOF  
[Unit]
Description=$PROJECT_NAME Node
After=network-online.target

[Service]
User=$USER
WorkingDirectory=$HOME/$PROJECT_PATH
ExecStart=$(which $SERVICE_NAME) start
Restart=always
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF
```

# Download genesis
```
wget -O $HOME/$PROJECT_PATH/config/genesis.json https://raw.githubusercontent.com/ArableProtocol/acrechain/testnet_bamboo/networks/bamboo/genesis.json

PEERS="44dd124ca34742245ad906f9f6ea377fae3af0cf@168.100.9.100:26656,6477921cdd4ba4503a1a2ff1f340c9d6a0e7b4a0@168.100.10.133:26656,9b53496211e75dbf33680b75e617830e874c8d93@168.100.8.9:26656,c55d79d6f76045ff7b68dc2bf6655348ebbfd795@168.100.8.60:26656"
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" $HOME/$PROJECT_PATH/config/config.toml
```

# Clear db
```
$SERVICE_NAME tendermint unsafe-reset-all --home $HOME/$PROJECT_PATH
```

# Start service
```
sudo systemctl daemon-reload && \
sudo systemctl enable $SERVICE_NAME && \
sudo systemctl restart $SERVICE_NAME && \
sudo journalctl -u $SERVICE_NAME -f -o cat
```

# Create a wallet and a validator after syncing
```
$SERVICE_NAME keys add $KEY_NAME 
$SERVICE_NAME tx staking create-validator \
  --amount="<AMOUNT>$TOKEN" \
  --pubkey=$($SERVICE_NAME tendermint show-validator) \
  --moniker=$MONIKER_NAME \
  --chain-id=$CHAIN_ID \
  --commission-rate=0.5 \
  --commission-max-rate=0.1 \
  --commission-max-change-rate=0.1 \
  --min-self-delegation=1 \
  --from=$KEY_NAME
```
