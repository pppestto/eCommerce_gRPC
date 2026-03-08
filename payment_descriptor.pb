
¯
payment/v1/payment.proto
payment.v1common/common.proto"w

PayRequest
order_id (	RorderId%
amount (2.common.MoneyRamount'
idempotency_key (	RidempotencyKey"s
PayResponse
success (Rsuccess%
transaction_id (	RtransactionId#
error_message (	RerrorMessage2H
PaymentService6
Pay.payment.v1.PayRequest.payment.v1.PayResponseB2Z0github.com/pppestto/ecommerce-grpc/pb/payment/v1bproto3