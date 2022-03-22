package ibc

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
)

var _ porttypes.IBCModule = &MockIBCModule{}

// MockIBCModule defines a mocked object that implements the IBCModule
// interface. It's used on tests to abstract the complexity of IBC callbacks.
type MockIBCModule struct {
	mock.Mock
}

// OnChanOpenInit implements the Module interface
// It calls the underlying app's OnChanOpenInit callback.
func (m MockIBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnChanOpenTry implements the Module interface.
// It calls the underlying app's OnChanOpenTry callback.
func (m MockIBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// OnChanOpenAck implements the Module interface.
// It calls the underlying app's OnChanOpenAck callback.
func (m MockIBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelID,
	counterpartyVersion string,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnChanOpenConfirm implements the Module interface.
// It calls the underlying app's OnChanOpenConfirm callback.
func (m MockIBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnChanCloseInit implements the Module interface
// It calls the underlying app's OnChanCloseInit callback.
func (m MockIBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnChanCloseConfirm implements the Module interface.
// It calls the underlying app's OnChanCloseConfirm callback.
func (m MockIBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnRecvPacket implements the Module interface.
// It calls the underlying app's OnRecvPacket callback.
func (m MockIBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	args := m.Called()
	return args.Get(0).(exported.Acknowledgement)
}

// OnAcknowledgementPacket implements the Module interface.
// It calls the underlying app's OnAcknowledgementPacket callback.
func (m MockIBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	args := m.Called()
	return args.Error(0)
}

// OnTimeoutPacket implements the Module interface.
// It calls the underlying app's OnTimeoutPacket callback.
func (m MockIBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	args := m.Called()
	return args.Error(0)
}

func TestModule(t *testing.T) {
	mockModule := &MockIBCModule{}
	mockModule.On("OnChanOpenInit").Return(nil)
	mockModule.On("OnChanOpenTry").Return("", nil)
	mockModule.On("OnChanOpenAck").Return(nil)
	mockModule.On("OnChanOpenConfirm").Return(nil)
	mockModule.On("OnChanCloseInit").Return(nil)
	mockModule.On("OnChanCloseConfirm").Return(nil)
	mockModule.On("OnRecvPacket").Return(channeltypes.NewResultAcknowledgement([]byte("ack")))
	mockModule.On("OnAcknowledgementPacket").Return(nil)
	mockModule.On("OnTimeoutPacket").Return(nil)

	module := NewModule(mockModule)

	// mock calls for abstraction
	err := module.OnChanOpenInit(sdk.Context{}, channeltypes.ORDERED, nil, transfertypes.PortID, "channel-0", &capabilitytypes.Capability{}, channeltypes.Counterparty{}, "")
	require.NoError(t, err)
	_, err = module.OnChanOpenTry(sdk.Context{}, channeltypes.ORDERED, nil, transfertypes.PortID, "channel-0", &capabilitytypes.Capability{}, channeltypes.Counterparty{}, "")
	require.NoError(t, err)
	err = module.OnChanOpenAck(sdk.Context{}, transfertypes.PortID, "channel-0", "channel-0", "")
	require.NoError(t, err)
	err = module.OnChanOpenConfirm(sdk.Context{}, transfertypes.PortID, "channel-0")
	require.NoError(t, err)
	err = module.OnChanCloseInit(sdk.Context{}, transfertypes.PortID, "channel-0")
	require.NoError(t, err)
	err = module.OnChanCloseConfirm(sdk.Context{}, transfertypes.PortID, "channel-0")
	require.NoError(t, err)
	ack := module.OnRecvPacket(sdk.Context{}, channeltypes.Packet{}, nil)
	require.NotNil(t, ack)
	err = module.OnAcknowledgementPacket(sdk.Context{}, channeltypes.Packet{}, nil, nil)
	require.NoError(t, err)
	err = module.OnTimeoutPacket(sdk.Context{}, channeltypes.Packet{}, nil)
	require.NoError(t, err)
}
