import base64
import configparser
import json
import os
import socket
import subprocess
import sys
import time
from collections import defaultdict
from pathlib import Path

import bech32
import requests
from dateutil.parser import isoparse
from dotenv import load_dotenv
from eth_account import Account
from hexbytes import HexBytes
from pystarport import ports
from pystarport.cluster import SUPERVISOR_CONFIG_FILE
from web3 import Web3
from web3._utils.transactions import fill_nonce, fill_transaction_defaults
from web3.exceptions import TimeExhausted

from .http_rpc import comet_status

load_dotenv(Path(__file__).parent.parent.parent / "scripts/.env")
Account.enable_unaudited_hdwallet_features()
MAX_UINT256 = 2**256 - 1
ACCOUNTS = {
    "validator": Account.from_mnemonic(os.getenv("VALIDATOR1_MNEMONIC")),
    "community": Account.from_mnemonic(os.getenv("COMMUNITY_MNEMONIC")),
    "signer1": Account.from_mnemonic(os.getenv("SIGNER1_MNEMONIC")),
    "signer2": Account.from_mnemonic(os.getenv("SIGNER2_MNEMONIC")),
}
KEYS = {name: account.key for name, account in ACCOUNTS.items()}
ADDRS = {name: account.address for name, account in ACCOUNTS.items()}
EVMOS_ADDRESS_PREFIX = "evmos"
DEFAULT_DENOM = "aevmos"
WEVMOS_ADDRESS = Web3.toChecksumAddress("0xD4949664cD82660AaE99bEdc034a0deA8A0bd517")
TEST_CONTRACTS = {
    "TestERC20A": "TestERC20A.sol",
    "TestRevert": "TestRevert.sol",
    "Greeter": "Greeter.sol",
    "BurnGas": "BurnGas.sol",
    "TestChainID": "ChainID.sol",
    "Mars": "Mars.sol",
    "StateContract": "StateContract.sol",
    "ICS20FromContract": "ICS20FromContract.sol",
    "InterchainSender": "evmos/testutil/contracts/InterchainSender.sol",
    "InterchainSenderCaller": "evmos/testutil/contracts/InterchainSenderCaller.sol",
    "ICS20I": "evmos/ics20/ICS20I.sol",
    "DistributionI": "evmos/distribution/DistributionI.sol",
    "DistributionCaller": "evmos/testutil/contracts/DistributionCaller.sol",
    "StakingI": "evmos/staking/StakingI.sol",
    "StakingCaller": "evmos/staking/testdata/StakingCaller.sol",
    "IStrideOutpost": "evmos/outposts/stride/IStrideOutpost.sol",
    "IOsmosisOutpost": "evmos/outposts/osmosis/IOsmosisOutpost.sol",
    "IERC20": "evmos/erc20/IERC20.sol",
}

OSMOSIS_POOLS = {
    "Evmos_Osmo": Path(__file__).parent / "osmosis/evmosOsmosisPool.json",
}

# If need to update these binaries
# you can use the compile-cosmwasm-contracts.sh
# script located in the 'scripts' directory
WASM_BINARIES = {
    "CrosschainSwap": "crosschain_swaps.wasm",
    "Swaprouter": "swaprouter.wasm",
}

REGISTER_ERC20_PROP = {
    "messages": [
        {
            "@type": "/evmos.erc20.v1.MsgRegisterERC20",
            "authority": "evmos10d07y265gmmuvt4z0w9aw880jnsr700jcrztvm",
            "erc20addresses": ["ADDRESS_HERE"],
        }
    ],
    "metadata": "ipfs://CID",
    "deposit": "1aevmos",
    "title": "register erc20",
    "summary": "register erc20",
    "expedited": False,
}

PROPOSAL_STATUS_DEPOSIT_PERIOD = 1
PROPOSAL_STATUS_VOTING_PERIOD = 2
PROPOSAL_STATUS_PASSED = 3
PROPOSAL_STATUS_REJECTED = 4
PROPOSAL_STATUS_FAILED = 5


