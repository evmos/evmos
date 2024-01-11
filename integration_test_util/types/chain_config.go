package types

//goland:noinspection SpellCheckingInspection
import (
	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"math/big"
)

type ChainConfig struct {
	CosmosChainId    string
	BaseDenom        string
	Bech32Prefix     string
	EvmChainId       int64
	EvmChainIdBigInt *big.Int // dynamic: calculated from EvmChainId
}

type TestConfig struct {
	SecondaryDenomUnits []banktypes.DenomUnit
	InitBalanceAmount   sdkmath.Int
	DefaultFeeAmount    sdkmath.Int
	DisableTendermint   bool
}

type ChainConstantConfig struct {
	cosmosChainId string
	minDenom      string
	baseExponent  int
}

func NewChainConstantConfig(cosmosChainId, minDenom string, baseExponent int) ChainConstantConfig {
	return ChainConstantConfig{
		cosmosChainId: cosmosChainId,
		minDenom:      minDenom,
		baseExponent:  baseExponent,
	}
}

func (c ChainConstantConfig) GetCosmosChainID() string {
	return c.cosmosChainId
}

func (c ChainConstantConfig) GetMinDenom() string {
	return c.minDenom
}

func (c ChainConstantConfig) GetBaseExponent() int {
	return c.baseExponent
}
