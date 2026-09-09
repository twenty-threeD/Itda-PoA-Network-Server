package app

import (
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type HandlerOptions struct {
	AccountKeeper   authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	SignModeHandler *txsigning.HandlerMap
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	return ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   options.AccountKeeper,
		BankKeeper:      options.BankKeeper,
		SignModeHandler: options.SignModeHandler,
		FeegrantKeeper:  nil,
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
}