def wasm_binaries_path(filename):
    return Path(__file__).parent / "cosmwasm/artifacts/" / filename


def contract_path(name, filename):
    return (
        Path(__file__).parent
        / "hardhat/artifacts/contracts/"
        / filename
        / (name + ".json")
    )


WASM_CONTRACTS = {
    **{name: wasm_binaries_path(filename) for name, filename in WASM_BINARIES.items()},
}


CONTRACTS = {
    **{
        name: contract_path(name, filename) for name, filename in TEST_CONTRACTS.items()
    },
}


def wait_for_port(port, host="127.0.0.1", timeout=40.0):
    start_time = time.perf_counter()
    while True:
        try:
            with socket.create_connection((host, port), timeout=timeout):
                break
        except OSError as ex:
            time.sleep(0.1)
            if time.perf_counter() - start_time >= timeout:
                raise TimeoutError(
                    "Waited too long for the port {} on host {} to start accepting "
                    "connections.".format(port, host)
                ) from ex


def w3_wait_for_new_blocks(w3, n, sleep=0.5):
    begin_height = w3.eth.block_number
    while True:
        time.sleep(sleep)
        cur_height = w3.eth.block_number
        if cur_height - begin_height >= n:
            break


def wait_for_new_blocks(cli, n, sleep=0.5):
    """
    Helper function to wait for new blocks on a cosmos chain.
    If the chain has sdk < 0.50, the sync_info field will be 'SyncInfo'.
    With cosmos-sdk v0.50+, the sync_info field is 'sync_info'
    """
    sync_info_field = "sync_info"
    try:
        cur_height = begin_height = int(
            (cli.status())[sync_info_field]["latest_block_height"]
        )
    except KeyError:
        sync_info_field = "SyncInfo"
        cur_height = begin_height = int(
            (cli.status())[sync_info_field]["latest_block_height"]
        )

    while cur_height - begin_height < n:
        time.sleep(sleep)
        cur_height = int((cli.status())[sync_info_field]["latest_block_height"])

    return cur_height


def wait_for_block(cli, height, timeout=240):
    for _ in range(timeout * 2):
        current_height = get_current_height(cli)
        if current_height >= height:
            break
        print("current block height", current_height)
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


def http_wait_for_block(port, height, timeout=240):
    for _ in range(timeout * 2):
        status = comet_status(port)
        if status is None:
            time.sleep(0.5)
            continue
        current_height = int(status["sync_info"]["latest_block_height"])
        if current_height >= height:
            break
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


def get_current_height(cli):
    try:
        status = cli.status()
    except AssertionError as e:
        print(f"get sync status failed: {e}", file=sys.stderr)
    else:
        current_height = int(status["sync_info"]["latest_block_height"])
    return current_height


def w3_wait_for_block(w3, height, timeout=240):
    for _ in range(timeout * 2):
        try:
            current_height = w3.eth.block_number
        except Exception as e:
            print(f"get json-rpc block number failed: {e}", file=sys.stderr)
        else:
            if current_height >= height:
                break
            print("current block height", current_height)
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


def wait_for_block_time(cli, t):
    print("wait for block time", t)
    while True:
        now = isoparse((cli.status())["sync_info"]["latest_block_time"])
        print("block time now: ", now)
        if now >= t:
            break
        time.sleep(0.5)


def wait_for_fn(name, fn, *, timeout=240, interval=1):
    for _ in range(int(timeout / interval)):
        result = fn()
        print("check", name, result)
        if result:
            break
        time.sleep(interval)
    else:
        raise TimeoutError(f"wait for {name} timeout")


