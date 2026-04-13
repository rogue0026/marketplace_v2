package domain

type Product struct {
	ID       uint64
	Name     string
	Price    uint64
	Quantity uint64
}
type BasketItem struct {
	ID              uint64
	UserID          uint64
	ProductID       uint64
	ProductQuantity uint64
}

type BasketItemAggregated struct {
	ProductID               uint64
	ProductName             string
	ProductPrice            uint64
	ProductQuantityInBasket uint64
}
