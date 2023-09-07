import json
import subprocess
from pathlib import Path
from typing import NamedTuple

from pystarport import ports

from .network import Chainmain, Evmos, Hermes, setup_custom_evmos
from .utils import ADDRS, eth_to_bech32, wait_for_port

# EVMOS_IBC_DENOM IBC denom of aevmos in crypto-org-chain
EVMOS_IBC_DENOM = "ibc/8EAC8061F4499F03D2D1419A3E73D346289AE9DB89CAB1486B72539572B1915E"
RATIO = 10**10


class IBCNetwork(NamedTuple):
    evmos: Evmos
    chainmain: Chainmain
    hermes: Hermes
    incentivized: bool


def prepare_network(tmp_path, file, incentivized=False):
    file = f"configs/{file}.jsonnet"
    gen = setup_custom_evmos(tmp_path, 26700, Path(__file__).parent / file)
    evmos = next(gen)
    chainmain = Chainmain(evmos.base_dir.parent / "chainmain-1")
    hermes = Hermes(evmos.base_dir.parent / "relayer.toml")
    # wait for grpc ready
    wait_for_port(ports.grpc_port(chainmain.base_port(0)))  # chainmain grpc
    wait_for_port(ports.grpc_port(evmos.base_port(0)))  # evmos grpc

    version = {"fee_version": "ics29-1", "app_version": "ics20-1"}
    incentivized_args = (
        [
            "--channel-version",
            json.dumps(version),
        ]
        if incentivized
        else []
    )

    subprocess.check_call(
        [
            "hermes",
            "--config",
            hermes.configpath,
            "create",
            "channel",
            "--a-port",
            "transfer",
            "--b-port",
            "transfer",
            "--a-chain",
            "evmos_9000-1",
            "--b-chain",
            "chainmain-1",
            "--new-client-connection",
            "--yes",
        ]
        + incentivized_args
    )

    if incentivized:
        # register fee payee
        src_chain = evmos.cosmos_cli()
        dst_chain = chainmain.cosmos_cli()
        rsp = dst_chain.register_counterparty_payee(
            "transfer",
            "channel-0",
            dst_chain.address("relayer"),
            src_chain.address("signer1"),
            from_="relayer",
            fees="100000000basecro",
        )
        assert rsp["code"] == 0, rsp["raw_log"]

    evmos.supervisorctl("start", "relayer-demo")
    wait_for_port(hermes.port)
    yield IBCNetwork(evmos, chainmain, hermes, incentivized)


def assert_ready(ibc):
    # wait for hermes
    output = subprocess.getoutput(
        f"curl -s -X GET 'http://127.0.0.1:{ibc.hermes.port}/state' | jq"
    )
    assert json.loads(output)["status"] == "success"


def hermes_transfer(ibc):
    assert_ready(ibc)
    # chainmain-1 -> evmos_9000-1
    my_ibc0 = "chainmain-1"
    my_ibc1 = "evmos_9000-1"
    my_channel = "channel-0"
    dst_addr = eth_to_bech32(ADDRS["signer2"])
    src_amount = 10
    src_denom = "basecro"
    # dstchainid srcchainid srcportid srchannelid
    cmd = (
        f"hermes --config {ibc.hermes.configpath} tx ft-transfer "
        f"--dst-chain {my_ibc1} --src-chain {my_ibc0} --src-port transfer "
        f"--src-channel {my_channel} --amount {src_amount} "
        f"--timeout-height-offset 1000 --number-msgs 1 "
        f"--denom {src_denom} --receiver {dst_addr} --key-name relayer"
    )
    subprocess.run(cmd, check=True, shell=True)
    return src_amount


def get_balance(chain, addr, denom):
    balance = chain.cosmos_cli().balance(addr, denom)
    print("balance", balance, addr, denom)
    return balance
