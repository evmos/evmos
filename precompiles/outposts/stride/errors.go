package stride

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPacketMetadata     = errorsmod.Register(ModuleName, 1501, "invalid packet metadata")
	ErrUnsupportedStakeibcAction = errorsmod.Register(ModuleName, 1502, "unsupported stakeibc action")
	ErrInvalidClaimAirdropId     = errorsmod.Register(ModuleName, 1503, "invalid claim airdrop ID (cannot be empty)")
	ErrInvalidModuleRoutes       = errorsmod.Register(ModuleName, 1504, "invalid number of module routes, only 1 module is allowed at a time")
	ErrUnsupportedAutopilotRoute = errorsmod.Register(ModuleName, 1505, "unsupported autpilot route")
	ErrInvalidReceiverAddress    = errorsmod.Register(ModuleName, 1506, "receiver address must be specified when using autopilot")
	ErrPacketForwardingInactive  = errorsmod.Register(ModuleName, 1507, "autopilot packet forwarding is disabled")
	ErrInvalidMemoSize           = errorsmod.Register(ModuleName, 1508, "the memo or receiver field exceeded the max allowable size")
)
