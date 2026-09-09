package types

import errorsmod "cosmossdk.io/errors"

const (
	DefaultMaxValidators uint32 = 21
	DefaultPower         int64  = 1
)

func DefaultParams() Params {
	return Params{
		MaxValidators: DefaultMaxValidators,
		DefaultPower:  DefaultPower,
	}
}

func (p Params) Validate() error {
	if p.MaxValidators == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "max_validators must be positive")
	}

	if p.DefaultPower <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "default_power must be positive")
	}

	return nil
}
