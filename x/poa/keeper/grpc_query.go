package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
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

func (q Querier) Validators(ctx context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	validators, pageRes, err := query.CollectionPaginate(
		ctx,
		q.Keeper.Validators,
		req.Pagination,
		func(_ string, validator types.Validator) (types.Validator, error) {
			return validator, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorsResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

func (q Querier) Validator(ctx context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.OperatorAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "operator address cannot be empty")
	}

	validator, err := q.GetValidator(ctx, req.OperatorAddress)
	if err != nil {
		if errorsIsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "validator %s not found", req.OperatorAddress)
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorResponse{Validator: validator}, nil
}

func errorsIsNotFound(err error) bool {
	return errors.Is(err, collections.ErrNotFound)
}
