package utils

type WebError struct {
	Message string `json:"message"`
}

func NewWebError(message string) WebError {
	return WebError{
		Message: message,
	}
}
