package utils

//goland:noinspection SpellCheckingInspection
import (
	"encoding/hex"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Keccak256 is function used in testing only, it has the dedicated implementation for purpose of double-check result
func Keccak256(input string) string {
	return hex.EncodeToString(gethcrypto.Keccak256Hash([]byte(input)).Bytes())
}
