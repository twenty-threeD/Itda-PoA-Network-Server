package types

import "cosmossdk.io/collections"

const (
	ModuleName = "payment"
	StoreKey   = ModuleName
)

var (
	ParamsKey        = collections.NewPrefix(0)
	RecordsKeyPrefix = collections.NewPrefix(1)
	AuthorityKey     = collections.NewPrefix(2)
)
