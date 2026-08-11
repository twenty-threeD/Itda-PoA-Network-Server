package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

var _ codectypes.UnpackInterfacesMessage = (*Validator)(nil)

func NewValidator(operatorAddress string, pubKey cryptotypes.PubKey, moniker string, power int64) (Validator, error) {
	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	if err != nil {
		return Validator{}, errorsmod.Wrap(ErrInvalidPubKey, err.Error())
	}

	return Validator{
		OperatorAddress: operatorAddress,
		ConsensusPubkey: pkAny,
		Moniker:         moniker,
		Power:           power,
	}, nil
}

func (v Validator) Validate() error {
	if v.OperatorAddress == "" {
		return errorsmod.Wrap(ErrValidatorNotFound, "operator address cannot be empty")
	}

	if v.ConsensusPubkey == nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "consensus pubkey cannot be nil")
	}

	if v.Power <= 0 {
		return errorsmod.Wrapf(ErrInvalidPower, "power must be positive, got %d", v.Power)
	}

	return nil
}

func (v Validator) ConsPubKey() (cryptotypes.PubKey, error) {
	if v.ConsensusPubkey == nil {
		return nil, errorsmod.Wrap(ErrInvalidPubKey, "consensus pubkey cannot be nil")
	}

	pk, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, errorsmod.Wrapf(ErrInvalidPubKey,
			"expected cryptotypes.PubKey, got %T", v.ConsensusPubkey.GetCachedValue())
	}

	return pk, nil
}

func (v Validator) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pk cryptotypes.PubKey

	return unpacker.UnpackAny(v.ConsensusPubkey, &pk)
}
