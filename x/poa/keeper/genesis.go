package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) ([]abci.ValidatorUpdate, error) {
	if err := gs.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetParams(ctx, gs.Params); err != nil {
		return nil, err
	}

	if err := k.SetAuthority(ctx, gs.Authority); err != nil {
		return nil, err
	}

	updates := make([]abci.ValidatorUpdate, 0, len(gs.Validators))

	for i := range gs.Validators {
		validator := gs.Validators[i]

		if err := k.SetValidator(ctx, validator); err != nil {
			return nil, err
		}

		pubKey, err := validator.ConsPubKey()
		if err != nil {
			return nil, err
		}

		cmtPubKey, err := cryptocodec.ToCmtProtoPublicKey(pubKey)
		if err != nil {
			return nil, err
		}

		updates = append(updates, abci.ValidatorUpdate{
			PubKey: cmtPubKey,
			Power:  validator.Power,
		})
	}

	return updates, nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	validators, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewGenesisState(params, validators, authority), nil
}
