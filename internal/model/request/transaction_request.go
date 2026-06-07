package request

// TransactionRequest is the input to the Transaction controller. UserID and
// SourceType are populated from the path and headers, not the JSON body.
type TransactionRequest struct {
	UserID        uint64 `json:"-"`
	SourceType    string `json:"-"`
	State         string `json:"state"`
	Amount        string `json:"amount"`
	TransactionID string `json:"transactionId"`
}
