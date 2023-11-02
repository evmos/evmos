package readonly

import (
	"github.com/cosmos/cosmos-sdk/store/types"
)

var _ types.KVStore = &Store{}

// Store is a read-only version of a KVStore.
type Store struct {
	types.KVStore
}

// NewStore create a new instance of a read-only store.
func NewStore(store types.KVStore) *Store {
	return &Store{KVStore: store}
}

// Set panics as it is not supported on a read-only KVStore.
func (s Store) Set(_, _ []byte) {
	panic("cannot call Set on a read-only store")
}

// Delete panics as it is not supported on a read-only KVStore.
func (s Store) Delete(_ []byte) {
	panic("cannot call Delete on a read-only store")
}

// Write panics as it is not supported on a read-only KVStore.
func (s Store) Write() {
	panic("cannot call Write on a read-only store")
}
