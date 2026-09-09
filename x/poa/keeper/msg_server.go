package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) AddValidator(ctx context.Context, msg *types.MsgAddValidator) (*types.MsgAddValidatorResponse, error) {
	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoAuthority, err.Error())
	}

	if authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	if _, err := k.addressCodec.StringToBytes(msg.OperatorAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "invalid operator address %s", msg.OperatorAddress)
	}

	exists, err := k.HasValidator(ctx, msg.OperatorAddress)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errorsmod.Wrapf(types.ErrValidatorExists, "operator %s", msg.OperatorAddress)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	count, err := k.CountValidators(ctx)
	if err != nil {
		return nil, err
	}

	if count >= params.MaxValidators {
		return nil, errorsmod.Wrapf(types.ErrMaxValidators,
			"validator set holds %d of %d entries", count, params.MaxValidators)
	}

	power := msg.Power
	if power == 0 {
		power = params.DefaultPower
	}

	if power <= 0 {
		return nil, errorsmod.Wrapf(types.ErrInvalidPower, "power must be positive, got %d", power)
	}

	validator := types.Validator{
		OperatorAddress: msg.OperatorAddress,
		ConsensusPubkey: msg.ConsensusPubkey,
		Moniker:         msg.Moniker,
		Power:           power,
	}

	if err := validator.Validate(); err != nil {
		return nil, err
	}

	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return nil, err
	}

	inUse, err := k.IsPubKeyInUse(ctx, pubKey.Bytes())
	if err != nil {
		return nil, err
	}

	if inUse {
		return nil, errorsmod.Wrapf(types.ErrDuplicatePubKey, "operator %s", msg.OperatorAddress)
	}

	if err := k.SetValidator(ctx, validator); err != nil {
		return nil, err
	}

	if err := k.QueueValidatorUpdate(ctx, validator.OperatorAddress, pubKey, power); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddValidator,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, validator.OperatorAddress),
			sdk.NewAttribute(types.AttributeKeyMoniker, validator.Moniker),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", validator.Power)),
		),
	)

	return &types.MsgAddValidatorResponse{}, nil
}

func (k msgServer) RemoveValidator(ctx context.Context, msg *types.MsgRemoveValidator) (*types.MsgRemoveValidatorResponse, error) {
	authority, err := k.GetAuthority(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoAuthority, err.Error())
	}

	if authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s", authority, msg.Authority)
	}

	if _, err := k.addressCodec.StringToBytes(msg.OperatorAddress); err != nil {
		return nil, errorsmod.Wrapf(err, "invalid operator address %s", msg.OperatorAddress)
	}

	validator, err := k.GetValidator(ctx, msg.OperatorAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrValidatorNotFound, "operator %s", msg.OperatorAddress)
	}

	count, err := k.CountValidators(ctx)
	if err != nil {
		return nil, err
	}

	if count <= 1 {
		return nil, errorsmod.Wrap(types.ErrEmptyValidators,
			"cannot remove the last validator of the chain")
	}

	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.RemoveValidator(ctx, msg.OperatorAddress); err != nil {
		return nil, err
	}

	if err := k.QueueValidatorUpdate(ctx, msg.OperatorAddress, pubKey, 0); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveValidator,
			sdk.NewAttribute(types.AttributeKeyOperatorAddress, msg.OperatorAddress),
		),
	)

	return &types.MsgRemoveValidatorResponse{}, nil
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

	count, err := k.CountValidators(ctx)
	if err != nil {
		return nil, err
	}

	if count > msg.Params.MaxValidators {
		return nil, errorsmod.Wrapf(types.ErrMaxValidators,
			"cannot lower max_validators to %d while %d validators are active",
			msg.Params.MaxValidators, count)
	}

	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeUpdateParams),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}
