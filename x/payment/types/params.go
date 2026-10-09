package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func DefaultParams() Params {
	return Params{}
}

func (p Params) Validate() error {
	seen := make(map[string]struct{}, len(p.Recorders))

	for i, recorder := range p.Recorders {
		if _, err := sdk.AccAddressFromBech32(recorder); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid recorder address at index %d: %s", i, err)
		}
		if _, exists := seen[recorder]; exists {
			return errorsmod.Wrapf(ErrDuplicateRecorder, "recorder %s", recorder)
		}

		seen[recorder] = struct{}{}
	}

	return nil
}

// IsRecorder reports whether addr is in the recorder allow-list.
func (p Params) IsRecorder(addr string) bool {
	for _, recorder := range p.Recorders {
		if recorder == addr {
			return true
		}
	}

	return false
}
