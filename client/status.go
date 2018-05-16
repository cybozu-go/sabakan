package client

import (
	"fmt"
	"net/http"

	"github.com/google/subcommands"
)

const (
	// ExitSuccess represents no error.
	ExitSuccess subcommands.ExitStatus = 0
	// ExitError represents general error.
	ExitError = 1
	// ExitBadUsage represents bad usage of command.
	ExitBadUsage = 2
	// ExitInvalidParams represents invalid input parameters for command.
	ExitInvalidParams = 3
	// ExitResponse4xx represents HTTP status 4xx.
	ExitResponse4xx = 4
	// ExitResponse5xx represents HTTP status 5xx.
	ExitResponse5xx = 5
	// ExitNotFound represents HTTP status 404.
	ExitNotFound = 14
	// ExitConflicted represents HTTP status 409.
	ExitConflicted = 19
)

// Status implements error interface as combination of exit code and native error.
type Status struct {
	code subcommands.ExitStatus
	err  error
}

// Code returns exit code.
func (s *Status) Code() subcommands.ExitStatus {
	return s.code
}

// Error executes native error's Error().
func (s *Status) Error() string {
	return s.err.Error()
}

// NewStatus creates new Status with exit code and native error.
func NewStatus(code subcommands.ExitStatus, err error) *Status {
	return &Status{code: code, err: err}
}

// ErrorStatus creates new Status with general error code and native error.
func ErrorStatus(err error) *Status {
	return NewStatus(ExitError, err)
}

// ErrorHTTPStatus creates new Status from HTTP response.
func ErrorHTTPStatus(res *http.Response) *Status {
	err := fmt.Errorf("Server returned HTTP status %s", res.Status)

	switch res.StatusCode / 100 {
	case 2:
		return nil
	case 4:
		switch res.StatusCode {
		case 404:
			return NewStatus(ExitNotFound, err)
		case 409:
			return NewStatus(ExitConflicted, err)
		default:
			return NewStatus(ExitResponse4xx, err)
		}
	case 5:
		return NewStatus(ExitResponse5xx, err)
	default:
		return ErrorStatus(err)
	}
}
