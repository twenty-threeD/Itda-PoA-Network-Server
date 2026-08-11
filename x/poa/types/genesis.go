package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ codectypes.UnpackInterfacesMessage = (*GenesisState)(nil)

func NewGenesisState(params Params, validators []Validator, authority string) *GenesisState {
	return &GenesisState{
		Params:     params,
		Validators: validators,
		Authority:  authority,
	}
}

func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []Validator{}, "")
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.Authority == "" {
		return errorsmod.Wrap(ErrNoAuthority, "genesis authority cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(gs.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	if len(gs.Validators) == 0 {
		return errorsmod.Wrap(ErrEmptyValidators, "genesis must contain at least one validator")
	}

	if uint32(len(gs.Validators)) > gs.Params.MaxValidators {
		return errorsmod.Wrapf(ErrMaxValidators,
			"genesis has %d validators, max_validators is %d",
			len(gs.Validators), gs.Params.MaxValidators)
	}

	seenOperators := make(map[string]struct{}, len(gs.Validators))
	seenPubKeys := make(map[string]struct{}, len(gs.Validators))

	for i := range gs.Validators {
		validator := gs.Validators[i]

		if err := validator.Validate(); err != nil {
			return errorsmod.Wrapf(err, "validator at index %d", i)
		}

		if _, exists := seenOperators[validator.OperatorAddress]; exists {
			return errorsmod.Wrapf(ErrValidatorExists,
				"duplicate operator address %s", validator.OperatorAddress)
		}

		seenOperators[validator.OperatorAddress] = struct{}{}

		pubKey, err := validator.ConsPubKey()
		if err != nil {
			return errorsmod.Wrapf(err, "validator at index %d", i)
		}

		pubKeyID := string(pubKey.Bytes())
		if _, exists := seenPubKeys[pubKeyID]; exists {
			return errorsmod.Wrapf(ErrDuplicatePubKey,
				"duplicate consensus pubkey for operator %s", validator.OperatorAddress)
		}

		seenPubKeys[pubKeyID] = struct{}{}
	}

	return nil
}

func (gs GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range gs.Validators {
		if err := gs.Validators[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	return nil
}