def approve_proposal(n, proposal_id, **kwargs):
    """
    helper function to vote 'yes' on the provided proposal id
    and wait it to pass
    """
    cli = n.cosmos_cli()

    # make the deposit (1 aevmos)
    rsp = cli.gov_deposit("signer2", proposal_id, 1)
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(cli, 1)

    for i in range(len(n.config["validators"])):
        rsp = n.cosmos_cli(i).gov_vote("validator", proposal_id, "yes", **kwargs)
        assert rsp["code"] == 0, rsp["raw_log"]
        wait_for_new_blocks(cli, 1)
    wait_for_new_blocks(cli, 2)
    assert (
        int(cli.query_tally(proposal_id)["yes_count"]) == cli.staking_pool()
    ), "all validators should have voted yes"
    print("wait for proposal to be activated")
    proposal = cli.query_proposal(proposal_id)
    wait_for_block_time(cli, isoparse(proposal["voting_end_time"]))
    proposal = cli.query_proposal(proposal_id)
    if isinstance(proposal["status"], int):
        assert int(proposal["status"]) == int(PROPOSAL_STATUS_PASSED), proposal
        return
    assert proposal["status"] == "PROPOSAL_STATUS_PASSED", proposal


def get_precompile_contract(w3, name):
    jsonfile = CONTRACTS[name]
    info = json.loads(jsonfile.read_text())
    if name == "StakingI":
        addr = "0x0000000000000000000000000000000000000800"
    elif name == "DistributionI":
        addr = "0x0000000000000000000000000000000000000801"
    elif name == "ICS20I":
        addr = "0x0000000000000000000000000000000000000802"
    elif name == "IStrideOutpost":
        addr = "0x0000000000000000000000000000000000000900"
    elif name == "IOsmosisOutpost":
        addr = "0x0000000000000000000000000000000000000901"
    else:
        raise ValueError(f"invalid precompile contract name: {name}")
    return w3.eth.contract(addr, abi=info["abi"])


def build_deploy_contract_tx(w3, info, args=(), key=KEYS["validator"]):
    """
    builds a tx to deploy contract without signature and returns it
    """
    acct = Account.from_key(key)
    contract = w3.eth.contract(abi=info["abi"], bytecode=info["bytecode"])
    return contract.constructor(*args).build_transaction({"from": acct.address})


def deploy_contract(w3, jsonfile, args=(), key=KEYS["validator"]):
    """
    deploy contract and return the deployed contract instance
    """
    info = json.loads(jsonfile.read_text())
    tx = build_deploy_contract_tx(w3, info, args, key)
    txreceipt = send_transaction(w3, tx, key)
    assert txreceipt.status == 1
    address = txreceipt.contractAddress
    return w3.eth.contract(address=address, abi=info["abi"]), txreceipt


def wait_for_cosmos_tx_receipt(cli, tx_hash):
    print(f"waiting receipt for tx_hash: {tx_hash}...")
    wait_for_new_blocks(cli, 1)
    res = cli.tx_search_rpc(f"tx.hash='{tx_hash}'")
    if len(res) == 0:
        return wait_for_cosmos_tx_receipt(cli, tx_hash)
    return res[0]


def wait_for_ack(cli, chain):
    """
    Helper function to wait for acknowledgment
    of an IBC transfer
    """
    print(f"{chain} waiting ack...")
    block_results = cli.block_results_rpc()
    txs_res = block_results["txs_results"]
    if txs_res is None:
        wait_for_new_blocks(cli, 1)
        return wait_for_ack(cli, chain)

    return None


def register_host_zone(
    stride,
    proposer,
    connection_id,
    host_denom,
    bech32_prefix,
    ibc_denom,
    channel_id,
    unbonding_frequency,
):
    """
    Register a Host Zone in Stride Chain.
    This helper function submits the corresponding
    transaction and checks that the host zone
    was registered successfully
    """
    prev_registered_zones = len(stride.cosmos_cli().get_host_zones())

    rsp = stride.cosmos_cli().register_host_zone_msg(
        proposer,
        connection_id,
        host_denom,
        bech32_prefix,
        ibc_denom,
        channel_id,
        unbonding_frequency,
        0,
        gas=700000,
    )
    assert rsp["code"] == 0, rsp["raw_log"]
    txhash = rsp["txhash"]

    # check the tx receipt to confirm was successful
    wait_for_new_blocks(stride.cosmos_cli(), 2)
    receipt = stride.cosmos_cli().tx_search_rpc(f"tx.hash='{txhash}'")[0]
    assert receipt["tx_result"]["code"] == 0

    updated_registered_zones = stride.cosmos_cli().get_host_zones()
    assert len(updated_registered_zones) == prev_registered_zones + 1
    return updated_registered_zones


