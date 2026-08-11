package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
)

func (k Keeper) EndBlocker(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return k.DrainValidatorUpdates(ctx)
}
