package keeper

import (
	"context"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return err
	}
	if err := k.SetAuthority(ctx, gs.Authority); err != nil {
		return err
	}
	for i := range gs.Records {
		if err := k.SetRecord(ctx, gs.Records[i]); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, err
	}

	records, err := k.GetAllRecords(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, authority, records), nil
}
