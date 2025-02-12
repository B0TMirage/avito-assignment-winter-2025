package models

type Transaction struct {
	Received []Receive `json:"received"`
	Sent     []Sent    `json:"sent"`
}

type Receive struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type Sent struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
