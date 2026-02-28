package model

type RechargeOrder struct {
	Amount        float32 `json:"amount"`
	ProductId     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	UserId        string  `json:"user_id"`
	OrderId       string  `json:"order_id"`
	GameUserId    string  `json:"game_user_id"`
	ServerId      string  `json:"server_id"`
	PaymentTime   string  `json:"payment_time"`
	ChannelNumber string  `json:"channel_number"`
}
