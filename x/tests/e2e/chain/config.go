package chain

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/evmos/ethermint/server/config"
	"github.com/spf13/viper"
	tmconfig "github.com/tendermint/tendermint/config"
	tmjson "github.com/tendermint/tendermint/libs/json"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/evmos/v9/tests/e2e/util"
	claimstypes "github.com/evmos/evmos/v9/x/claims/types"
	inflationtypes "github.com/evmos/evmos/v9/x/inflation/types"
)

type ValidatorConfig struct {
	Pruning            string // default, nothing, everything, or custom
	PruningKeepRecent  string // keep all of the last N states (only used with custom pruning)
	PruningInterval    string // delete old states from every Nth block (only used with custom pruning)
	SnapshotInterval   uint64 // statesync snapshot every Nth block (0 to disable)
	SnapshotKeepRecent uint32 // number of recent snapshots to keep and serve (0 to keep all)
}

const (
	// common
	CoinDenom     = "ucoin"
	StakeDenom    = "aevmos"
	MinGasPrice   = "0.000"
	IbcSendAmount = 3300000000
	VotingPeriod  = 30000000000 // 30 seconds
	// chainA
	ChainAID      = "evmos_9001-2"
	CoinBalanceA  = "2000000000000000000000000000"
	StakeBalanceA = "1100000000000000000000000000"
	StakeAmountA  = "1000000000000000000000000000"

	// Currently only running one chain, so this is not used
	// chainB
	ChainBID      = "evmos_9000-2"
	CoinBalanceB  = "500000000000000000000000"
	StakeBalanceB = "440000000000000000000000"
	StakeAmountB  = "400000000000000000000000"
)

var (
	StakeAmountIntA, _ = sdk.NewIntFromString(StakeAmountA)
	StakeAmountCoinA   = sdk.NewCoin(StakeDenom, StakeAmountIntA)
	StakeAmountIntB, _ = sdk.NewIntFromString(StakeAmountB)
	StakeAmountCoinB   = sdk.NewCoin(StakeDenom, StakeAmountIntB)

	InitBalanceStrA = fmt.Sprintf("%s%s,%s%s", CoinBalanceA, CoinDenom, StakeBalanceA, StakeDenom)
	InitBalanceStrB = fmt.Sprintf("%s%s,%s%s", CoinBalanceB, CoinDenom, StakeBalanceB, StakeDenom)
)

func addAccount(path, moniker, amountStr string, accAddr sdk.AccAddress) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(path)
	config.Moniker = moniker

	coins, err := sdk.ParseCoinsNormalized(amountStr)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	balances := banktypes.Balance{Address: accAddr.String(), Coins: coins.Sort()}
	genAccount := authtypes.NewBaseAccount(accAddr, nil, 0, 0)

	genFile := config.GenesisFile()
	appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	authGenState := authtypes.GetGenesisStateFromAppState(util.Cdc, appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return fmt.Errorf("failed to get accounts from any: %w", err)
	}

	if accs.Contains(accAddr) {
		return fmt.Errorf("failed to add account to genesis state; account already exists: %s", accAddr)
	}

	// Add the new account to the set of genesis accounts and sanitize the
	// accounts afterwards.
	accs = append(accs, genAccount)
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		return fmt.Errorf("failed to convert accounts into any's: %w", err)
	}

	authGenState.Accounts = genAccs

	authGenStateBz, err := util.Cdc.MarshalJSON(&authGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	appState[authtypes.ModuleName] = authGenStateBz

	bankGenState := banktypes.GetGenesisStateFromAppState(util.Cdc, appState)
	bankGenState.Balances = append(bankGenState.Balances, balances)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	bankGenStateBz, err := util.Cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}

	appState[banktypes.ModuleName] = bankGenStateBz

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal application genesis state: %w", err)
	}

	genDoc.AppState = appStateJSON
	return genutil.ExportGenesisFile(genDoc, genFile)
}

func updateBankModule(appGenState map[string]json.RawMessage) error {
	var bankGenState banktypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState); err != nil {
		return err
	}

	bankGenState.DenomMetadata = append(bankGenState.DenomMetadata, banktypes.Metadata{
		Description: "An example stable token",
		Display:     CoinDenom,
		Base:        CoinDenom,
		Symbol:      CoinDenom,
		Name:        CoinDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    CoinDenom,
				Exponent: 0,
			},
		},
	})

	bz, err := util.Cdc.MarshalJSON(&bankGenState)
	if err != nil {
		return err
	}
	appGenState[banktypes.ModuleName] = bz
	return nil
}

