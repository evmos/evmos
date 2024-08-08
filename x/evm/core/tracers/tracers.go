package tracers

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v19/x/evm/core/vm"
)

// Context contains some contextual infos for a transaction execution that is not
// available from within the EVM object.
type Context struct {
	BlockHash common.Hash // Hash of the block the tx is contained within (zero if dangling tx or call)
	TxIndex   int         // Index of the transaction within a block (zero if dangling tx or call)
	TxHash    common.Hash // Hash of the transaction being traced (zero if dangling call)
}

// Tracer interface extends vm.EVMLogger and additionally
// allows collecting the tracing result.
type Tracer interface {
	vm.EVMLogger
	GetResult() (json.RawMessage, error)
	// Stop terminates execution of the tracer at the first opportune moment.
	Stop(err error)
}

type lookupFunc func(string, *Context, json.RawMessage) (Tracer, error)

var lookups []lookupFunc

// RegisterLookup registers a method as a lookup for tracers, meaning that
// users can invoke a named tracer through that lookup. If 'wildcard' is true,
// then the lookup will be placed last. This is typically meant for interpreted
// engines (js) which can evaluate dynamic user-supplied code.
func RegisterLookup(wildcard bool, lookup lookupFunc) {
	if wildcard {
		lookups = append(lookups, lookup)
	} else {
		lookups = append([]lookupFunc{lookup}, lookups...)
	}
}

// registered lookups.
func New(code string, ctx *Context, cfg json.RawMessage) (Tracer, error) {
	for _, lookup := range lookups {
		if tracer, err := lookup(code, ctx, cfg); err == nil {
			return tracer, nil
		}
	}
	return nil, errors.New("tracer not found")
}
