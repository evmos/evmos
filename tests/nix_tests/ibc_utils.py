import json
import subprocess
from pathlib import Path
from typing import Any, Dict, List, NamedTuple

from pystarport import ports

from .network import CosmosChain, Hermes, setup_custom_evmos
from .utils import ADDRS, eth_to_bech32, update_evmos_bin, wait_for_port

# EVMOS_IBC_DENOM IBC denom of aevmos in crypto-org-chain
EVMOS_IBC_DENOM = "ibc/8EAC8061F4499F03D2D1419A3E73D346289AE9DB89CAB1486B72539572B1915E"
RATIO = 10**10
# IBC_CHAINS_META metadata of cosmos chains to setup these for IBC tests
IBC_CHAINS_META = {
    "chainmain": {
        "chain_name": "chainmain-1",
        "bin": "chain-maind",
        "denom": "basecro",
    },
    "stride": {
        "chain_name": "stride-1",
        "bin": "strided",
        "denom": "ustrd",
    },
    "osmosis": {
        "chain_name": "osmosis-1",
        "bin": "osmosisd",
        "denom": "uosmo",
    },
    "gaia": {
        "chain_name": "cosmoshub-1",
        "bin": "gaiad",
        "denom": "uatom",
    },
}


class IBCNetwork(NamedTuple):
    chains: Dict[str, Any]
    hermes: Hermes


def prepare_network(
    tmp_path: Path,
    file: str,
    other_chains_names: List[str],
    custom_scenario=None,
):
    file = f"configs/{file}.jsonnet"
    if custom_scenario is not None:
        # build the binary modified for a custom scenario
        # e.g. allow to register WEVMOS token
        # (removes a validation check in erc20 gov proposals)
        cmd = [
            "nix-build",
            "--no-out-link",
            str(Path(__file__).parent / f"configs/{custom_scenario}.nix"),
        ]
        print(*cmd)
        modified_bin = (
            Path(
                subprocess.check_output(cmd, universal_newlines=True, text=True).strip()
            )
            / "bin/evmosd"
        )
        print(f"patched bin: {modified_bin}")
        gen = setup_custom_evmos(
            tmp_path,
            26700,
            Path(__file__).parent / file,
            post_init=update_evmos_bin(modified_bin),
            chain_binary=modified_bin,
        )
    else:
        gen = setup_custom_evmos(tmp_path, 26700, Path(__file__).parent / file)

    evmos = next(gen)
    chains = {"evmos": evmos}
    # wait for grpc ready
    wait_for_port(ports.grpc_port(evmos.base_port(0)))  # evmos grpc

    # relayer
    hermes = Hermes(evmos.base_dir.parent / "relayer.toml")

    chains_to_connect = ["evmos_9000-1"]

    # set up the other chains to connect to evmos
    for chain in other_chains_names:
        meta = IBC_CHAINS_META[chain]
        other_chain_name = meta["chain_name"]
        chain_instance = CosmosChain(
            evmos.base_dir.parent / other_chain_name, meta["bin"]
        )
        # wait for grpc ready in other_chains
        wait_for_port(ports.grpc_port(chain_instance.base_port(0)))

        chains[chain] = chain_instance
        chains_to_connect.append(other_chain_name)
        # pystarport (used to start the setup), by default uses ethereum
        # hd-path to create the relayers keys on hermes.
        # If this is not needed (e.g. in Cosmos chains like Stride, Osmosis, etc.)
        # then overwrite the relayer key
        if "chainmain" not in other_chain_name:
            subprocess.run(
                [
                    "hermes",
                    "--config",
                    hermes.configpath,
                    "keys",
                    "add",
                    "--chain",
                    other_chain_name,
                    "--mnemonic-file",
                    evmos.base_dir.parent / "relayer.env",
                    "--overwrite",
                ],
                check=True,
            )

    # Nested loop to connect all chains with each other
    for i, chain_a in enumerate(chains_to_connect):
        for chain_b in chains_to_connect[i + 1 :]:
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
                    chain_a,
                    "--b-chain",
                    chain_b,
                    "--new-client-connection",
                    "--yes",
                ]
            )

    evmos.supervisorctl("start", "relayer-demo")
    wait_for_port(hermes.port)
    yield IBCNetwork(chains, hermes)


def assert_ready(ibc):
    # wait for hermes
    output = subprocess.getoutput(
        f"curl -s -X GET 'http://127.0.0.1:{ibc.hermes.port}/state' | jq"
    )
    assert json.loads(output)["status"] == "success"


def hermes_transfer(ibc, other_chain_name="chainmain-1", other_chain_denom="basecro"):
    assert_ready(ibc)
    # chainmain-1 -> evmos_9000-1
    my_ibc0 = other_chain_name
    my_ibc1 = "evmos_9000-1"
    my_channel = "channel-0"
    dst_addr = eth_to_bech32(ADDRS["signer2"])
    src_amount = 10
    src_denom = other_chain_denom
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
