package ante_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
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
			name:         "MsgEthereumTx is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name:         "MsgCreateVestingAccount is blocked",
			msg:          newMsgGrant(sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
		{
			name:         "MsgEthereumTx is blocked on EIP712",
			msg:          newMsgGrant(sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
			isEIP712:     true,
		},
	}

	for _, tc := range testcases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest(false)
			var (
				tx sdk.Tx
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
