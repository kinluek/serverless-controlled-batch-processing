// Package conf handles configuration and environment loading for the project applications.
package conf

// ErrCodeEnvNotSet happens when an environment variable name is not set when required.
const ErrCodeEnvNotSet = "ErrCodeEnvNotSet"

// Error is the package's custom error type.
type Error struct {
	Code    string
	Message string
}

func newError(code, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}

func (e *Error) Error() string {
	return e.Message
}