def fill_defaults(w3, tx):
    return fill_nonce(w3, fill_transaction_defaults(w3, tx))


def sign_transaction(w3, tx, key=KEYS["validator"]):
    "fill default fields and sign"
    acct = Account.from_key(key)
    tx["from"] = acct.address
    tx = fill_transaction_defaults(w3, tx)
    tx = fill_nonce(w3, tx)
    return acct.sign_transaction(tx)


def send_transaction(w3, tx, key=KEYS["validator"], i=0):
    if i > 3:
        raise TimeExhausted
    signed = sign_transaction(w3, tx, key)
    txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
    try:
        return w3.eth.wait_for_transaction_receipt(txhash, timeout=20)
    except TimeExhausted:
        return send_transaction(w3, tx, key, i + 1)


def send_successful_transaction(w3, i=0):
    if i > 3:
        raise TimeExhausted
    signed = sign_transaction(w3, {"to": ADDRS["community"], "value": 1000})
    txhash = w3.eth.send_raw_transaction(signed.rawTransaction)
    try:
        receipt = w3.eth.wait_for_transaction_receipt(txhash, timeout=20)
        assert receipt.status == 1
    except TimeExhausted:
        return send_successful_transaction(w3, i + 1)
    return txhash


def eth_to_bech32(addr, prefix=EVMOS_ADDRESS_PREFIX):
    bz = bech32.convertbits(HexBytes(addr), 8, 5)
    return bech32.bech32_encode(prefix, bz)


def decode_bech32(addr):
    _, bz = bech32.bech32_decode(addr)
    return HexBytes(bytes(bech32.convertbits(bz, 5, 8)))


def supervisorctl(inipath, *args):
    subprocess.run(
        (sys.executable, "-msupervisor.supervisorctl", "-c", inipath, *args),
        check=True,
    )


def parse_events(logs):
    return {
        ev["type"]: {attr["key"]: attr["value"] for attr in ev["attributes"]}
        for ev in logs[0]["events"]
    }


def parse_events_rpc(events):
    result = defaultdict(dict)
    for ev in events:
        for attr in ev["attributes"]:
            if attr["key"] is None:
                continue
            # after sdk v0.47, key and value are strings instead of byte arrays
            if isinstance(attr["key"], str):
                result[ev["type"]][attr["key"]] = attr["value"]
            else:
                key = base64.b64decode(attr["key"].encode()).decode()
                if attr["value"] is not None:
                    value = base64.b64decode(attr["value"].encode()).decode()
                else:
                    value = None
                result[ev["type"]][key] = value
    return result


def derive_new_account(n=1):
    # derive a new address
    account_path = f"m/44'/60'/0'/0/{n}"
    mnemonic = os.getenv("COMMUNITY_MNEMONIC")
    return Account.from_mnemonic(mnemonic, account_path=account_path)


def compare_fields(a, b, fields):
    for field in fields:
        assert a[field] == b[field], f"{field} field mismatch"


# get_fees_from_tx_result returns the fees by unpacking them
# from the events contained in the tx_result of a cosmos transaction.
def get_fees_from_tx_result(tx_result, denom=DEFAULT_DENOM):
    return int(
        get_event_attribute_value(
            tx_result["events"],
            "tx",
            "fee",
        ).split(
            denom
        )[0]
    )


def memiavl_config(tmp_path: Path, file_name):
    """
    Creates a new JSONnet file with memIAVL + versionDB configuration.
    It takes as base the provided JSONnet file
    """
    tests_dir = str(Path(__file__).parent)
    root_dir = os.path.join(tests_dir, "..", "..")
    jsonnet_content = f"""
local default = import '{tests_dir}/configs/{file_name}.jsonnet';

default {{
  dotenv: '{root_dir}/scripts/.env',
  'evmos_9002-1'+: {{
    cmd: 'evmosd-rocksdb',
    'app-config'+: {{
      'app-db-backend': 'rocksdb',
      memiavl: {{
        enable: true,
      }},
      versiondb: {{
        enable: true,
      }},
    }},
    config+: {{
       'db_backend': 'rocksdb',
    }},
  }},
}}
    """

    # Write the JSONnet content to the file
    file_path = tmp_path / "configs" / f"{file_name}-memiavl.jsonnet"
    os.makedirs(file_path.parent, exist_ok=True)
    with open(file_path, "w") as f:
        f.write(jsonnet_content)

    return file_path


