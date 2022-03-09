package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestClaimsRecordValidate(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord ClaimsRecord
		expError     bool
	}{
		{
			"fail - empty",
			ClaimsRecord{},
			true,
		},
		{
			"fail - non positive claimable amount",
			ClaimsRecord{InitialClaimableAmount: sdk.NewInt(-1)},
			true,
		},
		{
			"fail - empty actions",
			ClaimsRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			ClaimsRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{true, true, true, true},
			},
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claimsRecord.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestClaimsRecordHasClaimedAction(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord ClaimsRecord
		action       Action
		expBool      bool
	}{
		{
			"false - empty",
			ClaimsRecord{},
			ActionEVM,
			false,
		},
		{
			"false - unspecified action",
			ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			ActionUnspecified,
			false,
		},
		{
			"false - invalid action",
			ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			Action(10),
			false,
		},
		{
			"false - not claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			ActionEVM,
			false,
		},
		{
			"true - claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			ActionEVM,
			true,
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.expBool, tc.claimsRecord.HasClaimedAction(tc.action), tc.name)
	}
}

func TestClaimsRecordHasClaimedAll(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord ClaimsRecord
		expBool      bool
	}{
		{
			"false - empty",
			ClaimsRecord{},
			false,
		},
		{
			"false - not claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			false,
		},
		{
			"true - all claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			true,
		},
	}

	for _, tc := range testCases {
		require.True(t, tc.expBool == tc.claimsRecord.HasClaimedAll(), tc.name)
	}
}

func TestClaimsRecordHasAny(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord ClaimsRecord
		expBool      bool
	}{
		{
			"false - empty",
			ClaimsRecord{},
			false,
		},
		{
			"false - not claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			false,
		},
		{
			"true - single action claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{true, false, false, false},
			},
			true,
		},
		{
			"true - all claimed",
			ClaimsRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			true,
		},
	}

	for _, tc := range testCases {
		require.True(t, tc.expBool == tc.claimsRecord.HasClaimedAny(), tc.name)
	}
}

func TestClaimsRecordAddressValidate(t *testing.T) {
	addr := sdk.AccAddress(tests.GenerateAddress().Bytes())

	testCases := []struct {
		name         string
		claimsRecord ClaimsRecordAddress
		expError     bool
	}{
		{
			"fail - empty",
			ClaimsRecordAddress{},
			true,
		},
		{
			"fail - invalid address",
			NewClaimsRecordAddress(sdk.AccAddress{}, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty int",
			NewClaimsRecordAddress(addr, sdk.Int{}),
			true,
		},
		{
			"fail - non positive claimable amount",
			NewClaimsRecordAddress(addr, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty actions",
			ClaimsRecordAddress{
				Address:                addr.String(),
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			NewClaimsRecordAddress(addr, sdk.OneInt()),
			false,
		},
	}

	for _, tc := range testCases {
		err := tc.claimsRecord.Validate()
		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}
