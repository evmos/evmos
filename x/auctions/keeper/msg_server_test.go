package keeper_test

import (
	"testing"

	testkeyring "github.com/evmos/evmos/v19/testutil/integration/evmos/keyring"
	testnetwork "github.com/evmos/evmos/v19/testutil/integration/evmos/network"
	testutiltx "github.com/evmos/evmos/v19/testutil/tx"
	"github.com/stretchr/testify/assert"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/evmos/evmos/v19/utils"
	"github.com/evmos/evmos/v19/x/auctions/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestBid(t *testing.T) {
	var network *testnetwork.UnitTestNetwork

	// validSenderKey is populated during each
	// run and used to create the MsgBid.
	var validSenderKey testkeyring.Key
	bidAmount := sdk.NewInt(5)
	emptyAddress, _ := testutiltx.NewAccAddressAndKey()

	testCases := []struct {
		name        string
		malleate    func()               // used to modify the initial conditions of the test
		input       func() *types.MsgBid // return the message used for the bid
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			name: "pass with no previous bid",
			malleate: func() {
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck: func() {},
			expErr:    false,
		},
		{
			name: "pass with previous bid present",
			malleate: func() {
				// Send coins from the valid sender to an empty account. In thsi
				// way we can easily verify the expected final balance.
				emptyAccountCoin := sdk.NewCoin(utils.BaseDenom, bidAmount.Sub(sdk.NewInt(1)))
				err := network.App.BankKeeper.SendCoins(network.GetContext(), validSenderKey.AccAddr, emptyAddress, sdk.NewCoins(emptyAccountCoin))
				assert.NoError(t, err, "failed to send coins from valid sender to empty account")
				bigMsg := &types.MsgBid{
					Sender: emptyAddress.String(),
					Amount: emptyAccountCoin,
				}
				_, err = network.App.AuctionsKeeper.Bid(network.GetContext(), bigMsg)
				assert.NoError(t, err, "failed to create setup bid")
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck: func() {
				resp := network.App.BankKeeper.GetBalance(network.GetContext(), emptyAddress, utils.BaseDenom)
				assert.Equal(t, resp.Amount, bidAmount.Sub(sdk.NewInt(1)))
			},
			expErr: false,
		},
		{
			name: "fail auction not enabled",
			malleate: func() {
				// Update params to disable the auction.
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: types.ErrAuctionDisabled.Error(),
		},
		{
			name: "fail when an higher bid is already present",
			malleate: func() {
				bigMsg := &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount.Add(sdk.NewInt(1))),
				}
				_, err := network.App.AuctionsKeeper.Bid(network.GetContext(), bigMsg)
				assert.NoError(t, err, "failed to create setup bid")
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: types.ErrBidMustBeHigherThanCurrent.Error(),
		},
		{
			name: "fail when a bid with same amount is already present",
			malleate: func() {
				bigMsg := &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
				_, err := network.App.AuctionsKeeper.Bid(network.GetContext(), bigMsg)
				assert.NoError(t, err, "failed to create setup bid")
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: types.ErrBidMustBeHigherThanCurrent.Error(),
		},
		{
			name: "fail when sender is not valid bech32",
			malleate: func() {
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: "",
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: "invalid sender address",
		},
		{
			name: "fail when sender does not have enough funds",
			malleate: func() {
			},
			input: func() *types.MsgBid {
				return &types.MsgBid{
					Sender: emptyAddress.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, bidAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: "transfer bid coins failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			validSenderKey = keyring.GetKey(0)

			// Update chain environment before executing the message.
			tc.malleate()
			_, err := network.App.AuctionsKeeper.Bid(network.GetContext(), tc.input())

			if tc.expErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Contains(t, err.Error(), tc.errContains, "expected different error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "error not expected")
			}

			tc.postCheck()
		})
	}
}

func TestDepositCoin(t *testing.T) {
	var network *testnetwork.UnitTestNetwork

	// validSenderKey is populated during each
	// run and used to create the MsgDepositCoin.
	var validSenderKey testkeyring.Key
	depositAmount := sdk.NewInt(5)
	emptyAddress, _ := testutiltx.NewAccAddressAndKey()

	testCases := []struct {
		name        string
		malleate    func()                       // used to modify the initial conditions of the test or for pre test checks
		input       func() *types.MsgDepositCoin // return the message used for the bid
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			name: "pass",
			malleate: func() {
				auctionCollectorAddress := network.App.AccountKeeper.GetModuleAddress(types.AuctionCollectorName)
				resp := network.App.BankKeeper.GetBalance(network.GetContext(), auctionCollectorAddress, utils.BaseDenom)
				assert.Equal(t, resp.Amount.Equal(sdk.ZeroInt()), true, "auction collector not empty")
			},
			input: func() *types.MsgDepositCoin {
				return &types.MsgDepositCoin{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, depositAmount),
				}
			},
			postCheck: func() {
				auctionCollectorAddress := network.App.AccountKeeper.GetModuleAddress(types.AuctionCollectorName)
				resp := network.App.BankKeeper.GetBalance(network.GetContext(), auctionCollectorAddress, utils.BaseDenom)
				assert.Equal(t, resp.Amount, depositAmount, "expected the auction collector to have the deposit")
			},
			expErr: false,
		},
		{
			name: "fail auction not enabled",
			malleate: func() {
				// Update params to disable the auction.
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				updateParamsMsg := types.MsgUpdateParams{
					Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
					Params:    params,
				}
				_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), &updateParamsMsg)
				assert.NoError(t, err, "failed to update auctions params")
			},
			input: func() *types.MsgDepositCoin {
				return &types.MsgDepositCoin{
					Sender: validSenderKey.AccAddr.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, depositAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: types.ErrAuctionDisabled.Error(),
		},
		{
			name: "fail when sender is not valid bech32",
			malleate: func() {
			},
			input: func() *types.MsgDepositCoin {
				return &types.MsgDepositCoin{
					Sender: "",
					Amount: sdk.NewCoin(utils.BaseDenom, depositAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: "invalid sender address",
		},
		{
			name: "fail when sender does not have enough funds",
			malleate: func() {
			},
			input: func() *types.MsgDepositCoin {
				return &types.MsgDepositCoin{
					Sender: emptyAddress.String(),
					Amount: sdk.NewCoin(utils.BaseDenom, depositAmount),
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: "transfer of deposit failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			validSenderKey = keyring.GetKey(0)

			// Update chain environment before executing the message.
			tc.malleate()
			_, err := network.App.AuctionsKeeper.DepositCoin(network.GetContext(), tc.input())

			if tc.expErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Contains(t, err.Error(), tc.errContains, "expected different error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "error not expected")
			}

			tc.postCheck()
		})
	}
}

func TestUpdateParams(t *testing.T) {
	var network *testnetwork.UnitTestNetwork

	var eoaKey testkeyring.Key
	authorityAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	testCases := []struct {
		name        string
		preCheck    func()                        // used to modify the initial conditions of the test or for pre test checks
		input       func() *types.MsgUpdateParams // return the message used for the bid
		postCheck   func()
		expErr      bool
		errContains string
	}{
		{
			name: "pass",
			preCheck: func() {
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				assert.Equal(t, params.EnableAuction, true)
			},
			input: func() *types.MsgUpdateParams {
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				return &types.MsgUpdateParams{
					Authority: authorityAddress,
					Params:    params,
				}
			},
			postCheck: func() {
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				assert.Equal(t, params.EnableAuction, false, "expected params to be updated")
			},
			expErr: false,
		},
		{
			name:     "fail when wrong authority",
			preCheck: func() {},
			input: func() *types.MsgUpdateParams {
				params := network.App.AuctionsKeeper.GetParams(network.GetContext())
				params.EnableAuction = false
				return &types.MsgUpdateParams{
					Authority: eoaKey.AccAddr.String(),
					Params:    params,
				}
			},
			postCheck:   func() {},
			expErr:      true,
			errContains: govtypes.ErrInvalidSigner.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyring := testkeyring.New(1)
			network = testnetwork.NewUnitTestNetwork(
				testnetwork.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
			)

			eoaKey = keyring.GetKey(0)

			tc.preCheck()

			_, err := network.App.AuctionsKeeper.UpdateParams(network.GetContext(), tc.input())

			if tc.expErr {
				assert.Error(t, err, "expected error but got nil")
				assert.Contains(t, err.Error(), tc.errContains, "expected different error for test case: %s", tc.name)
			} else {
				assert.NoError(t, err, "error not expected")
			}

			tc.postCheck()
		})
	}
}