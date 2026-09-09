package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	logger       log.Logger

	Schema    collections.Schema
	Params    collections.Item[types.Params]
	Authority collections.Item[string]
	Records   collections.Map[string, types.PaymentRecord]
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
		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),
		Authority: collections.NewItem(
			sb,
			types.AuthorityKey,
			"authority",
			collections.StringValue,
		),
		Records: collections.NewMap(
			sb,
			types.RecordsKeyPrefix,
			"records",
			collections.StringKey,
			codec.CollValue[types.PaymentRecord](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

func (k Keeper) Logger() log.Logger                               { return k.logger }
func (k Keeper) GetAuthority(ctx context.Context) (string, error) { return k.Authority.Get(ctx) }
func (k Keeper) SetAuthority(ctx context.Context, authority string) error {
	return k.Authority.Set(ctx, authority)
}
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) { return k.Params.Get(ctx) }
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}
func (k Keeper) GetRecord(ctx context.Context, orderID string) (types.PaymentRecord, error) {
	return k.Records.Get(ctx, orderID)
}
func (k Keeper) HasRecord(ctx context.Context, orderID string) (bool, error) {
	return k.Records.Has(ctx, orderID)
}
func (k Keeper) SetRecord(ctx context.Context, record types.PaymentRecord) error {
	return k.Records.Set(ctx, record.OrderId, record)
}
func (k Keeper) GetAllRecords(ctx context.Context) ([]types.PaymentRecord, error) {
	records := make([]types.PaymentRecord, 0)

	err := k.Records.Walk(ctx, nil, func(_ string, record types.PaymentRecord) (bool, error) {
		records = append(records, record)

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}
