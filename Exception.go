package pgo

import (
    "fmt"
)

// NewException create new exception with status and message
func NewException(status int, msg ...interface{}) *Exception {
    message := ""
    if len(msg) == 1 {
        message = msg[0].(string)
    } else if len(msg) > 1 {
        message = fmt.Sprintf(msg[0].(string), msg[1:]...)
    }

    return &Exception{status, message}
}

// Exception panic as exception
type Exception struct {
    status  int
    message string
}

// GetStatus get exception status code
func (e *Exception) GetStatus() int {
    return e.status
}

// GetMessage get exception message string
func (e *Exception) GetMessage() string {
    return e.message
}

// Error implement error interface
func (e *Exception) Error() string {
    return fmt.Sprintf("exception: %d, message: %s", e.status, e.message)
}
