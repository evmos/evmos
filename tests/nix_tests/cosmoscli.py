import json
import tempfile

import requests
from dateutil.parser import isoparse
from pystarport.utils import build_cli_args_safe, interact

from .utils import DEFAULT_DENOM

DEFAULT_GAS_PRICE = f"5000000000000{DEFAULT_DENOM}"
DEFAULT_GAS = "250000"


class ChainCommand:
    def __init__(self, cmd):
        self.cmd = cmd

    def __call__(self, cmd, *args, stdin=None, **kwargs):
        "execute chain-maind"
        args = " ".join(build_cli_args_safe(cmd, *args, **kwargs))
        return interact(f"{self.cmd} {args}", input=stdin)


class CosmosCLI:
    "the apis to interact with wallet and blockchain"

    def __init__(
        self,
        data_dir,
        node_rpc,
        cmd,
    ):
        self.data_dir = data_dir
        self._genesis = json.loads(
            (self.data_dir / "config" / "genesis.json").read_text()
        )
        self.chain_id = self._genesis["chain_id"]
        self.node_rpc = node_rpc
        self.raw = ChainCommand(cmd)
        self.output = None
        self.error = None

    @property
    def node_rpc_http(self):
        return "http" + self.node_rpc.removeprefix("tcp")

    def init(self, moniker):
        "the node's config is already added"
        return self.raw(
            "init",
            moniker,
            chain_id=self.chain_id,
            home=self.data_dir,
        )

    def status(self):
        return json.loads(self.raw("status", node=self.node_rpc))

    def block_height(self):
        return int(self.status()["SyncInfo"]["latest_block_height"])

    def block_time(self):
        return isoparse(self.status()["SyncInfo"]["latest_block_time"])

    def rollback(self):
        self.raw("rollback", home=self.data_dir)

    # ==========================
    #       GENESIS cmds
    # ==========================

    def validate_genesis(self):
        return self.raw("validate-genesis", home=self.data_dir)

    def add_genesis_account(self, addr, coins, **kwargs):
        return self.raw(
            "add-genesis-account",
            addr,
            coins,
            home=self.data_dir,
            output="json",
            **kwargs,
        )

    def gentx(self, name, coins, min_self_delegation=1, pubkey=None):
        return self.raw(
            "gentx",
            name,
            coins,
            min_self_delegation=str(min_self_delegation),
            home=self.data_dir,
            chain_id=self.chain_id,
            keyring_backend="test",
            pubkey=pubkey,
        )

    def collect_gentxs(self, gentx_dir):
        return self.raw("collect-gentxs", gentx_dir, home=self.data_dir)

    # ==========================
    #     ACCOUNT KEYS utils
    # ==========================

    def migrate_keystore(self):
        return self.raw("keys", "migrate", home=self.data_dir)

    def address(self, name, bech="acc"):
        output = self.raw(
            "keys",
            "show",
            name,
            "-a",
            home=self.data_dir,
            keyring_backend="test",
            bech=bech,
        )
        return output.strip().decode()

    def create_account(self, name, mnemonic=None):
        "create new keypair in node's keyring"
        if mnemonic is None:
            output = self.raw(
                "keys",
                "add",
                name,
                home=self.data_dir,
                output="json",
                keyring_backend="test",
            )
        else:
            output = self.raw(
                "keys",
                "add",
                name,
                "--recover",
                home=self.data_dir,
                output="json",
                keyring_backend="test",
                stdin=mnemonic.encode() + b"\n",
            )
        return json.loads(output)

    def delete_account(self, name):
        "delete wallet account in node's keyring"
        return self.raw(
            "keys",
            "delete",
            name,
            "-y",
            "--force",
            home=self.data_dir,
            output="json",
            keyring_backend="test",
        )

    def make_multisig(self, name, signer1, signer2):
        self.raw(
            "keys",
            "add",
            name,
            multisig=f"{signer1},{signer2}",
            multisig_threshold="2",
            home=self.data_dir,
            keyring_backend="test",
        )

    # ==========================
    #        TX utils
    # ==========================
    def block_results_rpc(self):
        rsp = requests.get(f"{self.node_rpc_http}/block_results").json()
        assert "error" not in rsp, rsp["error"]
        return rsp["result"]

    def tx_search(self, events: str):
        "/tx_search"
        return json.loads(
            self.raw("query", "txs", events=events, output="json", node=self.node_rpc)
        )

    def tx_search_rpc(self, events: str):
        rsp = requests.get(
            f"{self.node_rpc_http}/tx_search",
            params={
                "query": f'"{events}"',
            },
        ).json()
        assert "error" not in rsp, rsp["error"]
        return rsp["result"]["txs"]

    def tx(self, value, **kwargs):
        "/tx"
        default_kwargs = {
            "home": self.data_dir,
        }
        return json.loads(self.raw("query", "tx", value, **(default_kwargs | kwargs)))

    def query_tx(self, tx_type, tx_value):
        tx = self.raw(
            "query",
            "tx",
            "--type",
            tx_type,
            tx_value,
            home=self.data_dir,
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return json.loads(tx)

    def query_all_txs(self, addr):
        txs = self.raw(
            "query",
            "txs-all",
            addr,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return json.loads(txs)

    def sign_multisig_tx(self, tx_file, multi_addr, signer_name):
        return json.loads(
            self.raw(
                "tx",
                "sign",
                tx_file,
                from_=signer_name,
                multisig=multi_addr,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def sign_batch_multisig_tx(
        self, tx_file, multi_addr, signer_name, account_number, sequence_number
    ):
        r = self.raw(
            "tx",
            "sign-batch",
            "--offline",
            tx_file,
            account_number=account_number,
            sequence=sequence_number,
            from_=signer_name,
            multisig=multi_addr,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return r.decode("utf-8")

    def encode_signed_tx(self, signed_tx):
        return self.raw(
            "tx",
            "encode",
            signed_tx,
        )

    def sign_tx(self, tx_file, signer):
        return json.loads(
            self.raw(
                "tx",
                "sign",
                tx_file,
                from_=signer,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def sign_tx_json(self, tx, signer, max_priority_price=None):
        if max_priority_price is not None:
            tx["body"]["extension_options"].append(
                {
                    "@type": "/ethermint.types.v1.ExtensionOptionDynamicFeeTx",
                    "max_priority_price": str(max_priority_price),
                }
            )
        with tempfile.NamedTemporaryFile("w") as fp:
            json.dump(tx, fp)
            fp.flush()
            return self.sign_tx(fp.name, signer)

    def combine_multisig_tx(self, tx_file, multi_name, signer1_file, signer2_file):
        return json.loads(
            self.raw(
                "tx",
                "multisign",
                tx_file,
                multi_name,
                signer1_file,
                signer2_file,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def combine_batch_multisig_tx(
        self, tx_file, multi_name, signer1_file, signer2_file
    ):
        r = self.raw(
            "tx",
            "multisign-batch",
            tx_file,
            multi_name,
            signer1_file,
            signer2_file,
            home=self.data_dir,
            keyring_backend="test",
            chain_id=self.chain_id,
            node=self.node_rpc,
        )
        return r.decode("utf-8")

    def broadcast_tx(self, tx_file, **kwargs):
        kwargs.setdefault("broadcast_mode", "sync")
        kwargs.setdefault("output", "json")
        return json.loads(
            self.raw("tx", "broadcast", tx_file, node=self.node_rpc, **kwargs)
        )

    def broadcast_tx_json(self, tx, **kwargs):
        with tempfile.NamedTemporaryFile("w") as fp:
            json.dump(tx, fp)
            fp.flush()
            return self.broadcast_tx(fp.name, **kwargs)

    # ==========================
    #       BANK module
    # ==========================

    def total_supply(self):
        return json.loads(
            self.raw("query", "bank", "total", output="json", node=self.node_rpc)
        )

    def transfer(self, from_, to, coins, generate_only=False, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "bank",
                "send",
                from_,
                to,
                coins,
                "-y",
                "--generate-only" if generate_only else None,
                home=self.data_dir,
                **kwargs,
            )
        )

    def balances(self, addr):
        return json.loads(
            self.raw("query", "bank", "balances", addr, home=self.data_dir)
        )["balances"]

    def balance(self, addr, denom=DEFAULT_DENOM):
        denoms = {coin["denom"]: int(coin["amount"]) for coin in self.balances(addr)}
        return denoms.get(denom, 0)

    # ==========================
    #    DISTRIBUTION module
    # ==========================

    def distribution_commission(self, addr):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "commission",
                addr,
                output="json",
                node=self.node_rpc,
            )
        )["commission"][0]
        return float(coin["amount"])

    def distribution_community(self):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "community-pool",
                output="json",
                node=self.node_rpc,
            )
        )["pool"][0]
        return float(coin["amount"])

    def distribution_reward(self, delegator_addr):
        coin = json.loads(
            self.raw(
                "query",
                "distribution",
                "rewards",
                delegator_addr,
                output="json",
                node=self.node_rpc,
            )
        )["total"][0]
        return float(coin["amount"])

    # from_delegator can be account name or address
    def withdraw_all_rewards(self, from_delegator):
        return json.loads(
            self.raw(
                "tx",
                "distribution",
                "withdraw-all-rewards",
                "-y",
                from_=from_delegator,
                home=self.data_dir,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    # ==========================
    #       SLASHING module
    # ==========================

    def unjail(self, addr):
        return json.loads(
            self.raw(
                "tx",
                "slashing",
                "unjail",
                "-y",
                from_=addr,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
            )
        )

    # ==========================
    #       STAKING module
    # ==========================

    def validator(self, addr):
        return json.loads(
            self.raw(
                "query",
                "staking",
                "validator",
                addr,
                output="json",
                node=self.node_rpc,
            )
        )

    def validators(self):
        return json.loads(
            self.raw(
                "query", "staking", "validators", output="json", node=self.node_rpc
            )
        )["validators"]

    def staking_params(self):
        return json.loads(
            self.raw("query", "staking", "params", output="json", node=self.node_rpc)
        )

    def staking_pool(self, bonded=True):
        return int(
            json.loads(
                self.raw("query", "staking", "pool", output="json", node=self.node_rpc)
            )["bonded_tokens" if bonded else "not_bonded_tokens"]
        )

    def get_delegated_amount(self, which_addr):
        return json.loads(
            self.raw(
                "query",
                "staking",
                "delegations",
                which_addr,
                home=self.data_dir,
                chain_id=self.chain_id,
                node=self.node_rpc,
                output="json",
            )
        )

    def delegate_amount(self, to_addr, amount, from_addr, **kwargs):
        kwargs.setdefault(
            "gas_prices", f"{self.query_base_fee() + 100000}{DEFAULT_DENOM}"
        )
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "delegate",
                to_addr,
                amount,
                "--generate-only",
                "--from",
                from_addr,
                home=self.data_dir,
                **kwargs,
            )
        )

    # to_addr: croclcl1...  , from_addr: cro1...
    def unbond_amount(self, to_addr, amount, from_addr):
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "unbond",
                to_addr,
                amount,
                "-y",
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    # to_validator_addr: crocncl1...  ,  from_from_validator_addraddr: crocl1...
    def redelegate_amount(
        self, to_validator_addr, from_validator_addr, amount, from_addr
    ):
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "redelegate",
                from_validator_addr,
                to_validator_addr,
                amount,
                "-y",
                home=self.data_dir,
                from_=from_addr,
                keyring_backend="test",
                chain_id=self.chain_id,
                node=self.node_rpc,
            )
        )

    def create_validator(
        self,
        amount,
        moniker=None,
        commission_max_change_rate="0.01",
        commission_rate="0.1",
        commission_max_rate="0.2",
        min_self_delegation="1",
        identity="",
        website="",
        security_contact="",
        details="",
    ):
        """MsgCreateValidator
        create the node with create_node before call this"""
        pubkey = (
            "'"
            + (
                self.raw(
                    "tendermint",
                    "show-validator",
                    home=self.data_dir,
                )
                .strip()
                .decode()
            )
            + "'"
        )
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "create-validator",
                "-y",
                from_=self.address("validator"),
                amount=amount,
                pubkey=pubkey,
                min_self_delegation=min_self_delegation,
                # commision
                commission_rate=commission_rate,
                commission_max_rate=commission_max_rate,
                commission_max_change_rate=commission_max_change_rate,
                # description
                moniker=moniker,
                identity=identity,
                website=website,
                security_contact=security_contact,
                details=details,
                # basic
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
            )
        )

    def edit_validator(
        self,
        commission_rate=None,
        moniker=None,
        identity=None,
        website=None,
        security_contact=None,
        details=None,
    ):
        """MsgEditValidator"""
        options = dict(
            commission_rate=commission_rate,
            # description
            moniker=moniker,
            identity=identity,
            website=website,
            security_contact=security_contact,
            details=details,
        )
        return json.loads(
            self.raw(
                "tx",
                "staking",
                "edit-validator",
                "-y",
                from_=self.address("validator"),
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
                **{k: v for k, v in options.items() if v is not None},
            )
        )

    # ==========================
    #         GOV module
    # ==========================
    def gov_proposal(self, proposer, proposal_file_name, **kwargs):
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "submit-proposal",
                proposal_file_name,
                "-y",
                from_=proposer,
                home=self.data_dir,
                **kwargs,
            )
        )

    def gov_legacy_proposal(self, proposer, kind, proposal, **kwargs):
        method = "submit-legacy-proposal"
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        if kind == "software-upgrade":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    method,
                    kind,
                    proposal["name"],
                    "-y",
                    from_=proposer,
                    # content
                    title=proposal.get("title"),
                    description=proposal.get("description"),
                    upgrade_height=proposal.get("upgrade-height"),
                    upgrade_time=proposal.get("upgrade-time"),
                    upgrade_info=proposal.get("upgrade-info"),
                    deposit=proposal.get("deposit"),
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        elif kind == "cancel-software-upgrade":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    method,
                    kind,
                    "-y",
                    from_=proposer,
                    # content
                    title=proposal.get("title"),
                    description=proposal.get("description"),
                    deposit=proposal.get("deposit"),
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        elif kind == "register-erc20":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    method,
                    kind,
                    proposal.get("erc20_address"),
                    "-y",
                    from_=proposer,
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        elif kind == "register-coin":
            return json.loads(
                self.raw(
                    "tx",
                    "gov",
                    method,
                    kind,
                    proposal.get("metadata"),
                    "-y",
                    from_=proposer,
                    # content
                    title=proposal.get("title"),
                    description=proposal.get("description"),
                    deposit=proposal.get("deposit"),
                    # basic
                    home=self.data_dir,
                    **kwargs,
                )
            )
        else:
            with tempfile.NamedTemporaryFile("w") as fp:
                json.dump(proposal, fp)
                fp.flush()
                return json.loads(
                    self.raw(
                        "tx",
                        "gov",
                        method,
                        kind,
                        fp.name,
                        "-y",
                        from_=proposer,
                        # basic
                        home=self.data_dir,
                        **kwargs,
                    )
                )

    def gov_vote(self, voter, proposal_id, option, **kwargs):
        kwargs.setdefault("gas_prices", DEFAULT_GAS_PRICE)
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "vote",
                proposal_id,
                option,
                "-y",
                from_=voter,
                home=self.data_dir,
                **kwargs,
            )
        )

    def gov_deposit(self, depositor, proposal_id, amount, denom=DEFAULT_DENOM):
        return json.loads(
            self.raw(
                "tx",
                "gov",
                "deposit",
                proposal_id,
                f"{amount}{denom}",
                "-y",
                from_=depositor,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
            )
        )

    def query_proposals(self, depositor=None, limit=None, status=None, voter=None):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "proposals",
                depositor=depositor,
                count_total=limit,
                status=status,
                voter=voter,
                output="json",
                node=self.node_rpc,
            )
        )

    def query_proposal(self, proposal_id):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "proposal",
                proposal_id,
                output="json",
                node=self.node_rpc,
            )
        )

    def query_tally(self, proposal_id):
        return json.loads(
            self.raw(
                "query",
                "gov",
                "tally",
                proposal_id,
                output="json",
                node=self.node_rpc,
            )
        )

    # ==========================
    #           IBC
    # ==========================

    def ibc_transfer(
        self,
        from_,
        to,
        amount,
        channel,  # src channel
        target_version,  # chain version number of target chain
        i=0,
        fees="0aevmos",
    ):
        return json.loads(
            self.raw(
                "tx",
                "ibc-transfer",
                "transfer",
                "transfer",  # src port
                channel,
                to,
                amount,
                "-y",
                # FIXME https://github.com/cosmos/cosmos-sdk/issues/8059
                "--absolute-timeouts",
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                chain_id=self.chain_id,
                packet_timeout_height=f"{target_version}-10000000000",
                packet_timeout_timestamp=0,
                fees=fees,
            )
        )

    def register_counterparty_payee(
        self, port_id, channel_id, relayer, counterparty_payee, **kwargs
    ):
        default_kwargs = {
            "home": self.data_dir,
        }
        return json.loads(
            self.raw(
                "tx",
                "ibc-fee",
                "register-counterparty-payee",
                port_id,
                channel_id,
                relayer,
                counterparty_payee,
                "-y",
                **(default_kwargs | kwargs),
            )
        )

    def register_payee(self, port_id, channel_id, relayer, payee, **kwargs):
        default_kwargs = {
            "home": self.data_dir,
        }
        return json.loads(
            self.raw(
                "tx",
                "ibc-fee",
                "register-payee",
                port_id,
                channel_id,
                relayer,
                payee,
                "-y",
                **(default_kwargs | kwargs),
            )
        )

    def pay_packet_fee(self, port_id, channel_id, packet_seq, **kwargs):
        default_kwargs = {
            "home": self.data_dir,
        }
        return json.loads(
            self.raw(
                "tx",
                "ibc-fee",
                "pay-packet-fee",
                port_id,
                channel_id,
                str(packet_seq),
                "-y",
                **(default_kwargs | kwargs),
            )
        )

    # ==========================
    #        EVM Module
    # ==========================

    def build_evm_tx(self, raw_tx: str, **kwargs):
        return json.loads(
            self.raw(
                "tx",
                "evm",
                "raw",
                raw_tx,
                "-y",
                "--generate-only",
                home=self.data_dir,
                **kwargs,
            )
        )

    def evm_params(self, **kwargs):
        default_kwargs = {
            "node": self.node_rpc,
            "output": "json",
        }
        return json.loads(
            self.raw(
                "q",
                "evm",
                "params",
                **(default_kwargs | kwargs),
            )
        )

    # ==========================
    #       FEEMARKET Module
    # ==========================

    def query_base_fee(self, **kwargs):
        default_kwargs = {"home": self.data_dir}
        return int(
            json.loads(
                self.raw(
                    "q",
                    "feemarket",
                    "base-fee",
                    **(default_kwargs | kwargs),
                )
            )["base_fee"]
        )

    # ==========================
    #        AUTHZ Module
    # ==========================

    def authz_exec(self, tx_json_file: str, grantee: str, **kwargs):
        return json.loads(
            self.raw(
                "tx",
                "authz",
                "exec",
                tx_json_file,
                "--generate-only",
                "--from",
                grantee,
                home=self.data_dir,
                **kwargs,
            )
        )

    # ==========================
    #        AUTH Module
    # ==========================

    def account(self, addr):
        return json.loads(
            self.raw(
                "query", "auth", "account", addr, output="json", node=self.node_rpc
            )
        )

    def query_module_accounts(self, **kwargs):
        default_kwargs = {"output": "json", "home": self.data_dir}
        return json.loads(
            self.raw(
                "q",
                "auth",
                "module-accounts",
                **(default_kwargs | kwargs),
            )
        )["accounts"]

    # ==========================
    #       VESTING Module
    # ==========================

    def vesting_balance(self, addr: str):
        balances = self.raw("q", "vesting", "balances", addr, home=self.data_dir)
        # the '--output json' flag has no effect in this query
        # so need to parse it
        lines = balances.decode("utf-8").split("\n")
        res = {}

        for line in lines:
            if ":" in line:
                key, value = line.split(":", 1)
                res[key.strip().lower()] = value.strip()

        return res

    def create_vesting_acc(self, funder: str, address: str, gov_clawback="0", **kwargs):
        kwargs.setdefault(
            "gas_prices", f"{self.query_base_fee() + 100000}{DEFAULT_DENOM}"
        )
        return json.loads(
            self.raw(
                "tx",
                "vesting",
                "create-clawback-vesting-account",
                funder,
                gov_clawback,
                "--from",
                address,
                "--generate-only",
                home=self.data_dir,
                **kwargs,
            )
        )

    def fund_vesting_acc(
        self, address: str, funder: str, lockup_file: str, vesting_file: str, **kwargs
    ):
        kwargs.setdefault(
            "gas_prices", f"{self.query_base_fee() + 100000}{DEFAULT_DENOM}"
        )
        return json.loads(
            self.raw(
                "tx",
                "vesting",
                "fund-vesting-account",
                address,
                "--generate-only",
                "--from",
                funder,
                "--lockup",
                lockup_file,
                "--vesting",
                vesting_file,
                home=self.data_dir,
                **kwargs,
            )
        )

    # ==========================
    #       ERC20 Module
    # ==========================
    def convert_coin(self, coin: str, account: str, **kwargs):
        kwargs.setdefault(
            "gas_prices", f"{self.query_base_fee() + 100000}{DEFAULT_DENOM}"
        )
        return json.loads(
            self.raw(
                "tx",
                "erc20",
                "convert-coin",
                coin,
                "-y",
                from_=account,
                home=self.data_dir,
                **kwargs,
            )
        )

    def get_token_pairs(self, **kwargs):
        default_kwargs = {"output": "json", "home": self.data_dir}
        res = json.loads(
            self.raw(
                "q",
                "erc20",
                "token-pairs",
                **(default_kwargs | kwargs),
            )
        )
        return res["token_pairs"]

    # ==========================
    #        Tendermint
    # ==========================

    def consensus_address(self):
        "get tendermint consensus address"
        output = self.raw("tendermint", "show-address", home=self.data_dir)
        return output.decode().strip()

    def node_id(self):
        "get tendermint node id"
        output = self.raw("tendermint", "show-node-id", home=self.data_dir)
        return output.decode().strip()

    def export(self):
        return self.raw("export", home=self.data_dir)

    def unsaferesetall(self):
        return self.raw("unsafe-reset-all")

    #   TODO: create different classes for each chains CLI
    # ==========================
    #       Stride specific
    # ==========================
    def register_host_zone_msg(
        self,
        sender_addr,
        connection_id,
        host_denom,
        bech32_prefix,
        ibc_denom,
        channel_id,
        unbonding_frequency,
        lsm_enabled=0,
        **kwargs,
    ):
        return json.loads(
            self.raw(
                "tx",
                "stakeibc",
                "register-host-zone",
                connection_id,
                host_denom,
                bech32_prefix,
                ibc_denom,
                channel_id,
                unbonding_frequency,
                str(lsm_enabled),
                "-y",
                from_=sender_addr,
                chain_id="stride-1",
                home=self.data_dir,
                node=self.node_rpc,
                keyring_backend="test",
                **kwargs,
            )
        )

    def get_host_zones(self, **kwargs):
        """
        Queries the host zones on the Stride chain.
        """
        default_kwargs = {"output": "json", "home": self.data_dir}
        res = json.loads(
            self.raw(
                "q",
                "stakeibc",
                "list-host-zone",
                **(default_kwargs | kwargs),
            )
        )
        return res["host_zone"]

    #   TODO: create different classes for each chains CLI
    # ==========================
    #       Osmosis specific
    # ==========================
    def wasm_store_binary(
        self,
        from_,
        contract_path,
        **kwargs,
    ):
        """
        Store wasm binary contract.
        """
        return json.loads(
            self.raw(
                "tx",
                "wasm",
                "store",
                contract_path,
                "-y",
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                gas_adjustment=1.3,
                gas=4000000,
                gas_prices="0.25uosmo",
                keyring_backend="test",
                chain_id=self.chain_id,
                **kwargs,
            )
        )

    def wasm_instante2(
        self,
        from_,
        contract_code,
        init_args,
        label,
        **kwargs,
    ):
        """
        Store instantiate wasm contract with reproducible address.
        """
        # This could be any constant number.
        # Its only meant to guarantee determinism.
        salt = 74657374
        return json.loads(
            self.raw(
                "tx",
                "wasm",
                "instantiate2",
                contract_code,
                init_args,
                salt,
                "--label",
                label,
                "--no-admin",
                "-y",
                "--fix-msg",
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                gas_adjustment=1.3,
                gas=2000000,
                gas_prices="0.25uosmo",
                keyring_backend="test",
                chain_id=self.chain_id,
                **kwargs,
            )
        )

    def wasm_execute(
        self,
        from_,
        contract_address,
        execute_args,
        **kwargs,
    ):
        """
        Execute a wasm contract.
        """
        # This could be any constant number.
        return json.loads(
            self.raw(
                "tx",
                "wasm",
                "execute",
                contract_address,
                execute_args,
                "-y",
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                gas_adjustment=1.3,
                gas=2000000,
                gas_prices="0.25uosmo",
                keyring_backend="test",
                chain_id=self.chain_id,
                **kwargs,
            )
        )

    def get_wasm_contract_by_code(self, code, **kwargs):
        """
        Queries all wasm instances associated with a code ID.
        """
        default_kwargs = {"output": "json", "home": self.data_dir}
        res = json.loads(
            self.raw(
                "q",
                "wasm",
                "list-contract-by-code",
                code,
                **(default_kwargs | kwargs),
            )
        )
        return res["contracts"]

    def get_wasm_contract_state(self, contract_addr, query_args, **kwargs):
        """
        Queries the wasm contract state.
        """
        default_kwargs = {"output": "json", "home": self.data_dir}
        res = json.loads(
            self.raw(
                "q",
                "wasm",
                "contract-state",
                "smart",
                contract_addr,
                query_args,
                **(default_kwargs | kwargs),
            )
        )
        return res["contracts"]

    def gamm_create_pool(
        self,
        from_,
        pool_file_path,
        **kwargs,
    ):
        """
        Create Osmosis pools in gamm.
        """
        return json.loads(
            self.raw(
                "tx",
                "gamm",
                "create-pool",
                "-y",
                pool_file=pool_file_path,
                from_=from_,
                home=self.data_dir,
                node=self.node_rpc,
                gas_adjustment=1.3,
                gas=2000000,
                gas_prices="0.25uosmo",
                keyring_backend="test",
                chain_id=self.chain_id,
                **kwargs,
            )
        )
