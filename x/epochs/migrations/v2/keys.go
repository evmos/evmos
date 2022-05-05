package v2

// prefix bytes for the epochs persistent store
const (
	prefixEpoch = iota + 1
	prefixEpochDuration
)

// KeyPrefixEpoch defines prefix key for storing epochs
var KeyPrefixEpoch = []byte{prefixEpoch}

// KeyPrefixEpochDuration defines prefix key for storing epochs durations
var KeyPrefixEpochDuration = []byte{prefixEpochDuration}
