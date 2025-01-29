package utils

import (
	"fmt"
	"log"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// NewAppError creates a new instance of AppError with the provided code, message, and underlying error.
// Parameters:
//   - code: A string representing the error code.
//   - message: A string containing the error message.
//   - err: An error that caused this AppError.
//
// Returns:
//
//	A pointer to an AppError instance.
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error returns a formatted error string for the AppError.
// If the underlying error (e.Err) is not nil, it includes the error code, message, and the underlying error.
// Otherwise, it includes only the error code and message.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Log logs the error message of the AppError instance using the standard log package.
// The log message is prefixed with "ERROR: ".
func (e *AppError) Log() {
	log.Printf("ERROR: %s", e.Error())
}

// Unwrap returns the underlying error that caused the AppError.
// This method allows AppError to be used with errors.Unwrap and
// errors.Is for error unwrapping and comparison.
func (e *AppError) Unwrap() error {
	return e.Err
}
