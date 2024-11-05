import json
import os
import signal
import subprocess
from pathlib import Path

import tomlkit
import web3
from pystarport import ports
from web3.middleware import geth_poa_middleware

from .cosmoscli import CosmosCLI
from .utils import (
    EVMOS_6DEC_CHAIN_ID,
    evm6dec_config,
    http_wait_for_block,
    memiavl_config,
    supervisorctl,
    wait_for_port,
)

DEFAULT_CHAIN_BINARY = "evmosd"


class Evmos:
    def __init__(self, base_dir, chain_binary=DEFAULT_CHAIN_BINARY):
        self._w3 = None
        self.base_dir = base_dir
        self.config = json.loads((base_dir / "config.json").read_text())
        self.enable_auto_deployment = False
        self._use_websockets = False
        self.chain_binary = chain_binary

    def copy(self):
        return Evmos(self.base_dir)

    @property
    def w3_http_endpoint(self):  # pylint: disable=property-with-parameters
        port = ports.evmrpc_port(self.base_port(0))
        return f"http://localhost:{port}"

    @property
    def w3_ws_endpoint(self):
        port = ports.evmrpc_ws_port(self.base_port(0))
        return f"ws://localhost:{port}"

    @property
    def w3(self):
        if self._w3 is None:
            if self._use_websockets:
                self._w3 = web3.Web3(
                    web3.providers.WebsocketProvider(self.w3_ws_endpoint)
                )
            else:
                self._w3 = web3.Web3(web3.providers.HTTPProvider(self.w3_http_endpoint))
        return self._w3

    def base_port(self, i=0):  # pylint: disable=property-with-parameters
        return self.config["validators"][i]["base_port"]

    def node_rpc(self, i):
        return "tcp://127.0.0.1:%d" % ports.rpc_port(self.base_port(i))

    def node_api(self, i=0):
        return "http://127.0.0.1:%d" % ports.api_port(self.base_port(i))

    def use_websocket(self, use=True):
        self._w3 = None
        self._use_websockets = use

    def cosmos_cli(self, i=0):
        return CosmosCLI(
            self.base_dir / f"node{i}",
            self.node_rpc(i),
            self.node_api(i),
            self.chain_binary,
        )

    def node_home(self, i=0):
        return self.base_dir / f"node{i}"

    def supervisorctl(self, *args):
        return supervisorctl(self.base_dir / "../tasks.ini", *args)


# another Cosmos chain to be used on IBC transactions,
# e.g. Crypto.org, Stride, etc.
class CosmosChain:
    def __init__(self, base_dir, daemon_name):
        self.base_dir = base_dir
        self.config = json.loads((base_dir / "config.json").read_text())
        self.daemon_name = daemon_name

    def base_port(self, i=0):
        return self.config["validators"][i]["base_port"]

    def node_rpc(self, i=0):
        return "tcp://127.0.0.1:%d" % ports.rpc_port(self.base_port(i))

    def node_api(self, i=0):
        return "http://127.0.0.1:%d" % ports.api_port(self.base_port(i))

    def cosmos_cli(self, i=0):
        node_path = self.base_dir / f"node{i}"
        if node_path.exists() and node_path.is_dir():
            return CosmosCLI(
                node_path, self.node_rpc(i), self.node_api(i), self.daemon_name
            )
        # in case the provided directory does not exist
        # try with the other node. This applies for
        # stride that has a special setup because is a consumer chain
        node_path = self.base_dir / "node1"
        return CosmosCLI(node_path, self.node_rpc(), self.node_api(), self.daemon_name)


#  Hermes IBC relayer
class Hermes:
    def __init__(self, config: Path):
        self.configpath = config
        self.config = tomlkit.loads(config.read_text())
        self.port = 3000


class Geth:
    def __init__(self, w3):
        self.w3 = w3


def setup_evmos(path, base_port, long_timeout_commit=False):
    config = "configs/default.jsonnet"
    if long_timeout_commit is True:
        config = "configs/long_timeout_commit.jsonnet"
    cfg = Path(__file__).parent / config
    yield from setup_custom_evmos(path, base_port, cfg)


