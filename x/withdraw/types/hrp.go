package types

import (
	"errors"
	fmt "fmt"
	"strings"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

func (p Prefix) Validate() error {
	if strings.TrimSpace(p.Bech32HRP) == "" {
		return errors.New("bech32 HRP cannot be blank")
	}

	if !channeltypes.IsValidChannelID(p.SourceChannel) {
		return fmt.Errorf("invalid channel id %s", p.SourceChannel)
	}

	return nil
}
