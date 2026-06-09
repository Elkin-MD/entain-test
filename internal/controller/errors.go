package controller

// RequestError is a malformed-request error that maps to 400 Bad Request.
type RequestError string

func (e RequestError) Error() string {
	return string(e)
}

const (
	ErrInvalidUserID RequestError = "userId must be a positive integer"
	ErrInvalidBody   RequestError = "invalid JSON body"
)
