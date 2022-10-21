# Guide

acred testnet init-files --keyring-backend=test --chain-id="bamboo_9000-1" --v=4 --output-dir ./testnet --starting-ip-address 192.168.10.2

## Get server addresses

ssh root@168.100.9.100
ssh root@168.100.8.60
ssh root@168.100.8.9
ssh root@168.100.10.133

## Copy testnet home files

```
rm -rf testnet.zip
rm -rf testnet/
rm -rf acrehome/
```

scp testnet.zip root@168.100.9.100:~/
scp testnet.zip root@168.100.8.60:~/
scp testnet.zip root@168.100.8.9:~/
scp testnet.zip root@168.100.10.133:~/

## Install go on server

VERSION="1.18.1"
ARCH="amd64"
curl -O -L "https://golang.org/dl/go${VERSION}.linux-${ARCH}.tar.gz"
tar -xf "go${VERSION}.linux-${ARCH}.tar.gz"
sudo chown -R root:root ./go
sudo mv -v go /usr/local

nano ~/.bashrc

```
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
```

source ~/.bashrc
go version

## Install unzip/gcc

```
apt install unzip
apt install gcc
```

unzip testnet.zip

## Setup home folder for daemon

```
mv testnet/node0/acred/ acrehome/
mv testnet/node1/acred/ acrehome/
mv testnet/node2/acred/ acrehome/
mv testnet/node3/acred/ acrehome/
```

## Install daemon on server

git clone https://github.com/ArableProtocol/acrechain.git
cd acrechain/
git checkout testnet_bamboo
go install ./cmd/acred/

## Setup systemctl

nano /etc/systemd/system/acred.service

```
[Unit]
Description=Acred Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root
ExecStart=/root/go/bin/acred start --home=/root/acrehome
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

systemctl start acred
systemctl stop acred
journalctl -u acred.service

acred tx slashing unjail --from=node0 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y
acred tx slashing unjail --from=node1 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y
acred tx slashing unjail --from=node2 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y
acred tx slashing unjail --from=node3 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

acred keys add node0 --keyring-backend=test --home=/root/acrehome --recover
about decrease option engine switch often assume raw lonely drink phone hard veteran fantasy lazy economy hat range law antique orchard submit drama winner

acred keys add node1 --keyring-backend=test --home=/root/acrehome --recover
diagram glide install lounge damage mammal load cheap concert lizard pulse garlic web half tower wrap human trade artwork final layer purse sibling music

acred tx bank send node1 acre1n2cn0y5m38pvtaru5slf6u5psmgnmu6fk6a7ld 100000000000000000000aacre --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y
acred tx bank send node1 acre1ljvjw0d6jce83nclnfn3qwla4najyty0n90gl9 100000000000000000000aacre --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

acred tx gov submit-proposal param-change proposal.json --from=node0 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

```
{
  "title": "Mint Param Change",
  "description": "Update mint denom",
  "changes": [
    {
      "subspace": "mint",
      "key": "MintDenom",
      "value": "aacre"
    }
  ],
  "deposit": "100000000000000000000aacre"
}
```

blocks_per_year: "6311520"
goal_bonded: "0.670000000000000000"
inflation_max: "0.200000000000000000"
inflation_min: "0.070000000000000000"
inflation_rate_change: "0.130000000000000000"
mint_denom: aacre

acred tx gov vote 1 Yes --from=node0 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

acred keys add node1 --keyring-backend=test --home=/root/acrehome --recover
acred tx gov vote 1 Yes --from=node1 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

acred keys add node2 --keyring-backend=test --home=/root/acrehome --recover
acred tx gov vote 1 Yes --from=node2 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y

acred keys add node3 --keyring-backend=test --home=/root/acrehome --recover
acred tx gov vote 1 Yes --from=node3 --keyring-backend=test --chain-id="bamboo_9000-1" --home=/root/acrehome --broadcast-mode=block -y
