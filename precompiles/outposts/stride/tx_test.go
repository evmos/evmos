package stride_test

import (
	"github.com/evmos/evmos/v14/precompiles/outposts/stride"
)

func (s *PrecompileTestSuite) TestLiquidStakeEvmos() {
	method := s.precompile.Methods[stride.LiquidStakeEvmosMethod]

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func()
		gas         uint64
		expError    bool
		errContains string
	}{
		//{
		//	"fail - empty input args",
		//	func() []interface{} {
		//		return []interface{}{}
		//	},
		//	func() {},
		//	200000,
		//	true,
		//	fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 2, 0),
		//},
		//{
		//	"fail - bond denom is not aevmos",
		//	func() []interface{} {
		//		return []interface{}{
		//			cmn.Coin{
		//				Denom:  "uosmos",
		//				Amount: big.NewInt(1000),
		//			},
		//			s.address.String(),
		//		}
		//	},
		//	func() {},
		//	200000,
		//	true,
		//	fmt.Sprintf(cmn.ErrInvalidDenom, "aevmos"),
		//},
		//{
		//	"fail - invalid receiver address (not a stride address)",
		//	func() []interface{} {
		//		return []interface{}{
		//			cmn.Coin{
		//				Denom:  "aevmos",
		//				Amount: big.NewInt(10000000),
		//			},
		//			"cosmos1xv9tklw7d82sezh9haa573wufgy59vmwe6xxe5",
		//		}
		//	},
		//	func() {},
		//	200000,
		//	true,
		//	"receiverAddress is not a stride address",
		//},
		//{
		//	"fail - receiver address is not a valid bech32",
		//	func() []interface{} {
		//		return []interface{}{
		//			cmn.Coin{
		//				Denom:  "aevmos",
		//				Amount: big.NewInt(10000000),
		//			},
		//			"stride1xv9tklw7d82sezh9haa573wufgy9vmwe6xxe5",
		//		}
		//	},
		//	func() {},
		//	200000,
		//	true,
		//	"invalid bech32 address: decoding bech32 failed: invalid checksum",
		//},
		{
			"success",
			func() []interface{} {
				return []interface{}{

					"stride1mdna37zrprxl7kn0rj4e58ndp084fzzwcxhrh2",
				}
			},
			func() {},
			20000,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			_, err := s.precompile.LiquidStake(s.ctx, s.address, s.stateDB, &method, tc.malleate())

			if tc.expError {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
				tc.postCheck()
			}
		})
	}
}
