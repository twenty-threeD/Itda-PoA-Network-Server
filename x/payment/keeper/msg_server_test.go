package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/keeper"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/testutil"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/payment/types"
)

func TestRecordPayment(t *testing.T) {
	t.Run("records a payment and stamps the block height", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		msg := testutil.NewRecordMsg()

		// When
		res, err := msgServer.RecordPayment(ctx, msg)

		// Then
		require.NoError(t, err)
		require.EqualValues(t, testutil.BlockHeight, res.RecordedHeight)

		stored, err := k.GetRecord(ctx, msg.OrderId)
		require.NoError(t, err)
		require.Equal(t, msg.Amount, stored.Amount)
		require.Equal(t, msg.PaidAt, stored.PaidAt)
		require.Equal(t, msg.ContractUrl, stored.ContractUrl)
		require.EqualValues(t, testutil.BlockHeight, stored.RecordedHeight,
			"recorded_height must come from the block, never from the submitter")
	})

	t.Run("rejects a duplicate order id", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)

		_, err := msgServer.RecordPayment(ctx, testutil.NewRecordMsg())
		require.NoError(t, err)

		// When
		_, err = msgServer.RecordPayment(ctx, testutil.NewRecordMsg(func(m *types.MsgRecordPayment) {
			m.Amount = 1
		}))

		// Then
		require.ErrorIs(t, err, types.ErrDuplicateOrderID)

		stored, err := k.GetRecord(ctx, testutil.NewRecordMsg().OrderId)
		require.NoError(t, err)
		require.EqualValues(t, 55000, stored.Amount, "the original record must survive a replay")
	})

	t.Run("rejects a foreign authority", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		msg := testutil.NewRecordMsg(func(m *types.MsgRecordPayment) {
			m.Authority = testutil.NewAccAddress("intruder")
		})

		// When
		_, err := msgServer.RecordPayment(ctx, msg)

		// Then
		require.ErrorContains(t, err, "invalid authority")

		_, err = k.GetRecord(ctx, msg.OrderId)
		require.Error(t, err, "state must be untouched when the authority check fails")
	})

	testCases := []struct {
		name      string
		mutate    func(*types.MsgRecordPayment)
		expectErr error
	}{
		{
			name:      "rejects an empty order id",
			mutate:    func(m *types.MsgRecordPayment) { m.OrderId = "" },
			expectErr: types.ErrInvalidOrderID,
		},
		{
			name:      "rejects a zero amount",
			mutate:    func(m *types.MsgRecordPayment) { m.Amount = 0 },
			expectErr: types.ErrInvalidAmount,
		},
		{
			name:      "rejects a negative amount",
			mutate:    func(m *types.MsgRecordPayment) { m.Amount = -1 },
			expectErr: types.ErrInvalidAmount,
		},
		{
			name:      "rejects a non rfc3339 paid_at",
			mutate:    func(m *types.MsgRecordPayment) { m.PaidAt = "2026-08-11 15:04:05" },
			expectErr: types.ErrInvalidPaidAt,
		},
		{
			name:      "rejects an empty contract url",
			mutate:    func(m *types.MsgRecordPayment) { m.ContractUrl = "" },
			expectErr: types.ErrInvalidRecord,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			k, ctx := testutil.NewKeeper(t)
			msgServer := keeper.NewMsgServerImpl(k)
			msg := testutil.NewRecordMsg(tc.mutate)

			// When
			_, err := msgServer.RecordPayment(ctx, msg)

			// Then
			require.ErrorIs(t, err, tc.expectErr)

			has, err := k.HasRecord(ctx, msg.OrderId)
			require.NoError(t, err)
			require.False(t, has)
		})
	}

	t.Run("rejects an invalid buyer address", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		msg := testutil.NewRecordMsg(func(m *types.MsgRecordPayment) {
			m.BuyerAddress = "not-a-bech32-address"
		})

		// When
		_, err := msgServer.RecordPayment(ctx, msg)

		// Then
		require.Error(t, err)

		has, err := k.HasRecord(ctx, msg.OrderId)
		require.NoError(t, err)
		require.False(t, has)
	})
}

func TestGenesisRoundTrip(t *testing.T) {
	// Given
	k, ctx := testutil.NewKeeper(t)
	msgServer := keeper.NewMsgServerImpl(k)

	_, err := msgServer.RecordPayment(ctx, testutil.NewRecordMsg())
	require.NoError(t, err)

	// When
	exported, err := k.ExportGenesis(ctx)
	require.NoError(t, err)

	// Then
	require.NoError(t, exported.Validate())
	require.Equal(t, testutil.Authority, exported.Authority)
	require.Len(t, exported.Records, 1)

	fresh, freshCtx := testutil.NewKeeper(t)
	require.NoError(t, fresh.InitGenesis(freshCtx, exported))

	reimported, err := fresh.GetRecord(freshCtx, exported.Records[0].OrderId)
	require.NoError(t, err)
	require.Equal(t, exported.Records[0], reimported)
}

func TestRecordPaymentByRecorder(t *testing.T) {
	recorder := testutil.NewAccAddress("recorder")

	t.Run("accepts a registered recorder as submitter", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: testutil.Authority,
			Params:    types.Params{Recorders: []string{recorder}},
		})
		require.NoError(t, err)

		msg := testutil.NewRecordMsg(func(m *types.MsgRecordPayment) { m.Authority = recorder })

		// When
		_, err = msgServer.RecordPayment(ctx, msg)

		// Then
		require.NoError(t, err)

		has, err := k.HasRecord(ctx, msg.OrderId)
		require.NoError(t, err)
		require.True(t, has)
	})

	t.Run("rejects a recorder after it is removed from params", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetParams(ctx, types.Params{Recorders: []string{recorder}}))
		require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

		msg := testutil.NewRecordMsg(func(m *types.MsgRecordPayment) { m.Authority = recorder })

		// When
		_, err := msgServer.RecordPayment(ctx, msg)

		// Then
		require.ErrorContains(t, err, "invalid authority")
	})

	t.Run("keeps accepting the authority when recorders are set", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetParams(ctx, types.Params{Recorders: []string{recorder}}))

		// When
		_, err := msgServer.RecordPayment(ctx, testutil.NewRecordMsg())

		// Then
		require.NoError(t, err)
	})
}

func TestUpdateParams(t *testing.T) {
	recorder := testutil.NewAccAddress("recorder")

	t.Run("rejects a recorder trying to change the recorder list", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetParams(ctx, types.Params{Recorders: []string{recorder}}))

		// When
		_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: recorder,
			Params:    types.Params{Recorders: []string{recorder, testutil.NewAccAddress("intruder")}},
		})

		// Then
		require.ErrorContains(t, err, "invalid authority")

		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{recorder}, params.Recorders)
	})

	testCases := []struct {
		name      string
		recorders []string
		expectErr error
	}{
		{
			name:      "rejects a duplicate recorder",
			recorders: []string{recorder, recorder},
			expectErr: types.ErrDuplicateRecorder,
		},
		{
			name:      "rejects a non bech32 recorder",
			recorders: []string{"not-a-bech32-address"},
			expectErr: sdkerrors.ErrInvalidAddress,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			k, ctx := testutil.NewKeeper(t)
			msgServer := keeper.NewMsgServerImpl(k)

			// When
			_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
				Authority: testutil.Authority,
				Params:    types.Params{Recorders: tc.recorders},
			})

			// Then
			require.ErrorIs(t, err, tc.expectErr)

			params, err := k.GetParams(ctx)
			require.NoError(t, err)
			require.Empty(t, params.Recorders)
		})
	}
}
