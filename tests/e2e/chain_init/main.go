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
		chainId   string
		config    string
	)

	flag.StringVar(&dataDir, "data-dir", "", "chain data directory")
	flag.StringVar(&chainId, "chain-id", "", "chain ID")
	flag.StringVar(&config, "config", "", "serialized config")

	flag.Parse()

	// err := json.Unmarshal([]byte(s), &valConfig)
	// if err != nil {
	// 	panic(err)
	// }

	valConfig = make([]*chain.ValidatorConfig, 2)
	valConfig[0] = &chain.ValidatorConfig{Pruning: "default", PruningKeepRecent: "0", PruningInterval: "0"}
	valConfig[1] = &chain.ValidatorConfig{Pruning: "default", PruningKeepRecent: "0", PruningInterval: "0"}
	// chainId = "evmos_9001-1"
	// dataDir = "/home/rama/chain"
	if len(dataDir) == 0 {
		panic("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		panic(err)
	}

	createdChain, err := chain.Init(chainId, dataDir, valConfig)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(createdChain)
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainId)
	if err = os.WriteFile(fileName, b, 0o777); err != nil {
		panic(err)
	}
}
