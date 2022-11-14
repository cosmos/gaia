package keeper

import (
	"context"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/capability/keeper"
	icamauthtypes "github.com/cosmos/gaia/v8/x/icamauth/types"
	controllerkeeper "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/keeper"
)

func TestKeeper_InterchainAccount(t *testing.T) {
	type fields struct {
		cdc                 codec.Codec
		storeKey            types.StoreKey
		scopedKeeper        keeper.ScopedKeeper
		icaControllerKeeper controllerkeeper.Keeper
	}
	type args struct {
		goCtx context.Context
		req   *icamauthtypes.QueryInterchainAccountRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *icamauthtypes.QueryInterchainAccountResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := Keeper{
				cdc:                 tt.fields.cdc,
				storeKey:            tt.fields.storeKey,
				scopedKeeper:        tt.fields.scopedKeeper,
				icaControllerKeeper: tt.fields.icaControllerKeeper,
			}
			got, err := k.InterchainAccount(tt.args.goCtx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterchainAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InterchainAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}
