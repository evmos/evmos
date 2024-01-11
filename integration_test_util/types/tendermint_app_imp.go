package types

import (
	"fmt"
	nm "github.com/cometbft/cometbft/node"
	"strings"
)

var _ TendermintApp = &tendermintAppImp{}

type tendermintAppImp struct {
	tendermintNode *nm.Node
	rpcAddr        string
	grpcAddr       string
}

func NewTendermintApp(tendermintNode *nm.Node, rpcPort int) TendermintApp {
	app := &tendermintAppImp{
		tendermintNode: tendermintNode,
	}
	if rpcPort > 0 {
		app.rpcAddr = fmt.Sprintf("tcp://localhost:%d", rpcPort)
	}
	return app
}

func (a *tendermintAppImp) TendermintNode() *nm.Node {
	return a.tendermintNode
}

func (a *tendermintAppImp) GetRpcAddr() (addr string, supported bool) {
	return a.rpcAddr, a.rpcAddr != ""
}

func (a *tendermintAppImp) Shutdown() {
	if a == nil || a.tendermintNode == nil || !a.tendermintNode.IsRunning() {
		return
	}
	err := a.tendermintNode.Stop()
	if err != nil {
		if strings.Contains(err.Error(), "already stopped") {
			// ignore
		} else {
			fmt.Println("Failed to stop tendermint node")
			fmt.Println(err)
		}
	}
	a.tendermintNode.Wait()
}