def get_event_attribute_value(events, _type, attribute):
    for event in events:
        if event["type"] == _type:
            attrs = event["attributes"]
            for attr in attrs:
                if attr["key"] == attribute:
                    return attr["value"]

    raise AttributeError(
        f"could not find attribute {attribute} in event logs: {events}"
    )


def update_node_cmd(path, cmd, i):
    ini_path = path / SUPERVISOR_CONFIG_FILE
    ini = configparser.RawConfigParser()
    ini.read(ini_path)
    for section in ini.sections():
        if section == f"program:evmos_9002-1-node{i}":
            ini[section].update(
                {
                    "command": f"{cmd} start --home %(here)s/node{i}",
                    "autorestart": "false",  # don't restart when stopped
                }
            )
    with ini_path.open("w") as fp:
        ini.write(fp)


def update_evmosd_and_setup_stride(modified_bin):
    def inner(path, base_port, config):  # pylint: disable=unused-argument
        update_evmos_bin(modified_bin)(path, base_port, config)
        setup_stride()(path, base_port, config)

    return inner


def update_evmos_bin(
    modified_bin, nodes=[0, 1]
):  # pylint: disable=dangerous-default-value
    """
    updates the evmos binary with a patched binary.
    Input parameters are the modified binary (modified_bin)
    and the nodes in which
    to apply the modified binary (nodes).
    Usually the setup comprise only 2 nodes (node0 & node1),
    so nodes should be an array containing only 0 and/or 1
    """

    def inner(path, base_port, config):  # pylint: disable=unused-argument
        chain_id = "evmos_9002-1"
        # by default, there're 2 nodes
        # need to update the bin in all these
        for i in nodes:
            update_node_cmd(path / chain_id, modified_bin, i)

    return inner


def setup_stride():
    def inner(path, base_port, config):  # pylint: disable=unused-argument
        chain_id = "stride-1"
        base_dir = Path(path / chain_id)
        os.environ["BASE_DIR"] = str(base_dir)
        os.environ["BASE_PORT"] = str(base_port)
        subprocess.run(["../../scripts/setup-stride.sh"], check=True)

    return inner


def erc20_balance(w3, erc20_contract_addr, addr):
    info = json.loads(CONTRACTS["IERC20"].read_text())
    contract = w3.eth.contract(erc20_contract_addr, abi=info["abi"])
    return contract.functions.balanceOf(addr).call()


def debug_trace_tx(evmos, tx_hash: str):
    url = f"http://127.0.0.1:{ports.evmrpc_port(evmos.base_port(0))}"
    params = {
        "method": "debug_traceTransaction",
        "params": [tx_hash, {"tracer": "callTracer"}],
        "id": 1,
        "jsonrpc": "2.0",
    }
    rsp = requests.post(url, json=params)
    assert rsp.status_code == 200
    return rsp.json()["result"]


def check_error(err: Exception, err_contains):
    if err_contains is not None:
        # stringify error in case it is an obj
        err_msg = json.dumps(err.args[0], separators=(",", ":"))
        assert err_contains in err_msg
    else:
        print(f"Unexpected {err=}, {type(err)=}")
        raise err


def erc20_transfer(w3, erc20_contract_addr, from_addr, to_addr, amount, key):
    info = json.loads(CONTRACTS["IERC20"].read_text())
    contract = w3.eth.contract(erc20_contract_addr, abi=info["abi"])
    tx = contract.functions.transfer(to_addr, amount).build_transaction(
        {"from": from_addr}
    )
    return send_transaction(w3, tx, key)
