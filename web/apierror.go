package web

import "net/http"

// APIError is to define REST API errors.
type APIError struct {
	Status  int
	Message string
	Err     error
}

// Error implements error interface.
func (e APIError) Error() string {
	if e.Err == nil {
		return e.Message
	}

	return e.Err.Error() + ": " + e.Message
}

// InternalServerError creates an APIError.
func InternalServerError(e error) APIError {
	return APIError{
		http.StatusInternalServerError,
		http.StatusText(http.StatusInternalServerError),
		e,
	}
}

func BadRequest(reason string) APIError {
	return APIError{http.StatusBadRequest, "invalid request: " + reason, nil}
}

// Common API errors
var (
	APIErrNotFound   = APIError{http.StatusNotFound, "requested resource is not found", nil}
	APIErrBadRequest = APIError{http.StatusBadRequest, "invalid request", nil}
	APIErrConflict   = APIError{http.StatusConflict, "conflicted", nil}
)
