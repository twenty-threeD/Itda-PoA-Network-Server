package payment

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "itda.payment.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the module parameters",
				},
				{
					RpcMethod:      "Payment",
					Use:            "payment [order-id]",
					Short:          "Query a payment record by order id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "order_id"}},
				},
				{
					RpcMethod: "Payments",
					Use:       "payments",
					Short:     "List all payment records",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "itda.payment.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RecordPayment",
					Use:       "record-payment [order-id] [buyer-address] [amount] [paid-at] [contract-url] [payment-hash] [buyer-signature]",
					Short:     "Record a confirmed payment on-chain",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "order_id"},
						{ProtoField: "buyer_address"},
						{ProtoField: "amount"},
						{ProtoField: "paid_at"},
						{ProtoField: "contract_url"},
						{ProtoField: "payment_hash"},
						{ProtoField: "buyer_signature"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Skip:      true,
				},
			},
		},
	}
}
