package response

// BalanceResponse is the body returned by the Balance controller.
type BalanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"`
}
