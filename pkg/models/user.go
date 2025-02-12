package models

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Coins    int    `json:"coins"`
}

type UserInfoResponse struct {
	Coins       int                  `json:"coins"`
	Inventory   []MerchInventoryItem `json:"inventory"`
	CoinHistory Transaction          `json:"coinHistory"`
}
