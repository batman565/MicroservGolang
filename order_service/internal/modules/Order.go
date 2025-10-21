package modules

type Order struct {
	Id        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Order     string `json:"order"`
	Count     int    `json:"count"`
	Status    string `json:"status"`
	Price     int    `json:"price"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
