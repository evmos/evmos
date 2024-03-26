package keeper_test

import (
	"fmt"
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v12/x/erc20/keeper"
	"github.com/evmos/evmos/v12/x/erc20/types"
	"github.com/stretchr/testify/require"
)

func TestERC20BankContractRegistrationHook_AfterDenomMetadataCreation(t *testing.T) {
	type fields struct {
		erc20Keeper *mockKeeper
	}
	type args struct {
		newDenomMetadata banktypes.Metadata
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		expectRegistered bool
		expectErr        string
	}{
		{
			name: "success",
			fields: fields{
				erc20Keeper: &mockKeeper{},
			},
			args: args{
				newDenomMetadata: banktypes.Metadata{
					Base: "ibc/eth",
				},
			},
			expectRegistered: true,
			expectErr:        "",
		}, {
			name: "not ibc denom",
			fields: fields{
				erc20Keeper: &mockKeeper{},
			},
			args: args{
				newDenomMetadata: banktypes.Metadata{
					Base: "eth",
				},
			},
			expectRegistered: false,
			expectErr:        "",
		}, {
			name: "error",
			fields: fields{
				erc20Keeper: &mockKeeper{
					err: fmt.Errorf("error"),
				},
			},
			args: args{
				newDenomMetadata: banktypes.Metadata{
					Base: "ibc/eth",
				},
			},
			expectRegistered: false,
			expectErr:        "deploy the erc20 contract for the ibc coin: ibc/eth; error: error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := keeper.NewERC20ContractRegistrationHook(tt.fields.erc20Keeper)
			err := e.AfterDenomMetadataCreation(sdk.Context{}, tt.args.newDenomMetadata)
			if tt.expectErr != "" {
				require.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.fields.erc20Keeper.registered, tt.expectRegistered)
		})
	}
}

type mockKeeper struct {
	registered bool
	err        error
	sync.Mutex
}

func (m *mockKeeper) RegisterCoin(sdk.Context, banktypes.Metadata) (*types.TokenPair, error) {
	m.Lock()
	defer m.Unlock()
	m.registered = m.err == nil
	return nil, m.err
}
