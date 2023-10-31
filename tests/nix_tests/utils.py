import base64
import configparser
import json
import os
import socket
import subprocess
import sys
import tempfile
import time
from collections import defaultdict
from pathlib import Path

import bech32
from dateutil.parser import isoparse
from dotenv import load_dotenv
from eth_account import Account
from hexbytes import HexBytes
from pystarport.cluster import SUPERVISOR_CONFIG_FILE
from web3._utils.transactions import fill_nonce, fill_transaction_defaults
from web3.exceptions import TimeExhausted

load_dotenv(Path(__file__).parent.parent.parent / "scripts/.env")
Account.enable_unaudited_hdwallet_features()
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
TEST_CONTRACTS = {
    "TestERC20A": "TestERC20A.sol",
    "Greeter": "Greeter.sol",
    "BurnGas": "BurnGas.sol",
    "TestChainID": "ChainID.sol",
    "Mars": "Mars.sol",
    "StateContract": "StateContract.sol",
    "ICS20I": "evmos/ics20/ICS20I.sol",
    "DistributionI": "evmos/distribution/DistributionI.sol",
    "StakingI": "evmos/staking/StakingI.sol",
    "IStrideOutpost": "evmos/outposts/stride/IStrideOutpost.sol",
    "IERC20": "evmos/erc20/IERC20.sol",
}
WEVMOS_META = {
    "description": "The native staking and governance token of the Evmos chain",
    "denom_units": [
        {"denom": "aevmos", "exponent": 0, "aliases": ["aevmos"]},
        {"denom": "WEVMOS", "exponent": 18},
    ],
    "base": "aevmos",
    "display": "WEVMOS",
    "name": "Wrapped EVMOS",
    "symbol": "WEVMOS",
}


def contract_path(name, filename):
    return (
        Path(__file__).parent
        / "hardhat/artifacts/contracts/"
        / filename
        / (name + ".json")
    )


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
    cur_height = begin_height = int((cli.status())["SyncInfo"]["latest_block_height"])
    while cur_height - begin_height < n:
        time.sleep(sleep)
        cur_height = int((cli.status())["SyncInfo"]["latest_block_height"])
    return cur_height


def wait_for_block(cli, height, timeout=240):
    for _ in range(timeout * 2):
        try:
            status = cli.status()
        except AssertionError as e:
            print(f"get sync status failed: {e}", file=sys.stderr)
        else:
            current_height = int(status["SyncInfo"]["latest_block_height"])
            if current_height >= height:
                break
            print("current block height", current_height)
        time.sleep(0.5)
    else:
        raise TimeoutError(f"wait for block {height} timeout")


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
        now = isoparse((cli.status())["SyncInfo"]["latest_block_time"])
        print("block time now: ", now)
        if now >= t:
            break
        time.sleep(0.5)


def wait_for_fn(name, fn, *, timeout=240, interval=1):
    for i in range(int(timeout / interval)):
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

    for i in range(len(n.config["validators"])):
        rsp = n.cosmos_cli(i).gov_vote("validator", proposal_id, "yes", **kwargs)
        assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(cli, 1)
    assert (
        int(cli.query_tally(proposal_id)["yes_count"]) == cli.staking_pool()
    ), "all validators should have voted yes"
    print("wait for proposal to be activated")
    proposal = cli.query_proposal(proposal_id)
    wait_for_block_time(cli, isoparse(proposal["voting_end_time"]))
    proposal = cli.query_proposal(proposal_id)
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
    else:
        raise ValueError(f"invalid precompile contract name: {name}")
    return w3.eth.contract(addr, abi=info["abi"])


def deploy_contract(w3, jsonfile, args=(), key=KEYS["validator"]):
    """
    deploy contract and return the deployed contract instance
    """
    acct = Account.from_key(key)
    info = json.loads(jsonfile.read_text())
    contract = w3.eth.contract(abi=info["abi"], bytecode=info["bytecode"])
    tx = contract.constructor(*args).build_transaction({"from": acct.address})
    txreceipt = send_transaction(w3, tx, key)
    assert txreceipt.status == 1
    address = txreceipt.contractAddress
    return w3.eth.contract(address=address, abi=info["abi"]), txreceipt


