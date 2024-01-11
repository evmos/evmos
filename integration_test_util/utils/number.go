package utils

import (
	cryptorand "crypto/rand"
	"github.com/pkg/errors"
	"math/rand"
)

// RandomPositiveInt64 returns a random int64, value >= 0.
func RandomPositiveInt64() int64 {
	val := rand.Int63()
	if val < 0 {
		val = -val
	}
	return val
}

// GenRandomBytes returns generated random bytes array, with specified length.
func GenRandomBytes(size int) []byte {
	bz := make([]byte, size)
	if size > 0 {
		_, err := cryptorand.Read(bz)
		if err != nil {
			panic(errors.Wrap(err, "failed to generate random bytes"))
		}
	}
	return bz
}
