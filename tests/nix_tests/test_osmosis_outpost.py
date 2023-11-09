import pytest

from .ibc_utils import EVMOS_IBC_DENOM, assert_ready, get_balance, prepare_network
from .utils import (
    ADDRS,
    KEYS,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
    wrap_evmos,
    register_ibc_coin,
    approve_proposal,
    erc20_balance,
)
from .network import Evmos, CosmosChain


@pytest.fixture(scope="module", params=["evmos", "evmos-rocksdb"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "osmosis-outpost"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # Setup the IBC connections
    # evmos     (channel-0) <> (channel-0)  gaia
    # evmos     (channel-1) <> (channel-0)  osmosis
    # osmosis   (channel-1) <> (channel-1)  gaia
    network = prepare_network(path, name, [evmos_build, "gaia", "osmosis"])
    yield from network


# TODO remove this test and replace with the outpost test
def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)

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
        "channel-1",  # Connection with Osmosis is on channel-1
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

def test_osmosis_swap(ibc):
    assert_ready(ibc)
    sender_addr = ADDRS["signer2"]
    amt = 1000000000000000000

    evmos: Evmos = ibc.chains["evmos"]
    osmosis: CosmosChain = ibc.chains["osmosis"]

    # --------- Register Evmos token (this could be wrapevmos I think)
    wevmos_addr = wrap_evmos(ibc.chains["evmos"], sender_addr, amt)

    # --------- Transfer Osmo to Evmos
    transfer_osmo_to_evmos(ibc, sender_addr, wevmos_addr, amt)

    # --------- Register Osmosis ERC20 token
    osmo_erc20_addr = register_osmo_token(evmos)

    # --------- Register contract on osmosis ??

    # define arguments
    testSlippagePercentage = 10
    testWindowSeconds = 20
    swap_amount = 1000000000000000000

    osmosis_cli = osmosis.cosmos_cli()
    osmosis_receiver = osmosis_cli.address("signer2")
    args = [
        sender_addr,
        wevmos_addr,
        osmo_erc20_addr,
        swap_amount,
        testSlippagePercentage,
        testWindowSeconds,
        osmosis_receiver,
    ]

    # --------- Swap Osmo to Evmos
    w3 = evmos.w3
    pc = get_precompile_contract(w3, "OsmosisOutpostAddress")
    evmos_gas_price = w3.eth.gas_price

    tx = pc.functions.swap(*args).build_transaction(
        {"from": sender_addr, "gasPrice": evmos_gas_price}
    )
    gas_estimation = evmos.w3.eth.estimate_gas(tx)
    receipt = send_transaction(w3, tx, KEYS["signer2"])

    assert receipt.status == 1
    # check gas estimation is accurate
    assert receipt.gasUsed == gas_estimation

    # check if osmos was received
    new_src_balance = erc20_balance(w3, osmo_erc20_addr, sender_addr)
    print(new_src_balance)
    assert new_src_balance == swap_amount

def transfer_osmo_to_evmos(ibc, src_addr, dst_addr, amt):
    src_chain: CosmosChain = ibc.chains["osmosis"]
    dst_chain: Evmos = ibc.chains["evmos"]

    dst_addr = dst_chain.cosmos_cli().address("signer2")

    cli = src_chain.cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "uosmo"


    rsp = cli.ibc_transfer(
        src_addr,
        dst_addr,
        f"{amt}{src_denom}",
        "channel-0",
        1,
    )
    assert rsp["code"] == 0

    # TODO: This needs to be changed to the osmosis ibc denom
    old_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
    new_dst_balance = 0
    def check_balance_change():
        nonlocal new_dst_balance
        # TODO: This needs to be changed to the osmosis ibc denom
        new_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
        return old_dst_balance != new_dst_balance

    wait_for_fn("balance change", check_balance_change)
    # TODO: This needs to be changed to the osmosis ibc denom
    new_dst_balance = get_balance(dst_chain, dst_addr, EVMOS_IBC_DENOM)
    assert new_dst_balance == amt

def register_osmo_token(evmos):
    evmos_cli = evmos.cosmos_cli()

    # --------- Register Osmosis ERC20 token
    # > For that I need the denom trace taken from the ibc info
    # >

    # TODO - generate the osmos ibc denom
    osmos_ibc_denom = "uosmo"

    ERC_OSMO_META = {
        "description": "The native staking and governance token of the Evmos chain",
        "denom_units": [
            # TODO - generate the osmos ibc denom
            {"denom": osmos_ibc_denom, "exponent": 0, "aliases": ["aevmos"]},
            {"denom": "uosmo", "exponent": 18},
        ],
        # TODO - generate the osmos ibc denom
        "base": osmos_ibc_denom,
        "display": "osmos",
        "name": "Evmos Osmo Token",
        "symbol": "OSMOS",
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
    # vote 'yes' on proposal and wait it to pass
    approve_proposal(evmos, proposal_id)
    # query token pairs and get WEVMOS address
    pairs = evmos_cli.get_token_pairs()
    assert len(pairs) == 1
    assert pairs[0]["denom"] == osmos_ibc_denom
    return pairs[0]["erc20_address"]

