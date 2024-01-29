import os
from cgi import test
import multiprocessing
import glob
import time


def run_tests(test_file, failed_tests):
    cmd = f"pytest {test_file} -s"
    print(cmd)
    exit_status = os.system(cmd)
    if exit_status != 0:
        failed_tests.append(test_file)


if __name__ == "__main__":
    start_time = time.time()
    # Set TMPDIR to /tmp
    # needed to force the tmp_dir_factory to use the /tmp dir on the processes
    # instead of creating a tmpfs for each
    os.environ["TMPDIR"] = "/tmp"

    # Find all test files in the current directory
    test_files = glob.glob("test_*.py")

    # These are the test files that should run in parallel.
    # These will spin up their own chain and run the tests within this file sequentially,
    # but in parallel to other tests running on other processes
    parallel_tests_files = [
        "test_filters.py",
        "test_account.py",
        "test_fee_history.py",
        "test_grpc_only.py",
        "test_ibc.py",
        "test_no_abci_resp.py",
        "test_osmosis_outpost.py",
        "test_precompiles.py",
        "test_priority.py",
        "test_pruned_node.py",
        "test_rollback.py",
        "test_stride_outpost.py",
        "test_storage_proof.py",
        "test_zero_fee.py",
    ]

    # Remove files in parallel_tests_files from test_files
    test_files = [file for file in test_files if file not in parallel_tests_files]
    # Sort the test_files alphabetically
    test_files.sort()

    # Create a shared list to store files with failed tests
    manager = multiprocessing.Manager()
    failed_tests = manager.list()

    # Run sequential tests on the same process
    seq_tests = multiprocessing.Process(
        target=run_tests, args=(" ".join(test_files), failed_tests)
    )

    # Gather the different processes running the tests
    processes = [seq_tests]

    # start the sequential tests process
    seq_tests.start()

    for test_file in parallel_tests_files:
        # wait for nodes of other tests to start and avoid port collision
        process = multiprocessing.Process(
            target=run_tests, args=(test_file, failed_tests)
        )
        processes.append(process)
        process.start()

    # Wait for all processes to finish
    for process in processes:
        process.join()

    elapsed_time = time.time() - start_time
    print(f"Elapsed time: {elapsed_time / 60:.2f} minutes")

    # Print files with failed tests
    if len(failed_tests) > 0:
        print("Files with FAILED tests:")
        for file in failed_tests:
            print(file)
        exit(1)
