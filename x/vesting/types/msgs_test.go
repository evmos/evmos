package types

import (
	"testing"

	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/require"
)

func TestClawbackVestingAccountMsg(t *testing.T) {
	_, _, fromAddr := KeyTestPubAddr()
	_, _, toAddr := KeyTestPubAddr()
	amount := NewTestCoins()
	startTime := int64(100200300)
	lockupPeriods := []sdkvesting.Period{
		{Length: 200000, Amount: amount},
	}
	vestingPeriods := []sdkvesting.Period{
		{Length: 300000, Amount: amount},
	}
	msg := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, vestingPeriods, false)
	route := msg.Route()
	require.Equal(t, RouterKey, route)
	tp := msg.Type()
	require.Equal(t, TypeMsgCreateClawbackVestingAccount, tp)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	badFromMsg := MsgCreateClawbackVestingAccount{
		FromAddress:    "foo",
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
	}
	err = badFromMsg.ValidateBasic()
	require.Error(t, err)

	badToMsg := MsgCreateClawbackVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      "foo",
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
	}
	err = badToMsg.ValidateBasic()
	require.Error(t, err)

	badPeriods := []sdkvesting.Period{{Length: 0, Amount: amount}}
	badLockup := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, badPeriods, vestingPeriods, false)
	err = badLockup.ValidateBasic()
	require.Error(t, err)

	badVesting := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, badPeriods, false)
	err = badVesting.ValidateBasic()
	require.Error(t, err)

	badAmounts := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, []sdkvesting.Period{
		{Length: 17, Amount: amount.Add(amount...)},
	}, false)
	err = badAmounts.ValidateBasic()
	require.Error(t, err)

	emptyPeriods := []sdkvesting.Period{}
	noLockupOk := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, emptyPeriods, vestingPeriods, false)
	err = noLockupOk.ValidateBasic()
	require.NoError(t, err)

	noVestingOk := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, emptyPeriods, false)
	err = noVestingOk.ValidateBasic()
	require.NoError(t, err)
}

func TestClawbackMsg(t *testing.T) {
	_, _, funderAddr := KeyTestPubAddr()
	_, _, addr := KeyTestPubAddr()
	_, _, destAddr := KeyTestPubAddr()

	okMsg := NewMsgClawback(funderAddr, addr, destAddr)
	route := okMsg.Route()
	require.Equal(t, RouterKey, route)
	tp := okMsg.Type()
	require.Equal(t, TypeMsgClawback, tp)
	err := okMsg.ValidateBasic()
	require.NoError(t, err)

	noDest := NewMsgClawback(funderAddr, addr, nil)
	require.Equal(t, noDest.DestAddress, "")
	err = noDest.ValidateBasic()
	require.NoError(t, err)

	badFunder := MsgClawback{
		FunderAddress: "foo",
		Address:       addr.String(),
		DestAddress:   destAddr.String(),
	}
	err = badFunder.ValidateBasic()
	require.Error(t, err)

	badAddr := MsgClawback{
		FunderAddress: funderAddr.String(),
		Address:       "foo",
		DestAddress:   destAddr.String(),
	}
	err = badAddr.ValidateBasic()
	require.Error(t, err)

	badDest := MsgClawback{
		FunderAddress: funderAddr.String(),
		Address:       addr.String(),
		DestAddress:   "foo",
	}
	err = badDest.ValidateBasic()
	require.Error(t, err)
}
