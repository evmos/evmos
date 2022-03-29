package app

import (
	"testing"

	"github.com/tharsis/ethermint/encoding"
)

func FuzzEncodingTxDecoder(f *testing.F) {
	f.Skip("TODO (@fedekunze, @odeke-em): Add the corpus")

	txConfig := encoding.MakeConfig(ModuleBasics).TxConfig

	// TODO: Add the corpus.
	f.Add([]byte(""))

	f.Fuzz(func(t *testing.T, data []byte) {
		tx, err := txConfig.TxDecoder()(data)
		if tx == nil && err == nil {
			t.Fatal("nil tx yet nil err")
		}
	})
}
