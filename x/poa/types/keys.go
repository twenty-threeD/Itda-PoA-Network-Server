package types

import "cosmossdk.io/collections"

const (
	ModuleName = "poa"
	StoreKey   = ModuleName
)

var (
	ParamsKey             = collections.NewPrefix(0)
	ValidatorsKeyPrefix   = collections.NewPrefix(1)
	PowerUpdatesKeyPrefix = collections.NewPrefix(2)
	AuthorityKey          = collections.NewPrefix(3)
)
