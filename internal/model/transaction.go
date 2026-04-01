package model

type Transaction struct {
	FromKey string  `json:"from_key"`
	ToKey   string  `json:"to_key"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}
