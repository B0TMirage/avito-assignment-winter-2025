package models

type Merch struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Price int    `json:"price"`
}

type MerchInventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}
