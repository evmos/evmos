from .utils import get_current_height, supervisorctl, wait_for_block


def test_block_cmd(evmos_cluster):
    """
    - start 2 evmos nodes
    - wait for a certain height
    - stop the node1
    - use the 'block' cli cmd
    - restart evmos node1
    """

    # wait for specific height
    node1 = evmos_cluster.cosmos_cli(1)
    current_height = get_current_height(node1)

    last_block = current_height + 2
    wait_for_block(node1, last_block)

    # stop node1
    supervisorctl(evmos_cluster.base_dir / "../tasks.ini", "stop", "evmos_9000-1-node1")

    # use 'block' CLI cmd in node1
    test_cases = [
        {
            "name": "success - get latest block",
            "flags": [],
            "exp_out": f'"last_commit":{{"height":{last_block - 1}',
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - get block #2",
            "flags": ["--height", 2],
            "exp_out": '"height":2',
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "fail - get inexistent block",
            "flags": ["--height", last_block + 10],
            "exp_out": None,
            "exp_err": True,
            "err_msg": f"invalid height, the latest height found in the db is {last_block}, "
            f"and you asked for {last_block + 10}",
        },
    ]
    for tc in test_cases:
        try:
            output = node1.raw("block", *tc["flags"], home=node1.data_dir)
            assert tc["exp_out"] in output.decode()
        except Exception as err:
            if tc["exp_err"] is True:
                assert tc["err_msg"] in err.args[0]
                continue

            print(f"Unexpected {err=}, {type(err)=}")
            raise

    # start node1 again
    supervisorctl(
        evmos_cluster.base_dir / "../tasks.ini", "start", "evmos_9000-1-node1"
    )
    # check if chain continues alright
    wait_for_block(node1, last_block + 3)


# TODO: why is the signer1 account not found?
def test_tx_flags(evmos_cluster):
    """
    Tests the expected responses for common fee and gas related CLI flags.
    """

    node = evmos_cluster.cosmos_cli(0)
    current_height = get_current_height(node)
    wait_for_block(node, current_height + 1)

    test_cases = [
        {
            "name": "fail - invalid flags combination (gas-prices & fees)",
            "flags": ["--fees=5000000000aevmos", "--gas-prices=50000aevmos"],
            "exp_err": True,
            "err_msg": "cannot provide both fees and gas prices",
        },
        {
            "name": "fail - no fees & insufficient gas",
            "flags": ["--gas=50000"],
            "exp_err": True,
            "err_msg": "gas prices too low",
        },
        {
            "name": "fail - insufficient fees",
            "flags": ["--fees=10aevmos", "--gas=50000"],
            "exp_err": True,
            "err_msg": "insufficient fee",
        },
        {
            "name": "fail - insufficient gas",
            "flags": ["--fees=500000000000aevmos", "--gas=1000"],
            "exp_err": True,
            "err_msg": "out of gas",
        },
        {
            "name": "success - defined fees & gas",
            "flags": ["--fees=10000000000000000aevmos", "--gas=1500000"],
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - using gas & gas-prices",
            "flags": ["--gas-prices=1000000000aevmos", "--gas=1500000"],
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - using gas 'auto' and specific fees",
            "flags": ["--gas=auto", "--fees=10000000000000000aevmos"],
            "exp_err": False,
            "err_msg": None,
        },
    ]

    for tc in test_cases:
        try:
            node.raw(
                "tx",
                "bank",
                "send",
                "signer1",
                "evmos10jmp6sgh4cc6zt3e8gw05wavvejgr5pwjnpcky",
                "100000000000000aevmos",
                *tc["flags"],
                home=node.data_dir,
            )
            assert not tc["exp_err"], "expected error to be found; got none"
        except Exception as err:
            if tc["exp_err"] is True:
                assert (
                    tc["err_msg"] in err.args[0]
                ), "expected different error to be found"
                continue

            print(f"Unexpected {err=}, {type(err)=}")
            raise
