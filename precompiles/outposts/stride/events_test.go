package stride_test

import (
	"fmt"
	"math/big"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v15/precompiles/common"
	"github.com/evmos/evmos/v15/precompiles/ics20"
	"github.com/evmos/evmos/v15/precompiles/outposts/stride"
	"github.com/evmos/evmos/v15/utils"
)

//nolint:dupl
func (s *PrecompileTestSuite) TestLiquidStakeEvent() {
	method := s.precompile.Methods[stride.LiquidStakeMethod]

	receiver := "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0"
	denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, "aevmos")
	tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success",
			func() []interface{} {
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				return []interface{}{
					s.address,
					common.HexToAddress(tokenPair.Erc20Address),
					big.NewInt(1e18),
					receiver,
				}
			},
			func() {
				ics20Log := s.stateDB.Logs()[0]
				s.Require().Equal(ics20Log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				ics20Event := s.precompile.ABI.Events[ics20.EventTypeIBCTransfer]
				s.Require().Equal(ics20Event.ID, common.HexToHash(ics20Log.Topics[0].Hex()))
				s.Require().Equal(ics20Log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var ibcTransferEvent ics20.EventIBCTransfer
				err := cmn.UnpackLog(s.precompile.ABI, &ibcTransferEvent, ics20.EventTypeIBCTransfer, *ics20Log)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), ibcTransferEvent.Sender)
				s.Require().Equal(crypto.Keccak256Hash([]byte(receiver)), ibcTransferEvent.Receiver)
				s.Require().Equal("transfer", ibcTransferEvent.SourcePort)
				s.Require().Equal("channel-0", ibcTransferEvent.SourceChannel)
				s.Require().Equal(big.NewInt(1e18), ibcTransferEvent.Amount)
				s.Require().Equal(utils.BaseDenom, ibcTransferEvent.Denom)

				memo, err := stride.CreateMemo(stride.LiquidStakeAction, receiver)
				s.Require().NoError(err)
				s.Require().Equal(memo, ibcTransferEvent.Memo)

				liquidStakeLog := s.stateDB.Logs()[1]
				s.Require().Equal(liquidStakeLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeLiquidStake]
				s.Require().Equal(event.ID, common.HexToHash(liquidStakeLog.Topics[0].Hex()))
				s.Require().Equal(liquidStakeLog.BlockNumber, uint64(s.ctx.BlockHeight()))

				var liquidStakeEvent stride.EventLiquidStake
				err = cmn.UnpackLog(s.precompile.ABI, &liquidStakeEvent, stride.EventTypeLiquidStake, *liquidStakeLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), liquidStakeEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), liquidStakeEvent.Token)
				s.Require().Equal(big.NewInt(1e18), liquidStakeEvent.Amount)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			_, err := s.precompile.LiquidStake(s.ctx, s.address, s.stateDB, contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}

//nolint:dupl
func (s *PrecompileTestSuite) TestRedeemEvent() {
	method := s.precompile.Methods[stride.RedeemMethod]

	bondDenom := s.app.StakingKeeper.BondDenom(s.ctx)
	denomTrace := transfertypes.DenomTrace{
		Path:      fmt.Sprintf("%s/%s", "transfer", "channel-0"),
		BaseDenom: "st" + bondDenom,
	}

	stEvmos := denomTrace.IBCDenom()

	receiver := "stride1rhe5leyt5w0mcwd9rpp93zqn99yktsxvyaqgd0"
	denomID := s.app.Erc20Keeper.GetDenomMap(s.ctx, stEvmos)
	tokenPair, ok := s.app.Erc20Keeper.GetTokenPair(s.ctx, denomID)
	s.Require().True(ok, "expected token pair to be found")

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		{
			"success",
			func() []interface{} {
				path := NewTransferPath(s.chainA, s.chainB)
				s.coordinator.Setup(path)
				return []interface{}{
					s.address,
					common.HexToAddress(tokenPair.Erc20Address),
					big.NewInt(1e18),
					receiver,
				}
			},
			func() {
				ics20Log := s.stateDB.Logs()[0]
				s.Require().Equal(ics20Log.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				ics20Event := s.precompile.ABI.Events[ics20.EventTypeIBCTransfer]
				s.Require().Equal(ics20Event.ID, common.HexToHash(ics20Log.Topics[0].Hex()))
				s.Require().Equal(ics20Log.BlockNumber, uint64(s.ctx.BlockHeight()))

				var ibcTransferEvent ics20.EventIBCTransfer
				err := cmn.UnpackLog(s.precompile.ABI, &ibcTransferEvent, ics20.EventTypeIBCTransfer, *ics20Log)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), ibcTransferEvent.Sender)
				s.Require().Equal(crypto.Keccak256Hash([]byte(receiver)), ibcTransferEvent.Receiver)
				s.Require().Equal("transfer", ibcTransferEvent.SourcePort)
				s.Require().Equal("channel-0", ibcTransferEvent.SourceChannel)
				s.Require().Equal(big.NewInt(1e18), ibcTransferEvent.Amount)
				s.Require().Equal(utils.BaseDenom, ibcTransferEvent.Denom)

				memo, err := stride.CreateMemo(stride.RedeemAction, receiver)
				s.Require().NoError(err)
				s.Require().Equal(memo, ibcTransferEvent.Memo)

				liquidStakeLog := s.stateDB.Logs()[1]
				s.Require().Equal(liquidStakeLog.Address, s.precompile.Address())
				// Check event signature matches the one emitted
				event := s.precompile.ABI.Events[stride.EventTypeRedeem]
				s.Require().Equal(event.ID, common.HexToHash(liquidStakeLog.Topics[0].Hex()))
				s.Require().Equal(liquidStakeLog.BlockNumber, uint64(s.ctx.BlockHeight()))

				var liquidStakeEvent stride.EventLiquidStake
				err = cmn.UnpackLog(s.precompile.ABI, &liquidStakeEvent, stride.EventTypeRedeem, *liquidStakeLog)
				s.Require().NoError(err)
				s.Require().Equal(common.BytesToAddress(s.address.Bytes()), liquidStakeEvent.Sender)
				s.Require().Equal(common.HexToAddress(tokenPair.Erc20Address), liquidStakeEvent.Token)
				s.Require().Equal(big.NewInt(1e18), liquidStakeEvent.Amount)
			},
			200000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			contract := vm.NewContract(vm.AccountRef(s.address), s.precompile, big.NewInt(0), tc.gas)

			_, err := s.precompile.Redeem(s.ctx, s.address, s.stateDB, contract, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
