package pgo

import (
    "fmt"
)

func NewException(status int, msg ...interface{}) *Exception {
    message := ""
    if len(msg) == 1 {
        message = msg[0].(string)
    } else if len(msg) > 1 {
        message = fmt.Sprintf(msg[0].(string), msg[1:]...)
    }

    return &Exception{status, message}
}

type Exception struct {
    status  int
    message string
}

func (e *Exception) GetStatus() int {
    return e.status
}

func (e *Exception) GetMessage() string {
    return e.message
}

// implement error interface
func (e *Exception) Error() string {
    return fmt.Sprintf("exception: %d, message: %s", e.status, e.message)
}
