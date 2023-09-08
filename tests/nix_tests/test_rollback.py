import configparser
import subprocess
from pathlib import Path

import pytest
from pystarport import ports
from pystarport.cluster import SUPERVISOR_CONFIG_FILE

from .network import setup_custom_evmos
from .utils import supervisorctl, wait_for_block, wait_for_port


def update_node_cmd(path, cmd, i):
    ini_path = path / SUPERVISOR_CONFIG_FILE
    ini = configparser.RawConfigParser()
    ini.read(ini_path)
    for section in ini.sections():
        if section == f"program:evmos_9000-1-node{i}":
            ini[section].update(
                {
                    "command": f"{cmd} start --home %(here)s/node{i}",
                    "autorestart": "false",  # don't restart when stopped
                }
            )
    with ini_path.open("w") as fp:
        ini.write(fp)


def post_init(broken_binary):
    def inner(path, base_port, config):
        chain_id = "evmos_9000-1"
        update_node_cmd(path / chain_id, broken_binary, 1)

    return inner


@pytest.fixture(scope="module")
def custom_evmos(tmp_path_factory):
    path = tmp_path_factory.mktemp("rollback")

    cmd = [
        "nix-build",
        "--no-out-link",
        Path(__file__).parent / "configs/broken-evmosd.nix",
    ]
    print(*cmd)
    broken_binary = Path(subprocess.check_output(cmd).strip().decode()) / "bin/evmosd"
    print(broken_binary)

    # init with genesis binary
    yield from setup_custom_evmos(
        path,
        26300,
        Path(__file__).parent / "configs/rollback-test.jsonnet",
        post_init=post_init(broken_binary),
        wait_port=False,
    )


def test_rollback(custom_evmos):
    """
    test using rollback command to fix app-hash mismatch situation.
    - the broken node will sync up to block 10 then crash.
    - use rollback command to rollback the db.
    - switch to correct binary should make the node syncing again.
    """
    target_port = ports.rpc_port(custom_evmos.base_port(1))
    wait_for_port(target_port)

    print("wait for node1 to sync the first 10 blocks")
    cli1 = custom_evmos.cosmos_cli(1)
    wait_for_block(cli1, 10)

    print("wait for a few more blocks on the healthy nodes")
    cli0 = custom_evmos.cosmos_cli(0)
    wait_for_block(cli0, 13)

    # (app hash mismatch happens after the 10th block, detected in the 11th block)
    print("check node1 get stuck at block 10")
    assert cli1.block_height() == 10

    print("stop node1")
    supervisorctl(custom_evmos.base_dir / "../tasks.ini", "stop", "evmos_9000-1-node1")

    print("do rollback on node1")
    cli1.rollback()

    print("switch to normal binary")
    update_node_cmd(custom_evmos.base_dir, "evmosd", 1)
    supervisorctl(custom_evmos.base_dir / "../tasks.ini", "update")
    wait_for_port(target_port)

    print("check node1 sync again")
    wait_for_block(cli1, 15)
