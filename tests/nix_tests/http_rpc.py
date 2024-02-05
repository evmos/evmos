import json
import subprocess

####################################
#          CometBFT RPC            #
####################################


def comet_status(port: int):
    """
    Gets node status from the CometBFT RPC
    """
    output = subprocess.getoutput(
        f"curl -s -X GET 'http://127.0.0.1:{port}/status' | jq"
    )
    if output == "":
        return None
    return json.loads(output)["result"]
