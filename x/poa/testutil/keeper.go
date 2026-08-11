package testutil

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/keeper"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

const Bech32Prefix = "cosmos"

// Authority is the fixed authority address used across the poa keeper tests.
var Authority = sdk.AccAddress([]byte("authority___________")).String()

// NewKeeper spins up a poa keeper backed by an in-memory store and returns it
// together with a context. The module authority is seeded to Authority.
func NewKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authcodec.NewBech32Codec(Bech32Prefix),
		log.NewNopLogger(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 1}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	require.NoError(t, k.SetAuthority(ctx, Authority))

	return k, ctx
}

// NewValidator builds a validator fixture with a fresh consensus key.
func NewValidator(t *testing.T, operatorAddress string, power int64) types.Validator {
	t.Helper()

	validator, err := types.NewValidator(operatorAddress, NewPubKey(), "moniker", power)
	require.NoError(t, err)

	return validator
}

// NewPubKey returns a fresh ed25519 consensus pubkey.
func NewPubKey() cryptotypes.PubKey {
	return ed25519.GenPrivKey().PubKey()
}

// NewAccAddress returns a deterministic bech32 account address for the given
// seed, so test cases stay reproducible.
func NewAccAddress(seed string) string {
	padded := make([]byte, 20)
	copy(padded, seed)

	return sdk.AccAddress(padded).String()
}