func updateGovModule(appGenState map[string]json.RawMessage) error {
	var govGenState govtypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState); err != nil {
		return err
	}

	govGenState.VotingParams = govtypes.VotingParams{
		VotingPeriod: VotingPeriod,
	}

	govGenState.DepositParams.MinDeposit = sdk.NewCoins(sdk.NewCoin("aevmos", sdk.ZeroInt()))

	gz, err := util.Cdc.MarshalJSON(&govGenState)
	if err != nil {
		return err
	}
	appGenState[govtypes.ModuleName] = gz
	return nil
}

func updateStakingModule(appGenState map[string]json.RawMessage) error {
	var stakingGenState stakingtypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakingGenState); err != nil {
		return err
	}

	stakingGenState.Params.BondDenom = StakeDenom

	bz, err := util.Cdc.MarshalJSON(&stakingGenState)
	if err != nil {
		return err
	}
	appGenState[stakingtypes.ModuleName] = bz
	return nil
}

func updateCrisisModule(appGenState map[string]json.RawMessage) error {
	var crisisGenState crisistypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[crisistypes.ModuleName], &crisisGenState); err != nil {
		return err
	}

	crisisGenState.ConstantFee.Denom = StakeDenom

	bz, err := util.Cdc.MarshalJSON(&crisisGenState)
	if err != nil {
		return err
	}
	appGenState[crisistypes.ModuleName] = bz
	return nil
}

func updateEvmModule(appGenState map[string]json.RawMessage) error {
	var evmGenState evmtypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[evmtypes.ModuleName], &evmGenState); err != nil {
		return err
	}

	evmGenState.Params.EvmDenom = StakeDenom

	bz, err := util.Cdc.MarshalJSON(&evmGenState)
	if err != nil {
		return err
	}
	appGenState[evmtypes.ModuleName] = bz
	return nil
}

func updateInflationModule(appGenState map[string]json.RawMessage) error {
	var inflationGenState inflationtypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[inflationtypes.ModuleName], &inflationGenState); err != nil {
		return err
	}

	inflationGenState.Params.MintDenom = StakeDenom

	bz, err := util.Cdc.MarshalJSON(&inflationGenState)
	if err != nil {
		return err
	}
	appGenState[inflationtypes.ModuleName] = bz
	return nil
}

func updateGenTxs(appGenState map[string]json.RawMessage, c *internalChain) error {
	var genUtilGenState genutiltypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[genutiltypes.ModuleName], &genUtilGenState); err != nil {
		return err
	}

	// generate genesis txs
	genTxs := make([]json.RawMessage, len(c.validators))
	for i, val := range c.validators {
		stakeAmountCoin := StakeAmountCoinA
		if c.chainMeta.ID != ChainAID {
			stakeAmountCoin = StakeAmountCoinB
		}
		createValmsg, err := val.buildCreateValidatorMsg(stakeAmountCoin)
		if err != nil {
			return err
		}

		signedTx, err := val.signMsg(createValmsg)
		if err != nil {
			return err
		}

		txRaw, err := util.Cdc.MarshalJSON(signedTx)
		if err != nil {
			return err
		}

		genTxs[i] = txRaw
	}

	genUtilGenState.GenTxs = genTxs

	bz, err := util.Cdc.MarshalJSON(&genUtilGenState)
	if err != nil {
		return err
	}
	appGenState[genutiltypes.ModuleName] = bz
	return nil
}

func updateClaimsModule(appGenState map[string]json.RawMessage) error {
	var claimsGenState claimstypes.GenesisState
	if err := util.Cdc.UnmarshalJSON(appGenState[claimstypes.ModuleName], &claimsGenState); err != nil {
		return err
	}

	claimsGenState.ClaimsRecords = append(claimsGenState.ClaimsRecords,
		claimstypes.ClaimsRecordAddress{
			Address:                "evmos13cf9npvns2vhh3097909mkhfxngmw6d6eppfm4",
			InitialClaimableAmount: sdk.NewInt(0),
			ActionsCompleted:       []bool{false, false, false, true},
		})
	claimsGenState.ClaimsRecords = append(claimsGenState.ClaimsRecords,
		claimstypes.ClaimsRecordAddress{
			Address:                "evmos17xpfvakm2amg962yls6f84z3kell8c5ljcjw34",
			InitialClaimableAmount: sdk.NewInt(0),
			ActionsCompleted:       []bool{true, true, false, true},
		})
	claimsGenState.ClaimsRecords = append(claimsGenState.ClaimsRecords,
		claimstypes.ClaimsRecordAddress{
			Address:                "evmos1x8eupnk7hhnnm5m824qt53203w0m6x7tkr5l9u",
			InitialClaimableAmount: sdk.NewInt(0),
			ActionsCompleted:       []bool{false, true, false, false},
		})

	bz, err := util.Cdc.MarshalJSON(&claimsGenState)
	if err != nil {
		return err
	}
	appGenState[claimstypes.ModuleName] = bz
	return nil
}

