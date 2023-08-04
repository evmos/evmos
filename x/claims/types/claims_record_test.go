package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	utiltx "github.com/evmos/evmos/v14/testutil/tx"
	"github.com/evmos/evmos/v14/x/claims/types"
	"github.com/stretchr/testify/require"
)

func TestClaimsRecordValidate(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecord
		expError     bool
	}{
		{
			"fail - empty",
			types.ClaimsRecord{},
			true,
		},
		{
			"fail - non positive claimable amount",
			types.ClaimsRecord{InitialClaimableAmount: sdk.NewInt(-1)},
			true,
		},
		{
			"fail - empty actions",
			types.ClaimsRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			types.ClaimsRecord{
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{true, true, true, true},
			},
			false,
		},
		{
			"success - valid instance with constructor",
			types.NewClaimsRecord(sdk.OneInt()),
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

func TestClaimAction(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecord
		action       types.Action
		expClaimed   bool
	}{
		{
			"fail - empty",
			types.ClaimsRecord{},
			types.ActionEVM,
			false,
		},
		{
			"fail - unspecified action",
			types.NewClaimsRecord(sdk.OneInt()),
			types.ActionUnspecified,
			false,
		},
		{
			"fail - invalid action",
			types.NewClaimsRecord(sdk.OneInt()),
			types.Action(10),
			false,
		},
		{
			"success - valid instance with constructor",
			types.NewClaimsRecord(sdk.OneInt()),
			types.ActionEVM,
			true,
		},
	}

	for _, tc := range testCases {
		tc.claimsRecord.MarkClaimed(tc.action)
		require.Equal(t, tc.expClaimed, tc.claimsRecord.HasClaimedAction(tc.action))
	}
}

func TestClaimsRecordHasClaimedAction(t *testing.T) {
	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecord
		action       types.Action
		expBool      bool
	}{
		{
			"false - empty",
			types.ClaimsRecord{},
			types.ActionEVM,
			false,
		},
		{
			"false - unspecified action",
			types.ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			types.ActionUnspecified,
			false,
		},
		{
			"false - invalid action",
			types.ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			types.Action(10),
			false,
		},
		{
			"false - not claimed",
			types.ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			types.ActionEVM,
			false,
		},
		{
			"true - claimed",
			types.ClaimsRecord{
				ActionsCompleted: []bool{true, true, true, true},
			},
			types.ActionEVM,
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
		claimsRecord types.ClaimsRecord
		expBool      bool
	}{
		{
			"false - empty",
			types.ClaimsRecord{},
			false,
		},
		{
			"false - not claimed",
			types.ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			false,
		},
		{
			"true - all claimed",
			types.ClaimsRecord{
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
		claimsRecord types.ClaimsRecord
		expBool      bool
	}{
		{
			"false - empty",
			types.ClaimsRecord{},
			false,
		},
		{
			"false - not claimed",
			types.ClaimsRecord{
				ActionsCompleted: []bool{false, false, false, false},
			},
			false,
		},
		{
			"true - single action claimed",
			types.ClaimsRecord{
				ActionsCompleted: []bool{true, false, false, false},
			},
			true,
		},
		{
			"true - all claimed",
			types.ClaimsRecord{
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
	addr := sdk.AccAddress(utiltx.GenerateAddress().Bytes())

	testCases := []struct {
		name         string
		claimsRecord types.ClaimsRecordAddress
		expError     bool
	}{
		{
			"fail - empty",
			types.ClaimsRecordAddress{},
			true,
		},
		{
			"fail - invalid address",
			types.NewClaimsRecordAddress(sdk.AccAddress{}, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty int",
			types.NewClaimsRecordAddress(addr, math.Int{}),
			true,
		},
		{
			"fail - non positive claimable amount",
			types.NewClaimsRecordAddress(addr, sdk.NewInt(-1)),
			true,
		},
		{
			"fail - empty actions",
			types.ClaimsRecordAddress{
				Address:                addr.String(),
				InitialClaimableAmount: sdk.OneInt(),
				ActionsCompleted:       []bool{},
			},
			true,
		},
		{
			"success - valid instance",
			types.NewClaimsRecordAddress(addr, sdk.OneInt()),
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
