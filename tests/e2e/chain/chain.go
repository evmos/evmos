package chain

import (
	"fmt"
)

const (
	keyringPassphrase = "testpassphrase"
)

// internalChain contains the same info as chain, but with the validator structs instead using the internal validator
// representation, with more derived data
type internalChain struct {
	chainMeta  Meta
	validators []*internalValidator
}

func new(id, dataDir string) *internalChain {
	chainMeta := Meta{
		ID:      id,
		DataDir: dataDir,
	}
	return &internalChain{
		chainMeta: chainMeta,
	}
}

func (c *internalChain) createAndInitValidators(count int) error {
	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(); err != nil {
			return err
		}

		c.validators = append(c.validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *internalChain) createValidator(index int) *internalValidator {
	return &internalValidator{
		chain:   c,
		index:   index,
		moniker: fmt.Sprintf("%s-%d", c.chainMeta.ID, index+1),
	}
}

func (c *internalChain) export() *Chain {
	exportValidators := make([]*Validator, 0, len(c.validators))
	for _, v := range c.validators {
		exportValidators = append(exportValidators, v.export())
	}

	return &Chain{
		ChainMeta:  c.chainMeta,
		Validators: exportValidators,
	}
}
