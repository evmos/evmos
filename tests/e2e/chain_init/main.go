package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tharsis/evmos/v4/tests/e2e/chain"
)

func main() {
	var (
		valConfig []*chain.ValidatorConfig
		dataDir   string
		chainID   string
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainID, "chain-id", "", "chain ID")

	flag.Parse()

	valConfig = make([]*chain.ValidatorConfig, 2)
	valConfig[0] = &chain.ValidatorConfig{Pruning: "default", PruningKeepRecent: "0", PruningInterval: "0"}
	valConfig[1] = &chain.ValidatorConfig{Pruning: "default", PruningKeepRecent: "0", PruningInterval: "0"}

	// To test locally
	// chainID = "evmos_9001-1"
	// dataDir = "/home/rama/chain"

	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o644); err != nil {
		panic(err)
	}

	createdChain, err := chain.Init(chainID, dataDir, valConfig)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(createdChain)
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainID)
	if err = os.WriteFile(fileName, b, 0o644); err != nil {
		panic(err)
	}
}
