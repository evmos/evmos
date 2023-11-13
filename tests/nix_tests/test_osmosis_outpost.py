import pytest
import json

from .ibc_utils import (
    EVMOS_IBC_DENOM,
    OSMO_IBC_DENOM,
    assert_ready,
    get_balance,
    prepare_network,
)
from .utils import (
    ADDRS,
    KEYS,
    OSMOSIS_POOLS,
    WASM_CONTRACTS,
    assert_successful_cosmos_tx,
    eth_to_bech32,
    get_event_attribute_value,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
    wrap_evmos,
    register_ibc_coin,
    approve_proposal,
    erc20_balance,
    wait_for_cosmos_tx_receipt,
)
from .network import Evmos, CosmosChain


@pytest.fixture(scope="module", params=["evmos"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "osmosis-outpost"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # Setup the IBC connections
    network = prepare_network(
        path, name, [evmos_build, "osmosis"], custom_scenario=name
    )
    yield from network


def test_osmosis_swap(ibc):
    assert_ready(ibc)
    evmos: Evmos = ibc.chains["evmos"]
    osmosis: CosmosChain = ibc.chains["osmosis"]

    evmos_addr = ADDRS["signer2"]

    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")
    amt = 1000000000000000000

    setup_osmos_chains(ibc)

    # --------- Register Evmos token (this could be wrapevmos I think)
    wevmos_addr = wrap_evmos(ibc.chains["evmos"], evmos_addr, amt)

    # --------- Transfer Osmo to Evmos
    transfer_osmo_to_evmos(ibc, osmosis_addr, evmos_addr)

    # --------- Register Osmosis ERC20 token
    osmo_erc20_addr = register_osmo_token(evmos)

    # --------- Register contract on osmosis ??

    # define arguments
    testSlippagePercentage = 10
    testWindowSeconds = 20
    swap_amount = 1000000000000000000

    # --------- Swap Osmo to Evmos
    w3 = evmos.w3
    pc = get_precompile_contract(w3, "IOsmosisOutpost")
    evmos_gas_price = w3.eth.gas_price

    tx = pc.functions.swap(
        evmos_addr,
        wevmos_addr,
        osmo_erc20_addr,
        swap_amount,
        testSlippagePercentage,
        testWindowSeconds,
        osmosis_addr,
    ).build_transaction({"from": evmos_addr, "gasPrice": evmos_gas_price})
    gas_estimation = evmos.w3.eth.estimate_gas(tx)
    receipt = send_transaction(w3, tx, KEYS["signer2"])

    print(receipt)
    assert receipt.status == 1
    # check gas estimation is accurate
    assert receipt.gasUsed == gas_estimation

    # check if osmos was received
    new_src_balance = erc20_balance(w3, osmo_erc20_addr, evmos_addr)
    print(new_src_balance)
    assert new_src_balance == swap_amount


def setup_osmos_chains(ibc):
    # Send Evmos to Osmosis to be able to set up pools
    send_evmos_to_osmos(ibc)

    osmosis = ibc.chains["osmosis"]
    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")

    # create evmos <> osmo pool
    rsp = osmosis_cli.gamm_create_pool(osmosis_addr, OSMOSIS_POOLS["Evmos_Osmo"])
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    assert_successful_cosmos_tx(osmosis_cli, rsp["txhash"])

    contracts_to_store = {
        "CrosschainRegistry": {
            "get_instantiate_params": lambda x: f'"{{\\"owner\\":\\"{x}\\"}}"',
        },
        "Swaprouter": {
            "get_instantiate_params": lambda x: f'"{{\\"owner\\":\\"{x}\\"}}"',
        },
        "CrosschainSwap": {
            "get_instantiate_params": lambda x, y, z: f'"{{\\"governor\\":\\"{x}\\", \\"swap_contract\\": \\"{y}\\", \\"registry_contract\\": \\"{z}\\"}}"',
        },
    }

    # ===== Deploy CrosschainRegistry =====
    registry_contract = WASM_CONTRACTS["CrosschainRegistry"]
    _, registry_contract_addr = deploy_wasm_contract(
        osmosis_cli,
        osmosis_addr,
        registry_contract,
        contracts_to_store["CrosschainRegistry"]["get_instantiate_params"](
            osmosis_addr
        ),
        "xcregistry1.0",
    )

    # ===== Deploy Swaprouter =====
    swap_contract = WASM_CONTRACTS["Swaprouter"]
    _, swap_contract_addr = deploy_wasm_contract(
        osmosis_cli,
        osmosis_addr,
        swap_contract,
        contracts_to_store["Swaprouter"]["get_instantiate_params"](osmosis_addr),
        "swaprouter1.0",
    )

    # ===== Deploy CrosschainSwap =====
    cross_swap_contract = WASM_CONTRACTS["CrosschainSwap"]
    _, cross_swap_contract_addr = deploy_wasm_contract(
        osmosis_cli,
        osmosis_addr,
        cross_swap_contract,
        contracts_to_store["CrosschainSwap"]["get_instantiate_params"](
            osmosis_addr, swap_contract_addr, registry_contract_addr
        ),
        "xcswap1.0",
    )
    # =================================

    # in the router one execute function `set_route` to have a route for evmos within the swap router contract
    execute_args = '{{"set_route":{{"input_denom": "uosmo","output_denom":"aevmos","pool_route":[{{"pool_id": "1","token_out_denom":"aevmos"}}]}}'
    print("**********")
    print(execute_args)
    rsp = osmosis_cli.wasm_execute(swap_contract_addr, execute_args)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    receipt_file_path = f"/tmp/wasm_exec_receipt.json"
    # TODO remove
    with open(receipt_file_path, "w") as receipt_file:
        json.dump(receipt, receipt_file, indent=2)
    # TODO remove ^^^^

    assert receipt["tx_result"]["code"] == 0


def send_evmos_to_osmos(ibc):
    src_chain = ibc.chains["evmos"]
    dst_chain = ibc.chains["osmosis"]

    dst_addr = dst_chain.cosmos_cli().address("signer2")
    amt = 1000000

    cli = src_chain.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aevmos"

    old_src_balance = get_balance(src_chain, src_addr, src_denom)
    old_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)

    pc = get_precompile_contract(src_chain.w3, "ICS20I")
    evmos_gas_price = src_chain.w3.eth.gas_price

    tx_hash = pc.functions.transfer(
        "transfer",
        "channel-0",  # Connection with Osmosis is on channel-1
        src_denom,
        amt,
        ADDRS["signer2"],
        dst_addr,
        [1, 10000000000],
        0,
        "",
    ).transact({"from": ADDRS["signer2"], "gasPrice": evmos_gas_price})

    receipt = src_chain.w3.eth.wait_for_transaction_receipt(tx_hash)

    assert receipt.status == 1
    # check gas used
    assert receipt.gasUsed == 74098

    fee = receipt.gasUsed * evmos_gas_price

    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    assert old_dst_balance + amt == new_dst_balance
    new_src_balance = get_balance(src_chain, src_addr, src_denom)
    assert old_src_balance - amt - fee == new_src_balance


def transfer_osmo_to_evmos(ibc, src_addr, dst_addr):
    src_chain: CosmosChain = ibc.chains["osmosis"]
    dst_chain: Evmos = ibc.chains["evmos"]

    cli = src_chain.cosmos_cli()
    src_addr = cli.address("signer2")

    bech_dst = eth_to_bech32(dst_addr)
    old_dst_balance = get_balance(dst_chain, bech_dst, OSMO_IBC_DENOM)
    rsp = (
        ibc.chains["osmosis"]
        .cosmos_cli()
        .ibc_transfer(src_addr, bech_dst, "200uosmo", "channel-0", 1, fees="10000uosmo")
    )
    assert rsp["code"] == 0

    # TODO: This needs to be changed to the osmosis ibc denom
    # old_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
    new_dst_balance = 0

    def check_balance_change():
        nonlocal new_dst_balance
        new_dst_balance = get_balance(dst_chain, bech_dst, OSMO_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)

    # TODO: This needs to be changed to the osmosis ibc denom
    # new_dst_balance = get_balance(dst_chain, dst_addr, OSMO_IBC_DENOM)
    # assert new_dst_balance == amt


def register_osmo_token(evmos):
    evmos_cli = evmos.cosmos_cli()

    # --------- Register Osmosis ERC20 token
    # > For that I need the denom trace taken from the ibc info
    # >

    # TODO - generate the osmos ibc denom
    osmos_ibc_denom = OSMO_IBC_DENOM
    ERC_OSMO_META = {
        "description": "Generic IBC token description",
        "denom_units": [
            # TODO - generate the osmos ibc denom
            {
                "denom": osmos_ibc_denom,
                "exponent": 0,
            },
        ],
        # TODO - generate the osmos ibc denom
        "base": osmos_ibc_denom,
        "display": osmos_ibc_denom,
        "name": "Generic IBC name",
        "symbol": "IBC",
    }

    proposal = {
        "title": "Register Osmosis ERC20 token",
        "description": "The IBC representation of OSMO on Evmos chain",
        "metadata": [ERC_OSMO_META],
        "deposit": "1aevmos",
    }
    proposal_id = register_ibc_coin(evmos_cli, proposal)
    assert (
        int(proposal_id) > 0
    ), "expected a non-zero proposal ID for the registration of the OSMO token."
    print("proposal id: ", proposal_id)
    # vote 'yes' on proposal and wait it to pass
    approve_proposal(evmos, proposal_id)
    # query token pairs and get WEVMOS address
    pairs = evmos_cli.get_token_pairs()
    assert len(pairs) == 2
    assert pairs[1]["denom"] == osmos_ibc_denom
    return pairs[1]["erc20_address"]


def deploy_wasm_contract(osmosis_cli, deployer_addr, contract_file, init_args, label):
    """
    Stores the contract binary and deploys one instance of it.
    Returns the contract code id
    """
    # 1. Store the binary
    rsp = osmosis_cli.wasm_store_binary(deployer_addr, contract_file)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    # get code_id from the receipt logs
    logs = json.loads(receipt["tx_result"]["log"])
    code_id = get_event_attribute_value(logs[0]["events"], "store_code", "code_id")

    # 2. instantiate contract
    rsp = osmosis_cli.wasm_instante2(
        deployer_addr,
        code_id,
        init_args,
        label,
    )
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    receipt_file_path = f"/tmp/{label}_receipt.json"
    # TODO remove
    with open(receipt_file_path, "w") as receipt_file:
        json.dump(receipt, receipt_file, indent=2)
    # TODO remove ^^^^
    assert receipt["tx_result"]["code"] == 0
    
    # get instantiated contract address from events in logs
    logs = json.loads(receipt["tx_result"]["log"])
    contract_addr = get_event_attribute_value(
        logs[0]["events"], "instantiate", "_contract_address"
    )
    print(f"deployed {label} CosmWasm contract @ {contract_addr}")

    return code_id, contract_addr
