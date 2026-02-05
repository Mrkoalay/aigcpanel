package errs

import (
	"fmt"
	"runtime"
	"strings"
)

func GetStacks() string {
	var stack []string
	for i := 1; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	joinStr := "\r\n"
	return strings.Join(stack, joinStr)
}
