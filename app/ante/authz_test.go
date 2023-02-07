package ante_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"

	"github.com/evmos/evmos/v11/app/ante"
)

func TestAuthzLimiterDecorator(t *testing.T) {
	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(5)
	require.NoError(t, err)

	distantFuture := time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)

	validator := sdk.ValAddress(testAddresses[4])
	stakingAuthDelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE, nil)
	require.NoError(t, err)
	stakingAuthUndelegate, err := stakingtypes.NewStakeAuthorization([]sdk.ValAddress{validator}, nil, stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE, nil)
	require.NoError(t, err)

	decorator := ante.NewAuthzLimiterDecorator(
		sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
		sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	)

	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		checkTx     bool
		expectedErr error
	}{
		{
			name: "a non blocked msg is not blocked",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(
					testAddresses[0],
					testAddresses[1],
					sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
				),
			},
			checkTx: false,
		},
		{
			name: "a blocked msg is not blocked when not wrapped in MsgExec",
			msgs: []sdk.Msg{
				&evmtypes.MsgEthereumTx{},
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					authz.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
					&distantFuture,
				),
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthDelegate,
					&distantFuture,
				),
			},
			checkTx: false,
		},
		{
			name: "when a MsgGrant contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					authz.NewGenericAuthorization(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
					&distantFuture,
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "when a MsgGrant contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthUndelegate,
					&distantFuture,
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "when a MsgExec contains a non blocked msg, it passes",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{banktypes.NewMsgSend(
						testAddresses[0],
						testAddresses[3],
						sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
					)}),
			},
			checkTx: false,
		},
		{
			name: "when a MsgExec contains a blocked msg, it is blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "blocked msg surrounded by valid msgs is still blocked",
			msgs: []sdk.Msg{
				newMsgGrant(
					testAddresses[0],
					testAddresses[1],
					stakingAuthDelegate,
					&distantFuture,
				),
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						banktypes.NewMsgSend(
							testAddresses[0],
							testAddresses[3],
							sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
						),
						&evmtypes.MsgEthereumTx{},
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "a nested MsgExec containing a blocked msg is still blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						newMsgExec(
							testAddresses[2],
							[]sdk.Msg{
								&evmtypes.MsgEthereumTx{},
							},
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "a nested MsgGrant containing a blocked msg is still blocked",
			msgs: []sdk.Msg{
				newMsgExec(
					testAddresses[1],
					[]sdk.Msg{
						newMsgGrant(
							testAddresses[0],
							testAddresses[1],
							authz.NewGenericAuthorization(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
							&distantFuture,
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "a nested MsgExec NOT containing a blocked msg but has more nesting levels than the allowed is still blocked",
			msgs: []sdk.Msg{
				createNestedMsgExec(
					testAddresses[1],
					6,
					[]sdk.Msg{
						banktypes.NewMsgSend(
							testAddresses[0],
							testAddresses[3],
							sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
		{
			name: "two nested MsgExec messages NOT containing a blocked msg but between the two have more nesting than the allowed, then is still blocked",
			msgs: []sdk.Msg{
				createNestedMsgExec(
					testAddresses[1],
					3,
					[]sdk.Msg{
						banktypes.NewMsgSend(
							testAddresses[0],
							testAddresses[3],
							sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
						),
					},
				),
				createNestedMsgExec(
					testAddresses[1],
					4,
					[]sdk.Msg{
						banktypes.NewMsgSend(
							testAddresses[0],
							testAddresses[3],
							sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
						),
					},
				),
			},
			checkTx:     false,
			expectedErr: sdkerrors.ErrUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			ctx := sdk.Context{}.WithIsCheckTx(tc.checkTx)
			tx, err := createTx(testPrivKeys[0], tc.msgs...)
			require.NoError(t, err)
			mmd := MockAnteHandler{}
			_, err = decorator.AnteHandle(ctx, tx, false, mmd.AnteHandle)
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
