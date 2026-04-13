package domain

type Product struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Price    uint64 `json:"price"`
	Quantity uint64 `json:"quanity"`
}
