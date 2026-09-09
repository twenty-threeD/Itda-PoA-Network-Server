package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) RecordPayment(ctx context.Context, msg *types.MsgRecordPayment) (*types.MsgRecordPaymentResponse, error) {
	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoAuthority, err.Error())
	}

	if authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	if _, err := k.addressCodec.StringToBytes(msg.BuyerAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "invalid buyer address %s", msg.BuyerAddress)
	}

	exists, err := k.HasRecord(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errorsmod.Wrapf(types.ErrDuplicateOrderID, "order id %s", msg.OrderId)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	record := msg.ToRecord(sdkCtx.BlockHeight())

	if err := record.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetRecord(ctx, record); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecordPayment,
			sdk.NewAttribute(types.AttributeKeyOrderID, record.OrderId),
			sdk.NewAttribute(types.AttributeKeyBuyerAddress, record.BuyerAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, fmt.Sprintf("%d", record.Amount)),
			sdk.NewAttribute(types.AttributeKeyRecordedHeight, fmt.Sprintf("%d", record.RecordedHeight)),
		),
	)

	return &types.MsgRecordPaymentResponse{RecordedHeight: record.RecordedHeight}, nil
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoAuthority, err.Error())
	}

	if authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	sdk.UnwrapSDKContext(ctx).EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeUpdateParams),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}
