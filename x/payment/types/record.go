package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxOrderIDLength    = 128
	MaxContractURLength = 2048
	MaxHashLength       = 128
	MaxSignatureLength  = 256
)

func (r PaymentRecord) Validate() error {
	if err := ValidateOrderID(r.OrderId); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(r.BuyerAddress); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid buyer address: %s", err)
	}

	if r.Amount <= 0 {
		return errorsmod.Wrapf(ErrInvalidAmount, "amount must be positive, got %d", r.Amount)
	}

	if err := ValidatePaidAt(r.PaidAt); err != nil {
		return err
	}

	if r.ContractUrl == "" {
		return errorsmod.Wrap(ErrInvalidRecord, "contract_url cannot be empty")
	}

	if len(r.ContractUrl) > MaxContractURLength {
		return errorsmod.Wrapf(ErrInvalidRecord,
			"contract_url exceeds %d characters", MaxContractURLength)
	}

	if len(r.PaymentHash) > MaxHashLength {
		return errorsmod.Wrapf(ErrInvalidRecord,
			"payment_hash exceeds %d characters", MaxHashLength)
	}

	if len(r.BuyerSignature) > MaxSignatureLength {
		return errorsmod.Wrapf(ErrInvalidRecord,
			"buyer_signature exceeds %d characters", MaxSignatureLength)
	}

	if r.RecordedHeight < 0 {
		return errorsmod.Wrapf(ErrInvalidRecord,
			"recorded_height cannot be negative, got %d", r.RecordedHeight)
	}

	return nil
}

func ValidateOrderID(orderID string) error {
	if orderID == "" {
		return errorsmod.Wrap(ErrInvalidOrderID, "order_id cannot be empty")
	}

	if len(orderID) > MaxOrderIDLength {
		return errorsmod.Wrapf(ErrInvalidOrderID,
			"order_id exceeds %d characters", MaxOrderIDLength)
	}

	return nil
}

func ValidatePaidAt(paidAt string) error {
	if paidAt == "" {
		return errorsmod.Wrap(ErrInvalidPaidAt, "paid_at cannot be empty")
	}

	if _, err := time.Parse(time.RFC3339, paidAt); err != nil {
		return errorsmod.Wrapf(ErrInvalidPaidAt, "paid_at must be RFC 3339: %s", err)
	}

	return nil
}
