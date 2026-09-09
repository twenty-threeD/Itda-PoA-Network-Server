package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/keeper"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/testutil"
	"github.com/twenty-threeD/Itda-PoA-Network-Server/x/poa/types"
)

func TestAddValidator(t *testing.T) {
	operator := testutil.NewAccAddress("operator-one")

	t.Run("adds a validator and queues a power update", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		pubKey := testutil.NewPubKey()
		pkAny, err := codectypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		// When
		_, err = msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.Authority,
			OperatorAddress: operator,
			ConsensusPubkey: pkAny,
			Moniker:         "node-one",
			Power:           0,
		})

		// Then
		require.NoError(t, err)

		stored, err := k.GetValidator(ctx, operator)
		require.NoError(t, err)
		require.Equal(t, types.DefaultPower, stored.Power, "zero power must fall back to default_power")

		updates, err := k.DrainValidatorUpdates(ctx)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.Equal(t, types.DefaultPower, updates[0].Power)
	})

	t.Run("rejects a foreign authority", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		pkAny, err := codectypes.NewAnyWithValue(testutil.NewPubKey())
		require.NoError(t, err)

		// When
		_, err = msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.NewAccAddress("intruder"),
			OperatorAddress: operator,
			ConsensusPubkey: pkAny,
			Power:           1,
		})

		// Then
		require.ErrorContains(t, err, "invalid authority")

		_, err = k.GetValidator(ctx, operator)
		require.Error(t, err, "state must be untouched when the authority check fails")
	})

	t.Run("rejects a duplicate operator", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, operator, 1)))

		pkAny, err := codectypes.NewAnyWithValue(testutil.NewPubKey())
		require.NoError(t, err)

		// When
		_, err = msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.Authority,
			OperatorAddress: operator,
			ConsensusPubkey: pkAny,
			Power:           1,
		})

		// Then
		require.ErrorIs(t, err, types.ErrValidatorExists)
	})

	t.Run("rejects a duplicate consensus pubkey", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)

		existing := testutil.NewValidator(t, operator, 1)
		require.NoError(t, k.SetValidator(ctx, existing))

		// When
		_, err := msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.Authority,
			OperatorAddress: testutil.NewAccAddress("operator-two"),
			ConsensusPubkey: existing.ConsensusPubkey,
			Power:           1,
		})

		// Then
		require.ErrorIs(t, err, types.ErrDuplicatePubKey)
	})

	t.Run("rejects an addition past max_validators", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetParams(ctx, types.Params{MaxValidators: 1, DefaultPower: 1}))
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, operator, 1)))

		pkAny, err := codectypes.NewAnyWithValue(testutil.NewPubKey())
		require.NoError(t, err)

		// When
		_, err = msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.Authority,
			OperatorAddress: testutil.NewAccAddress("operator-two"),
			ConsensusPubkey: pkAny,
			Power:           1,
		})

		// Then
		require.ErrorIs(t, err, types.ErrMaxValidators)
	})

	t.Run("rejects a negative power", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		pkAny, err := codectypes.NewAnyWithValue(testutil.NewPubKey())
		require.NoError(t, err)

		// When
		_, err = msgServer.AddValidator(ctx, &types.MsgAddValidator{
			Authority:       testutil.Authority,
			OperatorAddress: operator,
			ConsensusPubkey: pkAny,
			Power:           -1,
		})

		// Then
		require.ErrorIs(t, err, types.ErrInvalidPower)
	})
}

func TestRemoveValidator(t *testing.T) {
	first := testutil.NewAccAddress("operator-one")
	second := testutil.NewAccAddress("operator-two")

	t.Run("removes a validator and queues a zero power update", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, first, 1)))
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, second, 1)))

		// When
		_, err := msgServer.RemoveValidator(ctx, types.NewMsgRemoveValidator(testutil.Authority, second))

		// Then
		require.NoError(t, err)

		_, err = k.GetValidator(ctx, second)
		require.Error(t, err)

		updates, err := k.DrainValidatorUpdates(ctx)
		require.NoError(t, err)
		require.Len(t, updates, 1)
		require.Zero(t, updates[0].Power, "removal must be signalled to consensus as zero power")
	})

	t.Run("rejects removing the last validator", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, first, 1)))

		// When
		_, err := msgServer.RemoveValidator(ctx, types.NewMsgRemoveValidator(testutil.Authority, first))

		// Then
		require.ErrorIs(t, err, types.ErrEmptyValidators)
	})

	t.Run("rejects a foreign authority", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, first, 1)))
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, second, 1)))

		// When
		_, err := msgServer.RemoveValidator(ctx, types.NewMsgRemoveValidator(testutil.NewAccAddress("intruder"), second))

		// Then
		require.ErrorContains(t, err, "invalid authority")

		_, err = k.GetValidator(ctx, second)
		require.NoError(t, err, "state must be untouched when the authority check fails")
	})

	t.Run("rejects an unknown validator", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, first, 1)))

		// When
		_, err := msgServer.RemoveValidator(ctx, types.NewMsgRemoveValidator(testutil.Authority, second))

		// Then
		require.ErrorIs(t, err, types.ErrValidatorNotFound)
	})
}

func TestUpdateParams(t *testing.T) {
	t.Run("rejects lowering max_validators below the active count", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, testutil.NewAccAddress("operator-one"), 1)))
		require.NoError(t, k.SetValidator(ctx, testutil.NewValidator(t, testutil.NewAccAddress("operator-two"), 1)))

		// When
		_, err := msgServer.UpdateParams(ctx, types.NewMsgUpdateParams(
			testutil.Authority,
			types.Params{MaxValidators: 1, DefaultPower: 1},
		))

		// Then
		require.ErrorIs(t, err, types.ErrMaxValidators)
	})

	t.Run("rejects a zero default_power", func(t *testing.T) {
		// Given
		k, ctx := testutil.NewKeeper(t)
		msgServer := keeper.NewMsgServerImpl(k)

		// When
		_, err := msgServer.UpdateParams(ctx, types.NewMsgUpdateParams(
			testutil.Authority,
			types.Params{MaxValidators: 5, DefaultPower: 0},
		))

		// Then
		require.ErrorIs(t, err, types.ErrInvalidParams)
	})
}
