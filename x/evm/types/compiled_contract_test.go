package types_test

import (
	"encoding/json"
	"os"
	"testing"

	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestHardhatCompiledContract(t *testing.T) {
	contents, err := os.ReadFile("testdata/SimpleContractHardhat.json")
	require.NoError(t, err, "failed to read file")
	require.NotEmpty(t, contents, "expected contents not to be empty")

	var hardhatContract evmtypes.HardhatCompiledContract
	err = json.Unmarshal(contents, &hardhatContract)
	require.NoError(t, err, "failed to unmarshal contract")

	require.Equal(t, hardhatContract.ContractName, "SimpleContract")
	require.Contains(t,
		hardhatContract.ABI.Methods,
		"setValue",
		"missing setValue method in contract ABI methods",
	)

	compiledContract, err := hardhatContract.ToCompiledContract()
	require.NoError(t, err, "failed to convert hardhat contract to compiled contract type")
	require.Equal(t, compiledContract.ABI, hardhatContract.ABI, "expected ABIs to be equal")
	require.NotEmpty(t, compiledContract.Bin, "expected bin data not to be empty")
}
