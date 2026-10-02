package log

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

var pid = strconv.Itoa(os.Getpid())

func GetPID() string {
	return pid
}

func GetGID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField, _, _ := strings.Cut(strings.TrimPrefix(string(buf[:n]), "goroutine "), " ")
	if _, err := strconv.ParseUint(idField, 10, 64); err != nil {
		return placeholder
	}
	return idField
}
