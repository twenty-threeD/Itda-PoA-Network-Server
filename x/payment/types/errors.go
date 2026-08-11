package types

import "cosmossdk.io/errors"

var (
	ErrDuplicateOrderID = errors.Register(ModuleName, 2, "payment already recorded for order id")
	ErrRecordNotFound   = errors.Register(ModuleName, 3, "payment record not found")
	ErrInvalidOrderID   = errors.Register(ModuleName, 4, "invalid order id")
	ErrInvalidAmount    = errors.Register(ModuleName, 5, "invalid payment amount")
	ErrInvalidPaidAt    = errors.Register(ModuleName, 6, "invalid paid_at timestamp")
	ErrInvalidRecord    = errors.Register(ModuleName, 7, "invalid payment record")
	ErrNoAuthority      = errors.Register(ModuleName, 8, "module authority is not set")
)
