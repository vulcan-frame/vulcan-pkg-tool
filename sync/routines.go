package sync

import (
	"bytes"
	"runtime"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
)

// DefaultStackSize is the default size for stack traces
const DefaultStackSize = 64 << 10 // 64KB

const (
	initialRoutineIDBuffer = 128
)

// GoSafe executes a function in a separate goroutine with panic recovery.
// It logs any errors that occur during execution.
// msg: descriptive message for logging
// fn: function to execute safely
func GoSafe(msg string, fn func() error) {
	go func() {
		// 获取协程ID用于日志追踪
		rid := RoutineId()
		defer func() {
			if r := recover(); r != nil {
				log.Error("goroutine panic recovered",
					"message", msg,
					"routine_id", rid,
					"error", CatchErr(r),
				)
			}
		}()

		if err := RunSafe(fn); err != nil {
			log.Error("goroutine error occurred",
				"message", msg,
				"routine_id", rid,
				"error", err,
			)
		}
	}()
}

// RunSafe executes a function with panic recovery.
// Returns the error from the function or a wrapped error if a panic occurred.
func RunSafe(fn func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = CatchErr(p)
		}
	}()

	return fn()
}

// RoutineId returns the current goroutine ID.
// Warning: Only for debug purposes, never use it in production.
// The implementation is based on parsing the runtime stack.
func RoutineId() uint64 {
	buf := make([]byte, initialRoutineIDBuffer)
	n := runtime.Stack(buf, false)
	stack := buf[:n]

	const prefix = "goroutine "
	if !bytes.HasPrefix(stack, []byte(prefix)) {
		return 0
	}

	stack = stack[len(prefix):]
	end := bytes.IndexByte(stack, ' ')
	if end == -1 {
		return 0
	}

	var id uint64
	for _, c := range stack[:end] {
		if c < '0' || c > '9' {
			return 0
		}
		id = id*10 + uint64(c-'0')
	}
	return id
}

// CatchErr creates an error with stack trace from a recovered panic.
// It captures the current stack trace and formats it as part of the error message.
func CatchErr(p interface{}) error {
	return CatchErrWithSize(p, DefaultStackSize)
}

// CatchErrWithSize creates an error with a stack trace of the specified size from a recovered panic.
// stackSize: the maximum size of the stack trace to capture
func CatchErrWithSize(p interface{}, stackSize int) error {
	var buf []byte
	if stackSize <= DefaultStackSize {
		// reuse default stack size
		buf = make([]byte, DefaultStackSize)
	} else {
		buf = make([]byte, stackSize)
	}

	n := runtime.Stack(buf, false)
	buf = buf[:n]

	return errors.WithStack(
		errors.Errorf("panic recovered: %v\n%s", p, buf),
	)
}