def register_ibc_coin(cli, proposal, proposer_addr=ADDRS["validator"]):
    """
    submits a register_coin proposal for the provided coin metadata
    """
    proposer = eth_to_bech32(proposer_addr)
    # save the coin metadata in a json file
    with tempfile.NamedTemporaryFile("w") as meta_file:
        json.dump({"metadata": proposal.get("metadata")}, meta_file)
        meta_file.flush()
        proposal["metadata"] = meta_file.name
        rsp = cli.gov_legacy_proposal(proposer, "register-coin", proposal, gas=10000000)
        assert rsp["code"] == 0, rsp["raw_log"]
        txhash = rsp["txhash"]
        wait_for_new_blocks(cli, 2)
        receipt = cli.tx_search_rpc(f"tx.hash='{txhash}'")[0]
        return get_event_attribute_value(
            receipt["tx_result"]["events"],
            "submit_proposal",
            "proposal_id",
        )


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
    governance proposal, votes it, wait till it passes
    and checks that the host zone was registered successfully
    """
    prev_registered_zones = len(stride.cosmos_cli().get_host_zones())

    msg = stride.cosmos_cli().register_host_zone_msg(
        connection_id,
        host_denom,
        bech32_prefix,
        ibc_denom,
        channel_id,
        unbonding_frequency,
    )
    proposal = {
        "messages": [msg],
        "deposit": "1ustrd",
        "title": f"Register {bech32_prefix} zone",
        "summary": f"Proposal to register {bech32_prefix} zone",
    }
    with tempfile.NamedTemporaryFile("w") as proposal_file:
        json.dump(proposal, proposal_file)
        proposal_file.flush()
        rsp = stride.cosmos_cli().gov_proposal(proposer, proposal_file.name)
        assert rsp["code"] == 0, rsp["raw_log"]
        txhash = rsp["txhash"]

    wait_for_new_blocks(stride.cosmos_cli(), 2)
    receipt = stride.cosmos_cli().tx_search_rpc(f"tx.hash='{txhash}'")[0]
    proposal_id = get_event_attribute_value(
        receipt["tx_result"]["events"],
        "submit_proposal",
        "proposal_id",
    )
    assert int(proposal_id) > 0
    # vote 'yes' on proposal and wait it to pass
    approve_proposal(stride, proposal_id, gas_prices="2000000ustrd")

    # query token pairs and get WEVMOS address
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
            if type(attr["key"]) is str:
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
        if section == f"program:evmos_9000-1-node{i}":
            ini[section].update(
                {
                    "command": f"{cmd} start --home %(here)s/node{i}",
                    "autorestart": "false",  # don't restart when stopped
                }
            )
    with ini_path.open("w") as fp:
        ini.write(fp)


def update_evmos_bin(modified_bin, nodes=[0, 1]):
    """
    updates the evmos binary with a patched binary.
    Input parameters are the modified binary (modified_bin)
    and the nodes in which
    to apply the modified binary (nodes).
    Usually the setup comprise only 2 nodes (node0 & node1),
    so nodes should be an array containing only 0 and/or 1
    """

    def inner(path, base_port, config):
        chain_id = "evmos_9000-1"
        # by default, there're 2 nodes
        # need to update the bin in all these
        for i in nodes:
            update_node_cmd(path / chain_id, modified_bin, i)

    return inner


def register_wevmos(evmos):
    """
    this helper function registers the WEVMOS
    token in the ERC20 module and returns the contract address.
    Make sure to patch the evmosd binary with the allow-wevmos-register patch
    for this to be successful
    """
    cli = evmos.cosmos_cli()
    proposal = {
        "title": "Register WEVMOS",
        "description": "EVMOS erc20 representation",
        "metadata": [WEVMOS_META],
        "deposit": "1aevmos",
    }
    proposal_id = register_ibc_coin(cli, proposal)
    assert (
        int(proposal_id) > 0
    ), "expected a non-zero proposal ID for the registration of the WEVMOS token."
    # vote 'yes' on proposal and wait it to pass
    approve_proposal(evmos, proposal_id)

    # query token pairs and get WEVMOS address
    pairs = cli.get_token_pairs()
    assert len(pairs) == 1
    assert pairs[0]["denom"] == "aevmos"

    return pairs[0]["erc20_address"]


def wrap_evmos(evmos, addr, amt):
    """
    Helper function that registers WEVMOS token
    and wraps the specified amount
    for the provided Ethereum address
    Returns the WEVMOS contract address
    """
    cli = evmos.cosmos_cli()
    # submit proposal to register WEVMOS
    wevmos_addr = register_wevmos(evmos)

    # convert 'aevmos' to WEVMOS (wrap)
    rsp = cli.convert_coin(f"{amt}aevmos", eth_to_bech32(addr), gas=2000000)
    assert rsp["code"] == 0, rsp["raw_log"]
    wait_for_new_blocks(cli, 2)
    txhash = rsp["txhash"]
    receipt = cli.tx_search_rpc(f"tx.hash='{txhash}'")[0]
    assert receipt["tx_result"]["code"] == 0

    # check the desired amt was wrapped
    wevmos_balance = erc20_balance(evmos.w3, wevmos_addr, addr)
    assert wevmos_balance == amt

    return wevmos_addr


def erc20_balance(w3, erc20_contract_addr, addr):
    info = json.loads(CONTRACTS["IERC20"].read_text())
    contract = w3.eth.contract(erc20_contract_addr, abi=info["abi"])
    return contract.functions.balanceOf(addr).call()
