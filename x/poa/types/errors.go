package types

import "cosmossdk.io/errors"

var (
	ErrValidatorExists   = errors.Register(ModuleName, 2, "validator already exists")
	ErrValidatorNotFound = errors.Register(ModuleName, 3, "validator not found")
	ErrMaxValidators     = errors.Register(ModuleName, 4, "validator set is full")
	ErrEmptyValidators   = errors.Register(ModuleName, 5, "validator set cannot be empty")
	ErrInvalidPubKey     = errors.Register(ModuleName, 6, "invalid consensus pubkey")
	ErrDuplicatePubKey   = errors.Register(ModuleName, 7, "consensus pubkey already in use")
	ErrInvalidPower      = errors.Register(ModuleName, 8, "invalid voting power")
	ErrInvalidParams     = errors.Register(ModuleName, 9, "invalid params")
	ErrNoAuthority       = errors.Register(ModuleName, 10, "module authority is not set")
)
