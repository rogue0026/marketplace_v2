package domain

var (
	StatusWaitingForPayment = "waiting for payment"
	StatusProcessing        = "processing"
	StatusPayedSuccessfully = "payed successfully"
)

type OrderItem struct {
	ProductID           uint64
	ProductQuantity     uint64
	ProductPricePerUnit uint64
}
