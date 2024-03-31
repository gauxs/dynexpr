package data

// dynexpr:generate
type Transaction struct {
	UserID        *string `json:"user_id,omitempty" dynexpr:"partitionKey"`
	TransactionID *string `json:"transaction_id,omitempty"  dynexpr:"sortKey"`
	Amount        *int    `json:"amount,omitempty"`
}