def setup_evmos_6dec(path, base_port, long_timeout_commit=False):
    """
    setup_evmos_6dec returns an Evmos chain with
    an EVM with 6 decimals and a "0.1" base fee.
    """
    config = evm6dec_config(
        path, "default" if long_timeout_commit is False else "long_timeout_commit"
    )
    cfg = Path(__file__).parent / config
    yield from setup_custom_evmos(path, base_port, cfg, chain_id=EVMOS_6DEC_CHAIN_ID)


# for memiavl need to create the data/snapshots dir
# for the nodes
def create_snapshots_dir(
    path, base_port, config, n_nodes=2
):  # pylint: disable=unused-argument
    for idx in range(n_nodes):
        data_snapshots_dir = path / "evmos_9002-1" / f"node{idx}" / "data" / "snapshots"
        os.makedirs(data_snapshots_dir, exist_ok=True)


def setup_evmos_rocksdb(path, base_port, long_timeout_commit=False):
    """
    setup_evmos_rocksdb returns an Evmos chain compiled with RocksDB
    and configured to use memIAVL + versionDB.
    """
    config = memiavl_config(
        path, "default" if long_timeout_commit is False else "long_timeout_commit"
    )
    cfg = Path(__file__).parent / config
    yield from setup_custom_evmos(
        path,
        base_port,
        cfg,
        post_init=create_snapshots_dir,
        chain_binary="evmosd-rocksdb",
    )


def setup_geth(path, base_port):
    with (path / "geth.log").open("w") as logfile:
        cmd = [
            "start-geth",
            path,
            "--http.port",
            str(base_port),
            "--port",
            str(base_port + 1),
            "--http.api",
            "eth,net,web3,debug",
        ]
        print(*cmd)
        proc = subprocess.Popen(  # pylint: disable=consider-using-with,subprocess-popen-preexec-fn
            cmd,
            preexec_fn=os.setsid,
            stdout=logfile,
            stderr=subprocess.STDOUT,
        )
        try:
            wait_for_port(base_port)
            w3 = web3.Web3(web3.providers.HTTPProvider(f"http://127.0.0.1:{base_port}"))
            w3.middleware_onion.inject(geth_poa_middleware, layer=0)
            yield Geth(w3)
        finally:
            os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
            # proc.terminate()
            proc.wait()


def setup_custom_evmos(
    path,
    base_port,
    config,
    post_init=None,
    chain_binary=None,
    wait_port=True,
    chain_id="evmos_9002-1",
):
    cmd = [
        "pystarport",
        "init",
        "--config",
        config,
        "--data",
        path,
        "--base_port",
        str(base_port),
        "--no_remove",
    ]
    print(*cmd)
    subprocess.run(cmd, check=True)
    if post_init is not None:
        post_init(path, base_port, config)
    proc = subprocess.Popen(  # pylint: disable=consider-using-with,subprocess-popen-preexec-fn
        ["pystarport", "start", "--data", path, "--quiet"],
        preexec_fn=os.setsid,
    )
    try:
        if wait_port:
            wait_for_port(ports.evmrpc_port(base_port))
            wait_for_port(ports.evmrpc_ws_port(base_port))
            # wait for blocks
            # cause with sdkv0.50 the port starts faster
            http_wait_for_block(ports.rpc_port(base_port), 2)
        yield Evmos(path / chain_id, chain_binary=chain_binary or DEFAULT_CHAIN_BINARY)
    finally:
        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
        proc.wait()


def build_patched_evmosd(patch_nix_file):
    """
    build the binary modified for a custom scenario
    e.g. allow to register WEVMOS token
    (removes a validation check in erc20 gov proposals)
    """
    cmd = [
        "nix-build",
        "--no-out-link",
        str(Path(__file__).parent / f"configs/{patch_nix_file}.nix"),
    ]
    print(*cmd)
    return (
        Path(subprocess.check_output(cmd, universal_newlines=True, text=True).strip())
        / "bin/evmosd"
    )
