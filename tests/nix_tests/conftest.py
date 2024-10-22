import pytest

from .network import setup_eidon-chain, setup_eidon-chain_rocksdb, setup_geth


@pytest.fixture(scope="session")
def eidon-chain(tmp_path_factory):
    path = tmp_path_factory.mktemp("eidon-chain")
    yield from setup_eidon-chain(path, 26650)


@pytest.fixture(scope="session")
def eidon-chain_rocksdb(tmp_path_factory):
    path = tmp_path_factory.mktemp("eidon-chain-rocksdb")
    yield from setup_eidon-chain_rocksdb(path, 20650)


@pytest.fixture(scope="session")
def geth(tmp_path_factory):
    path = tmp_path_factory.mktemp("geth")
    yield from setup_geth(path, 8545)


@pytest.fixture(scope="session", params=["eidon-chain", "eidon-chain-ws"])
def eidon-chain_rpc_ws(request, eidon-chain):
    """
    run on both eidon-chain and eidon-chain websocket
    """
    provider = request.param
    if provider == "eidon-chain":
        yield eidon-chain
    elif provider == "eidon-chain-ws":
        eidon-chain_ws = eidon-chain.copy()
        eidon-chain_ws.use_websocket()
        yield eidon-chain_ws
    else:
        raise NotImplementedError


@pytest.fixture(scope="module", params=["eidon-chain", "eidon-chain-ws", "eidon-chain-rocksdb", "geth"])
def cluster(request, eidon-chain, eidon-chain_rocksdb, geth):
    """
    run on eidon-chain, eidon-chain websocket,
    eidon-chain built with rocksdb (memIAVL + versionDB)
    and geth
    """
    provider = request.param
    if provider == "eidon-chain":
        yield eidon-chain
    elif provider == "eidon-chain-ws":
        eidon-chain_ws = eidon-chain.copy()
        eidon-chain_ws.use_websocket()
        yield eidon-chain_ws
    elif provider == "geth":
        yield geth
    elif provider == "eidon-chain-rocksdb":
        yield eidon-chain_rocksdb
    else:
        raise NotImplementedError


@pytest.fixture(scope="module", params=["eidon-chain", "eidon-chain-rocksdb"])
def eidon-chain_cluster(request, eidon-chain, eidon-chain_rocksdb):
    """
    run on eidon-chain default build &
    eidon-chain with rocksdb build and memIAVL + versionDB
    """
    provider = request.param
    if provider == "eidon-chain":
        yield eidon-chain
    elif provider == "eidon-chain-rocksdb":
        yield eidon-chain_rocksdb
    else:
        raise NotImplementedError
