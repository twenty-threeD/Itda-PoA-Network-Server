package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

var _ types.QueryServer = Querier{}

func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := q.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) Payment(ctx context.Context, req *types.QueryPaymentRequest) (*types.QueryPaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := types.ValidateOrderID(req.OrderId); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	record, err := q.GetRecord(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "payment record %s not found", req.OrderId)
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentResponse{Record: record}, nil
}

func (q Querier) Payments(ctx context.Context, req *types.QueryPaymentsRequest) (*types.QueryPaymentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	records, pageRes, err := query.CollectionPaginate(
		ctx,
		q.Keeper.Records,
		req.Pagination,
		func(_ string, record types.PaymentRecord) (types.PaymentRecord, error) {
			return record, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentsResponse{
		Records:    records,
		Pagination: pageRes,
	}, nil
}
