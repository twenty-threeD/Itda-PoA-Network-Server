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
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/keeper"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

const (
	Bech32Prefix = "cosmos"
	BlockHeight  = 42
)

// Authority is the fixed submitter address used across the payment tests. It
// stands in for the application server's on-chain account.
var Authority = NewAccAddress("authority")

// NewKeeper spins up a payment keeper backed by an in-memory store, seeded
// with default params and Authority as the module authority.
func NewKeeper(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		authcodec.NewBech32Codec(Bech32Prefix),
		log.NewNopLogger(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: BlockHeight}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	require.NoError(t, k.SetAuthority(ctx, Authority))

	return k, ctx
}

// NewRecordMsg builds a valid MsgRecordPayment that individual test cases
// mutate to exercise a single failure mode at a time.
func NewRecordMsg(opts ...func(*types.MsgRecordPayment)) *types.MsgRecordPayment {
	msg := &types.MsgRecordPayment{
		Authority:      Authority,
		OrderId:        "ORDER-1001",
		BuyerAddress:   NewAccAddress("buyer"),
		Amount:         55000,
		PaidAt:         "2026-08-11T15:04:05+09:00",
		ContractUrl:    "https://itda.example/contracts/1001.pdf",
		PaymentHash:    "0f9b1c2d3e4f5061728394a5b6c7d8e9",
		BuyerSignature: "c2lnbmF0dXJl",
	}

	for _, opt := range opts {
		opt(msg)
	}

	return msg
}

// NewAccAddress returns a deterministic bech32 account address for the given
// seed, so test cases stay reproducible.
func NewAccAddress(seed string) string {
	padded := make([]byte, 20)
	copy(padded, seed)

	return sdk.AccAddress(padded).String()
}
