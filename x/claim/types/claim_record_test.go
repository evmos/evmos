package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestClaimRecordValidate(t *testing.T) {
	testCases := []struct {
		name        string
		claimRecord ClaimRecord
		expError    bool
	}{
		{
			"fail - empty",
			ClaimRecord{},
			true,
		},
		{
			"fail - non positive claimable amount",
			ClaimRecord{InitialClaimableAmount: sdk.NewInt(-1)},
			true,
		},
		{
			"fail - empty actions",
			ClaimRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			ClaimRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{true, true, true, true},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claimRecord.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestClaimRecordHasClaimedAction(t *testing.T) {
	testCases := []struct {
		name        string
		claimRecord ClaimRecord
		action      Action
		expBool     bool
	}{
		{
			"false - empty",
			ClaimRecord{},
			ActionEVM,
			false,
		},
		{
			"false - unspecified action",
			ClaimRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			ActionUnspecified,
			false,
		},
		{
			"false - invalid action",
			ClaimRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			Action(10),
			false,
		},
		{
			"false - not claimed",
			ClaimRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			ActionEVM,
			false,
		},
		{
			"true - claimed",
			ClaimRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			ActionEVM,
			true,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expBool, tc.claimRecord.HasClaimedAction(tc.action), tc.name)
	}
}

func TestClaimRecordHasClaimedAll(t *testing.T) {
	testCases := []struct {
		name        string
		claimRecord ClaimRecord
		expBool     bool
	}{
		{
			"false - empty",
			ClaimRecord{},
			false,
		},
		{
			"false - not claimed",
			ClaimRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			false,
		},
		{
			"true - all claimed",
			ClaimRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			true,
		},
	}

	for _, tc := range testCases {
		require.True(t, tc.expBool == tc.claimRecord.HasClaimedAll(), tc.name)
	}
}

func TestClaimRecordAddressValidate(t *testing.T) {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name        string
		claimRecord ClaimRecordAddress
		expError    bool
	}{
		{
			"fail - empty",
			ClaimRecordAddress{},
			true,
		},
		{
			"fail - invalid address",
			NewClaimRecordAddress(sdk.AccAddress{}, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty int",
			NewClaimRecordAddress(addr, sdk.Int{}),
			true,
		},
		{
			"fail - non positive claimable amount",
			NewClaimRecordAddress(addr, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty actions",
			ClaimRecordAddress{
				Address:                addr.String(),
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			NewClaimRecordAddress(addr, sdk.OneInt()),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claimRecord.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
