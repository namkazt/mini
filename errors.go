package mini

import "fmt"

var (
	ErrUserInvalidInput   = -100
	ErrUserBadRequest     = -110
	ErrUserRecordNotFound = -200
	ErrUserDatabaseError  = -300
)

// Generate error response
func ErrorResponse(err int, a ...interface{}) *ResponseJSON {
	msg := "Error | no message"
	switch err {
	case ErrUserInvalidInput:
		msg = fmt.Sprintf("Input field %s is invalid.", a...)
	case ErrUserBadRequest:
		msg = fmt.Sprint("Bad request parameter:", a)
	case ErrUserRecordNotFound:
		msg = fmt.Sprint("Record not found:", a)
	case ErrUserDatabaseError:
		msg = fmt.Sprint("Database error:", a)
	}
	return &ResponseJSON{
		Code:    ErrUserInvalidInput,
		Message: msg,
	}
}

// some quick common error
func BadRequest(a ...interface{}) *ResponseJSON {
	return ErrorResponse(ErrUserBadRequest, a...)
}

func InvalidInput(a ...interface{}) *ResponseJSON {
	return ErrorResponse(ErrUserInvalidInput, a...)
}

func DatabaseError(a ...interface{}) *ResponseJSON {
	return ErrorResponse(ErrUserDatabaseError, a...)
}

func NotFound(a ...interface{}) *ResponseJSON {
	return ErrorResponse(ErrUserRecordNotFound, a...)
}
