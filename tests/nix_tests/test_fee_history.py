from concurrent.futures import ThreadPoolExecutor, as_completed

import pytest
from web3 import Web3

from .network import setup_evmos, setup_evmos_rocksdb, setup_geth
from .utils import ADDRS, send_transaction, w3_wait_for_block


@pytest.fixture(scope="module")
def custom_evmos(tmp_path_factory):
    path = tmp_path_factory.mktemp("fee-history")
    yield from setup_evmos(path, 25040)


@pytest.fixture(scope="module")
def custom_evmos_rocksdb(tmp_path_factory):
    path = tmp_path_factory.mktemp("fee-history-rocksdb")
    yield from setup_evmos_rocksdb(path, 26040)


@pytest.fixture(scope="module")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8565)


@pytest.fixture(scope="module", params=["evmos", "evmos-rocksdb", "geth"])
def cluster(request, custom_evmos, custom_evmos_rocksdb, geth):
    """
    run on evmos, evmos built with rocksdb and geth
    """
    provider = request.param
    if provider == "evmos":
        yield custom_evmos
    elif provider == "evmos-rocksdb":
        yield custom_evmos_rocksdb
    elif provider == "geth":
        yield geth
    else:
        raise NotImplementedError


def test_basic(cluster):
    w3: Web3 = cluster.w3
    call = w3.provider.make_request
    tx = {"to": ADDRS["community"], "value": 10, "gasPrice": w3.eth.gas_price}
    send_transaction(w3, tx)
    size = 4
    # size of base fee + next fee
    max = size + 1
    # only 1 base fee + next fee
    min = 2
    method = "eth_feeHistory"
    field = "baseFeePerGas"
    percentiles = [100]
    height = w3.eth.block_number
    latest = dict(
        blocks=["latest", hex(height)],
        expect=max,
    )
    earliest = dict(
        blocks=["earliest", "0x0"],
        expect=min,
    )
    for tc in [latest, earliest]:
        res = []
        with ThreadPoolExecutor(len(tc["blocks"])) as exec:
            tasks = [
                exec.submit(call, method, [size, b, percentiles]) for b in tc["blocks"]
            ]
            res = [future.result()["result"][field] for future in as_completed(tasks)]
        assert len(res) == len(tc["blocks"])
        assert res[0] == res[1]
        assert len(res[0]) == tc["expect"]

    # make sure that chain is at least at max height
    # to avoid having an error on fee_history call
    w3_wait_for_block(w3, max)
    for x in range(max):
        i = x + 1
        fee_history = call(method, [size, hex(i), percentiles])
        # start to reduce diff on i <= size - min
        diff = size - min - i
        reduce = size - diff
        target = reduce if diff >= 0 else max
        res = fee_history["result"]
        assert len(res[field]) == target
        oldest = i + min - max
        assert res["oldestBlock"] == hex(oldest if oldest > 0 else 0)
