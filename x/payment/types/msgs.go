package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgRecordPayment)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)
)

func NewMsgRecordPayment(authority, orderID, buyerAddress string, amount int64, paidAt, contractURL, paymentHash, buyerSignature string) *MsgRecordPayment {
	return &MsgRecordPayment{
		Authority:      authority,
		OrderId:        orderID,
		BuyerAddress:   buyerAddress,
		Amount:         amount,
		PaidAt:         paidAt,
		ContractUrl:    contractURL,
		PaymentHash:    paymentHash,
		BuyerSignature: buyerSignature,
	}
}

func (m MsgRecordPayment) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	return m.ToRecord(0).Validate()
}

func (m MsgRecordPayment) ToRecord(recordedHeight int64) PaymentRecord {
	return PaymentRecord{
		OrderId:        m.OrderId,
		BuyerAddress:   m.BuyerAddress,
		Amount:         m.Amount,
		PaidAt:         m.PaidAt,
		ContractUrl:    m.ContractUrl,
		PaymentHash:    m.PaymentHash,
		BuyerSignature: m.BuyerSignature,
		RecordedHeight: recordedHeight,
	}
}

func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	return m.Params.Validate()
}
