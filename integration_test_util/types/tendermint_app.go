package types

import nm "github.com/cometbft/cometbft/node"

type TendermintApp interface {
	TendermintNode() *nm.Node
	GetRpcAddr() (addr string, supported bool)
	Shutdown()
}
