package errors

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// type TracedError error
type TracedError struct {
	Msg   string
	Stack string
	Err   error
}

func (e *TracedError) Error() string {
	return fmt.Sprintf("traceback: %s\n%s", e.Stack, e.Err.Error())
}

func (e *TracedError) Unwrap() error {
	return e.Err
}

// func TraceError(err error) TracedError {
// 	if err == nil {
// 		return nil
// 	}
// 	lines := strings.Split(string(debug.Stack()), "\n")
// 	if strings.Contains(err.Error(), "traceback:") {
// 		return err
// 	}
// 	stack := cleanStack(lines)
// 	return fmt.Errorf("traceback: %s\n%w", stack, err)
// }

// func TraceError(err error) error {
// 	if err == nil {
// 		return nil
// 	}
// 	if _, ok := err.(*TracedError); ok {
// 		return err
// 	}

// 	stack := cleanStack(strings.Split(string(debug.Stack()), "\n"))
// 	return &TracedError{
// 		Msg:   err.Error(),
// 		Stack: stack,
// 		Err:   err,
// 	}
// }

func TraceError(input interface{}) error {
	var baseErr error

	switch v := input.(type) {
	case nil:
		return nil
	case string:
		baseErr = fmt.Errorf(v)
	case error:
		// avoid double wrapping
		if _, ok := v.(*TracedError); ok {
			return v
		}
		baseErr = v
	default:
		baseErr = fmt.Errorf("unknown error type: %v", v)
	}

	stack := cleanStack(strings.Split(string(debug.Stack()), "\n"))
	return &TracedError{
		Msg:   baseErr.Error(),
		Stack: stack,
		Err:   baseErr,
	}
}

func cleanStack(lines []string) string {
	filtered := []string{}
	for _, line := range lines {
		if strings.Contains(line, "runtime/debug") || strings.Contains(line, "traceError") {
			continue
		}
		if strings.Contains(line, ".go:") {
			line = strings.Split(line, " ")[0]
			filtered = append(filtered, line)
		}
	}

	if len(filtered) > 1 {
		filtered = filtered[1:]
	}

	reversed := make([]string, len(filtered))
	for i, v := range filtered {
		reversed[len(filtered)-1-i] = v
	}

	return strings.Join(reversed, "\n")
}
