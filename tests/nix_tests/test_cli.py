from .utils import supervisorctl, wait_for_block


def test_block_cmd(evmos):
    """
    - start 2 evmos nodes
    - wait for a certain height
    - stop the node1
    - use the 'block' cli cmd
    - restart evmos node1
    """

    # wait for height 10
    node1 = evmos.cosmos_cli(1)
    wait_for_block(node1, 10)

    # stop node1
    supervisorctl(evmos.base_dir / "../tasks.ini", "stop", "evmos_9000-1-node1")

    # use 'block' CLI cmd in node1
    test_cases = [
        {
            "name": "success - get latest block",
            "flags": [],
            "exp_out": '"last_commit":{"height":9',
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
            "flags": ["--height", 20],
            "exp_out": None,
            "exp_err": True,
            "err_msg": "invalid height, the latest height found in the db is 10, and you asked for 20",
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
    supervisorctl(evmos.base_dir / "../tasks.ini", "start", "evmos_9000-1-node1")
    # check is chain continues alright
    wait_for_block(node1, 12)
