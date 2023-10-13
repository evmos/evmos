package stride

import (
	errorsmod "cosmossdk.io/errors"
)

const (
	OutpostName = "stride-outpost"
)

var (
	ErrInvalidPacketMetadata     = errorsmod.Register(OutpostName, 1501, "invalid packet metadata")
	ErrUnsupportedStakeibcAction = errorsmod.Register(OutpostName, 1502, "unsupported stakeibc action")
	ErrInvalidClaimAirdropId     = errorsmod.Register(OutpostName, 1503, "invalid claim airdrop ID (cannot be empty)")
	ErrInvalidModuleRoutes       = errorsmod.Register(OutpostName, 1504, "invalid number of module routes, only 1 module is allowed at a time")
	ErrUnsupportedAutopilotRoute = errorsmod.Register(OutpostName, 1505, "unsupported autpilot route")
	ErrInvalidReceiverAddress    = errorsmod.Register(OutpostName, 1506, "receiver address must be specified when using autopilot")
	ErrPacketForwardingInactive  = errorsmod.Register(OutpostName, 1507, "autopilot packet forwarding is disabled")
	ErrInvalidMemoSize           = errorsmod.Register(OutpostName, 1508, "the memo or receiver field exceeded the max allowable size")
)
