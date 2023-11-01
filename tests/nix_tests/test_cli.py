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
            "exp_out": f'"last_commit":{{"height":{last_block-1}',
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
            "err_msg": f"invalid height, the latest height found in the db is {last_block}, and you asked for {last_block+10}",  # noqa: E501 - ignore line too long linter
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
            else:
                print(f"Unexpected {err=}, {type(err)=}")
                raise

    # start node1 again
    supervisorctl(
        evmos_cluster.base_dir / "../tasks.ini", "start", "evmos_9000-1-node1"
    )
    # check if chain continues alright
    wait_for_block(node1, last_block + 3)
