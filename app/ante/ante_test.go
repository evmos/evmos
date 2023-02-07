package ante_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *AnteTestSuite) TestRejectMsgsInAuthz() {
	testPrivKeys, testAddresses, err := generatePrivKeyAddressPairs(10)
	suite.Require().NoError(err)

	distantFuture := time.Date(9000, 1, 1, 0, 0, 0, 0, time.UTC)
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)

	newMsgGrant := func(msgTypeUrl string) *authz.MsgGrant {
		msg, err := authz.NewMsgGrant(
			testAddresses[0],
			testAddresses[1],
			authz.NewGenericAuthorization(msgTypeUrl),
			&distantFuture,
		)
		if err != nil {
			panic(err)
		}
		return msg
	}

	testcases := []struct {
		name         string
		msg          sdk.Msg
		expectedCode uint32
		isEIP712     bool
	}{
		{
			name:         "a MsgGrant with MsgEthereumTx typeURL on the authorization field is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name:         "a MsgGrant with MsgCreateVestingAccount typeURL on the authorization field is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name:         "a MsgGrant with MsgEthereumTx typeURL on the authorization field included on EIP712 tx is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
			isEIP712:     true,
		},
		{
			name: "a MsgExec with nested messages (valid: MsgSend and invalid: MsgEthereumTx) is blocked",
			msg: newMsgExec(
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
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgExec with nested MsgExec messages that has invalid messages is blocked",
			msg: newMsgExec(
				testAddresses[1],
				[]sdk.Msg{
					newMsgExec(
						testAddresses[1],
						[]sdk.Msg{
							newMsgExec(
								testAddresses[1],
								[]sdk.Msg{
									&evmtypes.MsgEthereumTx{},
								},
							),
						},
					),
				},
			),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name: "a MsgExec with more nested MsgExec messages than allowed and with valid messages is blocked",
			msg: newMsgExec(
				testAddresses[1],
				[]sdk.Msg{
					newMsgExec(
						testAddresses[2],
						[]sdk.Msg{
							newMsgExec(
								testAddresses[3],
								[]sdk.Msg{
									newMsgExec(
										testAddresses[4],
										[]sdk.Msg{
											newMsgExec(
												testAddresses[3],
												[]sdk.Msg{
													newMsgExec(
														testAddresses[3],
														[]sdk.Msg{
															banktypes.NewMsgSend(
																testAddresses[0],
																testAddresses[3],
																sdk.NewCoins(sdk.NewInt64Coin(evmtypes.DefaultEVMDenom, 100e6)),
															),
														},
													),
												},
											),
										},
									),
								},
							),
						},
					),
				},
			),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
	}

	for _, tc := range testcases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(false)
			var (
				tx  sdk.Tx
				err error
			)

			if tc.isEIP712 {
				tx, err = createEIP712CosmosTx(testAddresses[0], testPrivKeys[0], []sdk.Msg{tc.msg})
			} else {
				tx, err = createTx(testPrivKeys[0], tc.msg)
			}
			suite.Require().NoError(err)

			txEncoder := encodingConfig.TxConfig.TxEncoder()
			bz, err := txEncoder(tx)
			suite.Require().NoError(err)

			resCheckTx := suite.app.CheckTx(
				abci.RequestCheckTx{
					Tx:   bz,
					Type: abci.CheckTxType_New,
				},
			)
			suite.Require().Equal(resCheckTx.Code, tc.expectedCode, resCheckTx.Log)

			resDeliverTx := suite.app.DeliverTx(
				abci.RequestDeliverTx{
					Tx: bz,
				},
			)
			suite.Require().Equal(resDeliverTx.Code, tc.expectedCode, resDeliverTx.Log)
		})
	}
}
