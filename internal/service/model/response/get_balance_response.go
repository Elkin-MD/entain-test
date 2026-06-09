package response

// GetBalanceResponse is the service-level result of a balance lookup.
type GetBalanceResponse struct {
	UserID  uint64
	Balance int64
}
