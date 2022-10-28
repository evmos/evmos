package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/evmos/evmos/v9/tests/e2e/chain"
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
	// dataDir = "/home/rama/tests/chain"

	if len(dataDir) == 0 {
		log.Fatal("data-dir is required")
	}

	if err := os.MkdirAll(dataDir, 0o436); err != nil {
		log.Fatalf("can't create data-dir")
	}

	createdChain, err := chain.Init(chainID, dataDir, valConfig)
	if err != nil {
		log.Fatalf("can't initialize chain")
	}

	b, err := json.Marshal(createdChain)
	if err != nil {
		log.Fatalf("marshaling chain error: %s", err)
	}
	fileName := fmt.Sprintf("%v/%v-encode", dataDir, chainID)
	if err = os.WriteFile(fileName, b, 0o436); err != nil {
		log.Fatalf("can't write chain data")
	}
}
