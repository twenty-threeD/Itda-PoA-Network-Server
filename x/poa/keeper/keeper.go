package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	logger       log.Logger

	Schema       collections.Schema
	Authority    collections.Item[string]
	Params       collections.Item[types.Params]
	Validators   collections.Map[string, types.Validator]
	PowerUpdates collections.Map[string, abci.ValidatorUpdate]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	addressCodec address.Codec,
	logger log.Logger,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		logger:       logger.With("module", "x/"+types.ModuleName),
		Authority: collections.NewItem(
			sb,
			types.AuthorityKey,
			"authority",
			collections.StringValue,
		),
		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		Validators: collections.NewMap(
			sb,
			types.ValidatorsKeyPrefix,
			"validators",
			collections.StringKey,
			codec.CollValue[types.Validator](cdc),
		),
		PowerUpdates: collections.NewMap(
			sb,
			types.PowerUpdatesKeyPrefix,
			"power_updates",
			collections.StringKey,
			codec.CollValue[abci.ValidatorUpdate](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

func (k Keeper) GetAuthority(ctx context.Context) (string, error) {
	return k.Authority.Get(ctx)
}

func (k Keeper) SetAuthority(ctx context.Context, authority string) error {
	return k.Authority.Set(ctx, authority)
}

func (k Keeper) Logger() log.Logger {
	return k.logger
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.Params.Get(ctx)
}

func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}

func (k Keeper) GetValidator(ctx context.Context, operatorAddress string) (types.Validator, error) {
	return k.Validators.Get(ctx, operatorAddress)
}

func (k Keeper) HasValidator(ctx context.Context, operatorAddress string) (bool, error) {
	return k.Validators.Has(ctx, operatorAddress)
}

func (k Keeper) SetValidator(ctx context.Context, validator types.Validator) error {
	return k.Validators.Set(ctx, validator.OperatorAddress, validator)
}

func (k Keeper) RemoveValidator(ctx context.Context, operatorAddress string) error {
	return k.Validators.Remove(ctx, operatorAddress)
}

func (k Keeper) CountValidators(ctx context.Context) (uint32, error) {
	var count uint32

	err := k.Validators.Walk(ctx, nil, func(string, types.Validator) (bool, error) {
		count++

		return false, nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (k Keeper) GetAllValidators(ctx context.Context) ([]types.Validator, error) {
	validators := make([]types.Validator, 0)

	err := k.Validators.Walk(ctx, nil, func(_ string, validator types.Validator) (bool, error) {
		validators = append(validators, validator)

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return validators, nil
}

func (k Keeper) QueueValidatorUpdate(ctx context.Context, operatorAddress string, pubKey cryptotypes.PubKey, power int64) error {
	cmtPubKey, err := cryptocodec.ToCmtProtoPublicKey(pubKey)
	if err != nil {
		return err
	}

	return k.PowerUpdates.Set(ctx, operatorAddress, abci.ValidatorUpdate{
		PubKey: cmtPubKey,
		Power:  power,
	})
}

func (k Keeper) DrainValidatorUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	updates := make([]abci.ValidatorUpdate, 0)
	keys := make([]string, 0)

	err := k.PowerUpdates.Walk(ctx, nil, func(key string, update abci.ValidatorUpdate) (bool, error) {
		updates = append(updates, update)
		keys = append(keys, key)

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if err := k.PowerUpdates.Remove(ctx, key); err != nil {
			return nil, err
		}
	}

	return updates, nil
}

func (k Keeper) IsPubKeyInUse(ctx context.Context, pubKeyBytes []byte) (bool, error) {
	inUse := false

	err := k.Validators.Walk(ctx, nil, func(_ string, validator types.Validator) (bool, error) {
		existing, err := validator.ConsPubKey()
		if err != nil {
			return true, err
		}

		if string(existing.Bytes()) == string(pubKeyBytes) {
			inUse = true

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return false, err
	}

	return inUse, nil
}
