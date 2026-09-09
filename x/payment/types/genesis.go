package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params Params, authority string, records []PaymentRecord) *GenesisState {
	return &GenesisState{
		Params:    params,
		Authority: authority,
		Records:   records,
	}
}

func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), "", []PaymentRecord{})
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

	seenOrderIDs := make(map[string]struct{}, len(gs.Records))

	for i := range gs.Records {
		record := gs.Records[i]
		if err := record.Validate(); err != nil {
			return errorsmod.Wrapf(err, "record at index %d", i)
		}
		if _, exists := seenOrderIDs[record.OrderId]; exists {
			return errorsmod.Wrapf(ErrDuplicateOrderID, "order id %s", record.OrderId)
		}

		seenOrderIDs[record.OrderId] = struct{}{}
	}

	return nil
}