func initGenesis(c *internalChain) error {
	serverCtx := server.NewDefaultContext()
	config := serverCtx.Config

	config.SetRoot(c.validators[0].configDir())
	config.Moniker = c.validators[0].getMoniker()

	genFilePath := config.GenesisFile()
	appGenState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFilePath)
	if err != nil {
		return err
	}

	if err := updateBankModule(appGenState); err != nil {
		return err
	}
	if err := updateGovModule(appGenState); err != nil {
		return err
	}
	if err := updateGenTxs(appGenState, c); err != nil {
		return err
	}
	if err := updateClaimsModule(appGenState); err != nil {
		return err
	}
	if err := updateStakingModule(appGenState); err != nil {
		return err
	}
	if err := updateCrisisModule(appGenState); err != nil {
		return err
	}
	if err := updateEvmModule(appGenState); err != nil {
		return err
	}
	if err := updateInflationModule(appGenState); err != nil {
		return err
	}

	bz, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc.AppState = bz

	bz, err = tmjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}

	// write the updated genesis file to each validator
	for _, val := range c.validators {
		if err := util.WriteFile(filepath.Join(val.configDir(), "config", "genesis.json"), bz); err != nil {
			return err
		}
	}
	return nil
}

func initNodes(c *internalChain, numVal int) error {
	if err := c.createValidators(numVal); err != nil {
		return err
	}

	// initialize a genesis file for the first validator
	val0ConfigDir := c.validators[0].configDir()
	for _, val := range c.validators {
		if c.chainMeta.ID == ChainAID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrA, val.getKeyInfo().GetAddress()); err != nil {
				return err
			}
		} else if c.chainMeta.ID == ChainBID {
			if err := addAccount(val0ConfigDir, "", InitBalanceStrB, val.getKeyInfo().GetAddress()); err != nil {
				return err
			}
		}
	}

	// copy the genesis file to the remaining validators
	for _, val := range c.validators[1:] {
		_, err := util.CopyFile(
			filepath.Join(val0ConfigDir, "config", "genesis.json"),
			filepath.Join(val.configDir(), "config", "genesis.json"),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func initValidatorConfigs(c *internalChain, validatorConfigs []*ValidatorConfig) error {
	for i, val := range c.validators {
		tmCfgPath := filepath.Join(val.configDir(), "config", "config.toml")

		vpr := viper.New()
		vpr.SetConfigFile(tmCfgPath)
		if err := vpr.ReadInConfig(); err != nil {
			return err
		}

		valConfig := &tmconfig.Config{}
		if err := vpr.Unmarshal(valConfig); err != nil {
			return err
		}

		valConfig.P2P.ListenAddress = "tcp://0.0.0.0:26656"
		valConfig.P2P.AddrBookStrict = false
		valConfig.P2P.ExternalAddress = fmt.Sprintf("%s:%d", val.instanceName(), 26656)
		valConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"
		valConfig.StateSync.Enable = false
		valConfig.LogLevel = "info"

		valConfig.Storage = &tmconfig.StorageConfig{DiscardABCIResponses: false}

		var peers []string

		for j := 0; j < len(c.validators); j++ {
			if i == j {
				continue
			}

			peer := c.validators[j]
			peerID := fmt.Sprintf("%s@%s%d:26656", peer.getNodeKey().ID(), peer.getMoniker(), j)
			peers = append(peers, peerID)
		}

		valConfig.P2P.PersistentPeers = strings.Join(peers, ",")

		tmconfig.WriteConfigFile(tmCfgPath, valConfig)

		// set application configuration
		appCfgPath := filepath.Join(val.configDir(), "config", "app.toml")
		customAppTemplate, _ := config.AppConfig("aevmos")
		srvconfig.SetConfigTemplate(customAppTemplate)

		appConfig := config.DefaultConfig()
		appConfig.BaseConfig.Pruning = validatorConfigs[i].Pruning
		appConfig.BaseConfig.PruningKeepRecent = validatorConfigs[i].PruningKeepRecent
		appConfig.BaseConfig.PruningInterval = validatorConfigs[i].PruningInterval
		appConfig.API.Enable = true
		appConfig.MinGasPrices = fmt.Sprintf("%s%s", MinGasPrice, CoinDenom)
		appConfig.StateSync.SnapshotInterval = validatorConfigs[i].SnapshotInterval
		appConfig.StateSync.SnapshotKeepRecent = validatorConfigs[i].SnapshotKeepRecent

		srvconfig.WriteConfigFile(appCfgPath, appConfig)
	}
	return nil
}
