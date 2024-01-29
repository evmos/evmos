import pytest

from .network import setup_evmos, setup_evmos_rocksdb, setup_geth


@pytest.fixture(scope="session")
def evmos(tmp_path_factory):
    path = tmp_path_factory.mktemp("evmos")
    yield from setup_evmos(path, 26650)


@pytest.fixture(scope="session")
def evmos_rocksdb(tmp_path_factory):
    path = tmp_path_factory.mktemp("evmos-rocksdb")
    yield from setup_evmos_rocksdb(path, 20650)


@pytest.fixture(scope="session")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8545)


@pytest.fixture(scope="session", params=["evmos", "evmos-ws"])
def evmos_rpc_ws(request, evmos):
    """
    run on both evmos and evmos websocket
    """
    provider = request.param
    if provider == "evmos":
        yield evmos
    elif provider == "evmos-ws":
        evmos_ws = evmos.copy()
        evmos_ws.use_websocket()
        yield evmos_ws
    else:
        raise NotImplementedError


@pytest.fixture(scope="module", params=["evmos", "evmos-ws", "evmos-rocksdb", "geth"])
def cluster(request, evmos, evmos_rocksdb, geth):
    """
    run on evmos, evmos websocket,
    evmos built with rocksdb (memIAVL + versionDB)
    and geth
    """
    provider = request.param
    if provider == "evmos":
        yield evmos
    elif provider == "evmos-ws":
        evmos_ws = evmos.copy()
        evmos_ws.use_websocket()
        yield evmos_ws
    elif provider == "geth":
        yield geth
    elif provider == "evmos-rocksdb":
        yield evmos_rocksdb
    else:
        raise NotImplementedError


@pytest.fixture(scope="module", params=["evmos", "evmos-rocksdb"])
def evmos_cluster(request, evmos, evmos_rocksdb):
    """
    run on evmos default build &
    evmos with rocksdb build and memIAVL + versionDB
    """
    provider = request.param
    if provider == "evmos":
        yield evmos
    elif provider == "evmos-rocksdb":
        yield evmos_rocksdb
    else:
        raise NotImplementedError
