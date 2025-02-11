package models

type User struct {
	ID             int     `json:"id"`
	Username       string  `json:"username"`
	Password       string  `json:"password"`
	Coins          uint    `json:"coins"`
	MerchPurchased []Merch `json:"merch"`
}
