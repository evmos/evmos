import pytest

from .ibc_utils import assert_ready, get_balance, prepare_network
from .network import CosmosChain, Evmos
from .utils import ADDRS, eth_to_bech32, wait_for_ack, wait_for_cosmos_tx_receipt

# The token factory IBC denom on Evmos
TOKEN_FACTORY_IBC_DENOM = (
    "ibc/19616F5020D74FD2314577BF0B0CB99615C4C959665E308646291AF3B35FA4F2"
)


@pytest.fixture(scope="module", params=["evmos", "evmos-6dec", "evmos-rocksdb"])
def ibc(request, tmp_path_factory):
    """Prepare the network"""
    name = "str-v2-token-factory"
    evmos_build = request.param
    path = tmp_path_factory.mktemp(name)
    # specify the custom_scenario
    network = prepare_network(path, name, [evmos_build, "osmosis"])
    yield from network


def test_str_v2_token_factory(ibc):
    """
    Test Single Token Representation v2 with a token factory Coin from Osmosis.
    It should not create an ECR20 extension contract for the token factory coin.
    """
    assert_ready(ibc)

    evmos: Evmos = ibc.chains["evmos"]
    osmosis: CosmosChain = ibc.chains["osmosis"]

    evmos_cli = evmos.cosmos_cli()
    evmos_addr = ADDRS["signer2"]
    bech_dst = eth_to_bech32(evmos_addr)

    osmosis_cli = osmosis.cosmos_cli()
    osmosis_addr = osmosis_cli.address("signer2")

    # create a token factory coin
    token_factory_denom = create_token_factory_coin("utest", osmosis_addr, osmosis_cli)
    rsp = osmosis_cli.ibc_transfer(
        osmosis_addr,
        bech_dst,
        f"100{token_factory_denom}",
        "channel-0",
        1,
        fees="100000uosmo",
    )
    assert rsp["code"] == 0

    wait_for_ack(evmos_cli, "Evmos")

    token_pairs = evmos_cli.get_token_pairs()
    assert len(token_pairs) == 1

    active_dynamic_precompiles = evmos_cli.erc20_params()["params"][
        "dynamic_precompiles"
    ]
    assert len(active_dynamic_precompiles) == 0

    balance = get_balance(evmos, bech_dst, TOKEN_FACTORY_IBC_DENOM)
    assert balance == 100


def create_token_factory_coin(denom, creator_addr, osmosis_cli):
    full_denom = f"factory/{creator_addr}/{denom}"
    rsp = osmosis_cli.token_factory_create_denom(denom, creator_addr)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    print("Created token factory token", full_denom)

    rsp = osmosis_cli.token_factory_mint_denom(creator_addr, 1000, full_denom)
    assert rsp["code"] == 0

    # check for tx receipt to confirm tx was successful
    receipt = wait_for_cosmos_tx_receipt(osmosis_cli, rsp["txhash"])
    assert receipt["tx_result"]["code"] == 0
    print("Minted token factory token", full_denom)

    return full_denom
