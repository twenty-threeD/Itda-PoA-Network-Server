package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgAddValidator)(nil)
	_ sdk.Msg = (*MsgRemoveValidator)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)

	_ codectypes.UnpackInterfacesMessage = (*MsgAddValidator)(nil)
)

func NewMsgAddValidator(authority, operatorAddress string, pubKey cryptotypes.PubKey, moniker string, power int64) (*MsgAddValidator, error) {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return nil, errorsmod.Wrap(ErrInvalidPubKey, err.Error())
	}

	return &MsgAddValidator{
		Authority:       authority,
		OperatorAddress: operatorAddress,
		ConsensusPubkey: pkAny,
		Moniker:         moniker,
		Power:           power,
	}, nil
}

func (m MsgAddValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.OperatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address: %s", err)
	}

	if m.ConsensusPubkey == nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "consensus pubkey cannot be nil")
	}

	if _, ok := m.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey); !ok {
		return errorsmod.Wrapf(ErrInvalidPubKey,
			"expected cryptotypes.PubKey, got %T", m.ConsensusPubkey.GetCachedValue())
	}

	if m.Power < 0 {
		return errorsmod.Wrapf(ErrInvalidPower, "power cannot be negative, got %d", m.Power)
	}

	return nil
}

func (m MsgAddValidator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey

	return unpacker.UnpackAny(m.ConsensusPubkey, &pk)
}

func NewMsgRemoveValidator(authority, operatorAddress string) *MsgRemoveValidator {
	return &MsgRemoveValidator{
		Authority:       authority,
		OperatorAddress: operatorAddress,
	}
}

func (m MsgRemoveValidator) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.OperatorAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address: %s", err)
	}

	return nil
}

func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	return m.Params.Validate()
}
