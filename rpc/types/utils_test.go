package types

import (
	"math/big"
	"testing"
)

func TestCheckTxFeeInvalidArgs(t *testing.T) {
	if err := CheckTxFee(nil, 10, 100); err == nil {
		t.Fatal("expecting a non-nil error")
	}

	gp := big.NewInt(-1)
	if err := CheckTxFee(gp, 10, 100); err == nil {
		t.Fatal("expecting a non-nil error")
	}
}
