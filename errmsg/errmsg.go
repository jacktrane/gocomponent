package errmsg

import "errors"

var (

	// gopool错误
	ErrConnPoolClosed = errors.New("ErrConnPoolClosed")
)
