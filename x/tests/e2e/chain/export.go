package chain

import "fmt"

type Meta struct {
	DataDir string `json:"dataDir"`
	ID      string `json:"id"`
}

type Validator struct {
	Name          string `json:"name"`
	ConfigDir     string `json:"configDir"`
	Index         int    `json:"index"`
	Mnemonic      string `json:"mnemonic"`
	PublicAddress string `json:"publicAddress"`
}

type Chain struct {
	ChainMeta  Meta         `json:"chainMeta"`
	Validators []*Validator `json:"validators"`
}

func (c *Meta) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.ID)
}
