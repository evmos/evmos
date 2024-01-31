package ante_test

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type EvmAnteTestSuite struct {
	suite.Suite
}

func TestEvmAnteTestSuite(t *testing.T) {
	suite.Run(t, &EvmAnteTestSuite{})
}
