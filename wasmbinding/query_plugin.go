package wasmbinding

import (
	"encoding/json"
	"fmt"

	tokenfactorybindings "github.com/cosmos/gaia/v23/x/tokenfactory/bindings"

	tokenfactorybindingstypes "github.com/cosmos/gaia/v23/x/tokenfactory/bindings/types"

	errorsmod "cosmossdk.io/errors"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CustomQuerier(tfqp *tokenfactorybindings.QueryPlugin) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var contractQuery tokenfactorybindingstypes.TokenFactoryQuery
		if err := json.Unmarshal(request, &contractQuery); err != nil {
			switch {
			case contractQuery.FullDenom != nil:
				creator := contractQuery.FullDenom.CreatorAddr
				subdenom := contractQuery.FullDenom.Subdenom

				fullDenom, err := tokenfactorybindings.GetFullDenom(creator, subdenom)
				if err != nil {
					return nil, errorsmod.Wrap(err, "gaia full denom query")
				}

				res := tokenfactorybindingstypes.FullDenomResponse{
					Denom: fullDenom,
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, errorsmod.Wrap(err, "failed to marshal FullDenomResponse")
				}

				return bz, nil

			case contractQuery.Admin != nil:
				res, err := tfqp.GetDenomAdmin(ctx, contractQuery.Admin.Denom)
				if err != nil {
					return nil, err
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("failed to JSON marshal AdminResponse: %w", err)
				}

				return bz, nil

			case contractQuery.Metadata != nil:
				res, err := tfqp.GetMetadata(ctx, contractQuery.Metadata.Denom)
				if err != nil {
					return nil, err
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("failed to JSON marshal MetadataResponse: %w", err)
				}

				return bz, nil

			case contractQuery.DenomsByCreator != nil:
				res, err := tfqp.GetDenomsByCreator(ctx, contractQuery.DenomsByCreator.Creator)
				if err != nil {
					return nil, err
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("failed to JSON marshal DenomsByCreatorResponse: %w", err)
				}

				return bz, nil

			case contractQuery.Params != nil:
				res, err := tfqp.GetParams(ctx)
				if err != nil {
					return nil, err
				}

				bz, err := json.Marshal(res)
				if err != nil {
					return nil, fmt.Errorf("failed to JSON marshal ParamsResponse: %w", err)
				}

				return bz, nil
			}
		}
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown custom query"}
	}
}
